package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	roomdomain "hotelService/internal/domain/room"
	roomdto "hotelService/internal/dto/room"
	"hotelService/internal/kafka"
	kafkaevent "hotelService/internal/kafka/event"
	"hotelService/internal/model"
	"hotelService/internal/repository"
)

type RoomService struct {
	repo      *repository.RoomRepository
	publisher kafka.RoomEventPublisher
}

func NewRoomService(repo *repository.RoomRepository, publisher kafka.RoomEventPublisher) *RoomService {
	return &RoomService{
		repo:      repo,
		publisher: publisher,
	}
}

func (s *RoomService) GetByID(ctx context.Context, rawID string) (roomdto.Response, error) {
	id, err := uuid.Parse(rawID)
	if err != nil {
		return roomdto.Response{}, roomdomain.ErrInvalidRoomID
	}

	room, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return roomdto.Response{}, err
	}
	if room == nil {
		return roomdto.Response{}, roomdomain.ErrNotFound
	}

	return toRoomResponse(*room), nil
}

func (s *RoomService) Create(ctx context.Context, request roomdto.Request) (roomdto.Response, error) {
	if err := validateRoomRequest(request); err != nil {
		return roomdto.Response{}, err
	}

	room := model.Room{
		ID:        uuid.New(),
		Floor:     request.Floor,
		Number:    request.Number,
		Capacity:  request.Capacity,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.repo.Insert(ctx, room); err != nil {
		return roomdto.Response{}, err
	}

	if err := s.publishRoomEvent(ctx, kafkaevent.RoomEvent{
		EventType: kafkaevent.EventTypeCreated,
		RoomID:    room.ID,
		Timestamp: time.Now().UTC(),
	}); err != nil {
		return roomdto.Response{}, err
	}

	return toRoomResponse(room), nil
}

func (s *RoomService) Update(ctx context.Context, rawID string, request roomdto.Request) (roomdto.Response, error) {
	id, err := uuid.Parse(rawID)
	if err != nil {
		return roomdto.Response{}, roomdomain.ErrInvalidRoomID
	}

	if err := validateRoomRequest(request); err != nil {
		return roomdto.Response{}, err
	}

	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return roomdto.Response{}, err
	}
	if existing == nil {
		return roomdto.Response{}, roomdomain.ErrNotFound
	}

	existing.Floor = request.Floor
	existing.Number = request.Number
	existing.Capacity = request.Capacity

	if err := s.repo.Update(ctx, *existing); err != nil {
		if errors.Is(err, repository.ErrRoomNotFound) {
			return roomdto.Response{}, roomdomain.ErrNotFound
		}

		return roomdto.Response{}, err
	}

	if err := s.publishRoomEvent(ctx, kafkaevent.RoomEvent{
		EventType: kafkaevent.EventTypeUpdated,
		RoomID:    existing.ID,
		Timestamp: time.Now().UTC(),
	}); err != nil {
		return roomdto.Response{}, err
	}

	return toRoomResponse(*existing), nil
}

func (s *RoomService) Delete(ctx context.Context, rawID string) error {
	id, err := uuid.Parse(rawID)
	if err != nil {
		return roomdomain.ErrInvalidRoomID
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrRoomNotFound) {
			return roomdomain.ErrNotFound
		}

		return err
	}

	if err := s.publishRoomEvent(ctx, kafkaevent.RoomEvent{
		EventType: kafkaevent.EventTypeDeleted,
		RoomID:    id,
		Timestamp: time.Now().UTC(),
	}); err != nil {
		return err
	}

	return nil
}

func validateRoomRequest(request roomdto.Request) error {
	if request.Floor < 0 {
		return roomdomain.ErrInvalidFloor
	}

	if request.Capacity < 1 {
		return roomdomain.ErrInvalidCapacity
	}

	return nil
}

func toRoomResponse(room model.Room) roomdto.Response {
	return roomdto.Response{
		ID:       room.ID,
		Floor:    room.Floor,
		Number:   room.Number,
		Capacity: room.Capacity,
	}
}

func (s *RoomService) publishRoomEvent(ctx context.Context, roomEvent kafkaevent.RoomEvent) error {
	if s.publisher == nil {
		return nil
	}

	if err := s.publisher.PublishRoomEvent(ctx, roomEvent); err != nil {
		return fmt.Errorf("publish room event: %w", err)
	}

	return nil
}
