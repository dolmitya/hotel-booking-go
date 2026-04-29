package service

import (
	"context"

	"github.com/google/uuid"

	kafkaevent "hotelService/internal/kafka/event"
	"hotelService/internal/model"
)

type stubGuestRepository struct {
	guestByID map[uuid.UUID]*model.Guest
	inserted  *model.Guest
	updateErr error
}

func (s *stubGuestRepository) FindByID(_ context.Context, id uuid.UUID) (*model.Guest, error) {
	if guest, ok := s.guestByID[id]; ok {
		copyGuest := *guest
		return &copyGuest, nil
	}
	return nil, nil
}

func (s *stubGuestRepository) FindAllByIDs(_ context.Context, ids []uuid.UUID) ([]model.Guest, error) {
	result := make([]model.Guest, 0, len(ids))
	for _, id := range ids {
		if guest, ok := s.guestByID[id]; ok {
			result = append(result, *guest)
		}
	}
	return result, nil
}

func (s *stubGuestRepository) Insert(_ context.Context, entity model.Guest) error {
	s.inserted = &entity
	return nil
}

func (s *stubGuestRepository) Update(_ context.Context, _ model.Guest) error {
	return s.updateErr
}

type stubGuestPublisher struct {
	event *kafkaevent.GuestEvent
}

func (s *stubGuestPublisher) PublishGuestEvent(_ context.Context, event kafkaevent.GuestEvent) error {
	s.event = &event
	return nil
}
