package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	guestdomain "hotelService/internal/domain/guest"
	guestdto "hotelService/internal/dto/guest"
	kafkaevent "hotelService/internal/kafka/event"
)

func TestGuestService_CreateRejectsInvalidBirthDate(t *testing.T) {
	t.Parallel()

	svc := NewGuestService(&stubGuestRepository{}, nil)

	_, err := svc.Create(context.Background(), guestdto.Request{
		LastName:    "Ivanov",
		FirstName:   "Ivan",
		BirthDate:   "12-05-1990",
		PhoneNumber: "+79991234567",
	})

	if !errors.Is(err, guestdomain.ErrInvalidBirthDate) {
		t.Fatalf("expected ErrInvalidBirthDate, got %v", err)
	}
}

func TestGuestService_CreatePublishesEvent(t *testing.T) {
	t.Parallel()

	repo := &stubGuestRepository{}
	publisher := &stubGuestPublisher{}
	svc := NewGuestService(repo, publisher)

	response, err := svc.Create(context.Background(), guestdto.Request{
		LastName:    "Ivanov",
		FirstName:   "Ivan",
		BirthDate:   "1990-05-12",
		PhoneNumber: "+79991234567",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if response.ID == uuid.Nil {
		t.Fatal("expected generated id")
	}
	if repo.inserted == nil {
		t.Fatal("expected inserted guest")
	}
	if publisher.event == nil || publisher.event.EventType != kafkaevent.EventTypeCreated {
		t.Fatalf("unexpected event: %+v", publisher.event)
	}
}

func TestGuestService_GetByIDRejectsInvalidID(t *testing.T) {
	t.Parallel()

	svc := NewGuestService(&stubGuestRepository{}, nil)

	_, err := svc.GetByID(context.Background(), "bad-id")
	if !errors.Is(err, guestdomain.ErrInvalidGuestID) {
		t.Fatalf("expected ErrInvalidGuestID, got %v", err)
	}
}
