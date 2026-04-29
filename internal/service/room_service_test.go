package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	roomdomain "hotelService/internal/domain/room"
	roomdto "hotelService/internal/dto/room"
	kafkaevent "hotelService/internal/kafka/event"
	"hotelService/internal/repository"
)

func TestRoomService_CreateRejectsInvalidFloor(t *testing.T) {
	t.Parallel()

	svc := NewRoomService(&stubRoomRepository{}, nil)

	_, err := svc.Create(context.Background(), roomdto.Request{
		Floor:    -1,
		Number:   "101",
		Capacity: 2,
	})

	if !errors.Is(err, roomdomain.ErrInvalidFloor) {
		t.Fatalf("expected ErrInvalidFloor, got %v", err)
	}
}

func TestRoomService_CreatePublishesEvent(t *testing.T) {
	t.Parallel()

	repo := &stubRoomRepository{}
	publisher := &stubRoomPublisher{}
	svc := NewRoomService(repo, publisher)

	response, err := svc.Create(context.Background(), roomdto.Request{
		Floor:    3,
		Number:   "305",
		Capacity: 2,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if response.ID == uuid.Nil {
		t.Fatal("expected generated room id")
	}
	if repo.inserted == nil {
		t.Fatal("expected inserted room")
	}
	if publisher.event == nil || publisher.event.EventType != kafkaevent.EventTypeCreated {
		t.Fatalf("unexpected event: %+v", publisher.event)
	}
}

func TestRoomService_DeleteMapsRepositoryNotFound(t *testing.T) {
	t.Parallel()

	svc := NewRoomService(&stubRoomRepository{deleteErr: repository.ErrRoomNotFound}, nil)

	err := svc.Delete(context.Background(), uuid.NewString())
	if !errors.Is(err, roomdomain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
