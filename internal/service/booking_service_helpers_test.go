package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"

	kafkaevent "hotelService/internal/kafka/event"
	"hotelService/internal/model"
)

func gatherCounterValue(t *testing.T, registry *prometheus.Registry, metricName string) float64 {
	t.Helper()

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("gather metrics: %v", err)
	}

	for _, family := range families {
		if family.GetName() != metricName || len(family.Metric) == 0 {
			continue
		}
		return family.Metric[0].GetCounter().GetValue()
	}

	t.Fatalf("metric %s not found", metricName)
	return 0
}

type stubBookingRepository struct {
	bookingByID    map[uuid.UUID]*model.Booking
	inserted       *model.Booking
	addedGuestIDs  []uuid.UUID
	available      bool
	availableRooms []model.Room
}

func (s *stubBookingRepository) FindByID(_ context.Context, id uuid.UUID) (*model.Booking, error) {
	if booking, ok := s.bookingByID[id]; ok {
		copyBooking := *booking
		return &copyBooking, nil
	}
	return nil, nil
}

func (s *stubBookingRepository) Insert(_ context.Context, entity model.Booking) error {
	s.inserted = &entity
	return nil
}

func (s *stubBookingRepository) Update(_ context.Context, _ model.Booking) error {
	return nil
}

func (s *stubBookingRepository) AddGuests(_ context.Context, _ uuid.UUID, guestIDs []uuid.UUID) error {
	s.addedGuestIDs = append([]uuid.UUID(nil), guestIDs...)
	return nil
}

func (s *stubBookingRepository) DeleteGuestsByBookingID(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (s *stubBookingRepository) FindGuestIDsByBookingID(_ context.Context, _ uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}

func (s *stubBookingRepository) Delete(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (s *stubBookingRepository) FindAvailableRooms(_ context.Context, _, _ time.Time) ([]model.Room, error) {
	return append([]model.Room(nil), s.availableRooms...), nil
}

func (s *stubBookingRepository) IsRoomAvailable(_ context.Context, _ uuid.UUID, _, _ time.Time) (bool, error) {
	return s.available, nil
}

func (s *stubBookingRepository) IsRoomAvailableForUpdate(_ context.Context, _, _ uuid.UUID, _, _ time.Time) (bool, error) {
	return s.available, nil
}

func (s *stubBookingRepository) CountActiveBookingsNow(_ context.Context) (int64, error) {
	return 0, nil
}

type stubBookingPublisher struct {
	event *kafkaevent.BookingEvent
}

func (s *stubBookingPublisher) PublishBookingEvent(_ context.Context, event kafkaevent.BookingEvent) error {
	s.event = &event
	return nil
}
