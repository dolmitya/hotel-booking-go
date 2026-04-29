package service

import (
	"context"

	"github.com/google/uuid"

	kafkaevent "hotelService/internal/kafka/event"
	"hotelService/internal/model"
)

type stubRoomRepository struct {
	roomByID  map[uuid.UUID]*model.Room
	inserted  *model.Room
	deleteErr error
}

func (s *stubRoomRepository) FindByID(_ context.Context, id uuid.UUID) (*model.Room, error) {
	if room, ok := s.roomByID[id]; ok {
		copyRoom := *room
		return &copyRoom, nil
	}
	return nil, nil
}

func (s *stubRoomRepository) Insert(_ context.Context, room model.Room) error {
	s.inserted = &room
	return nil
}

func (s *stubRoomRepository) Update(_ context.Context, _ model.Room) error {
	return nil
}

func (s *stubRoomRepository) Delete(_ context.Context, _ uuid.UUID) error {
	return s.deleteErr
}

func (s *stubRoomRepository) CountAll(_ context.Context) (int64, error) {
	return 0, nil
}

type stubRoomPublisher struct {
	event *kafkaevent.RoomEvent
}

func (s *stubRoomPublisher) PublishRoomEvent(_ context.Context, event kafkaevent.RoomEvent) error {
	s.event = &event
	return nil
}
