package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"

	bookingdomain "hotelService/internal/domain/booking"
	bookingdto "hotelService/internal/dto/booking"
	kafkaevent "hotelService/internal/kafka/event"
	"hotelService/internal/metrics"
	"hotelService/internal/model"
)

func TestBookingService_CreateRejectsCapacityExceeded(t *testing.T) {
	t.Parallel()

	roomID := uuid.New()
	guest1 := uuid.New()
	guest2 := uuid.New()

	svc := NewBookingService(
		&stubBookingRepository{},
		&stubGuestRepository{
			guestByID: map[uuid.UUID]*model.Guest{
				guest1: {ID: guest1},
				guest2: {ID: guest2},
			},
		},
		&stubRoomRepository{
			roomByID: map[uuid.UUID]*model.Room{
				roomID: {ID: roomID, Number: "101", Capacity: 1},
			},
		},
		nil,
		nil,
	)

	_, err := svc.Create(context.Background(), bookingdto.Request{
		GuestIDs:  []string{guest1.String(), guest2.String()},
		RoomID:    roomID.String(),
		StartTime: "2026-04-26T10:00:00Z",
		EndTime:   "2026-04-26T12:00:00Z",
	})

	if !errors.Is(err, bookingdomain.ErrRoomCapacityExceeded) {
		t.Fatalf("expected ErrRoomCapacityExceeded, got %v", err)
	}
}

func TestBookingService_CreateRejectsUnavailableRoom(t *testing.T) {
	t.Parallel()

	roomID := uuid.New()
	guestID := uuid.New()

	svc := NewBookingService(
		&stubBookingRepository{available: false},
		&stubGuestRepository{
			guestByID: map[uuid.UUID]*model.Guest{
				guestID: {ID: guestID},
			},
		},
		&stubRoomRepository{
			roomByID: map[uuid.UUID]*model.Room{
				roomID: {ID: roomID, Number: "101", Capacity: 2},
			},
		},
		nil,
		nil,
	)

	_, err := svc.Create(context.Background(), bookingdto.Request{
		GuestIDs:  []string{guestID.String()},
		RoomID:    roomID.String(),
		StartTime: "2026-04-26T10:00:00Z",
		EndTime:   "2026-04-26T12:00:00Z",
	})

	if !errors.Is(err, bookingdomain.ErrRoomNotAvailable) {
		t.Fatalf("expected ErrRoomNotAvailable, got %v", err)
	}
}

func TestBookingService_CreatePublishesEventAndIncrementsMetrics(t *testing.T) {
	t.Parallel()

	roomID := uuid.New()
	guestID := uuid.New()
	registry := prometheus.NewRegistry()
	bookingMetrics := metrics.NewBookingMetrics(registry)
	bookingRepo := &stubBookingRepository{available: true}
	publisher := &stubBookingPublisher{}

	svc := NewBookingService(
		bookingRepo,
		&stubGuestRepository{
			guestByID: map[uuid.UUID]*model.Guest{
				guestID: {ID: guestID, FirstName: "Ivan"},
			},
		},
		&stubRoomRepository{
			roomByID: map[uuid.UUID]*model.Room{
				roomID: {ID: roomID, Number: "101", Capacity: 2},
			},
		},
		publisher,
		bookingMetrics,
	)

	response, err := svc.Create(context.Background(), bookingdto.Request{
		GuestIDs:  []string{guestID.String()},
		RoomID:    roomID.String(),
		StartTime: "2026-04-26T10:00:00Z",
		EndTime:   "2026-04-26T12:00:00Z",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if response.ID == uuid.Nil {
		t.Fatal("expected created booking id")
	}
	if bookingRepo.inserted == nil {
		t.Fatal("expected booking insert")
	}
	if len(bookingRepo.addedGuestIDs) != 1 || bookingRepo.addedGuestIDs[0] != guestID {
		t.Fatalf("unexpected linked guests: %+v", bookingRepo.addedGuestIDs)
	}
	if publisher.event == nil || publisher.event.EventType != kafkaevent.EventTypeCreated {
		t.Fatalf("unexpected event: %+v", publisher.event)
	}
	if got := gatherCounterValue(t, registry, "booking_created_total"); got != 1 {
		t.Fatalf("expected booking_created_total=1, got %v", got)
	}
}

func TestBookingService_FindAvailableRoomsRejectsInvalidRange(t *testing.T) {
	t.Parallel()

	svc := NewBookingService(&stubBookingRepository{}, &stubGuestRepository{}, &stubRoomRepository{}, nil, nil)

	_, err := svc.FindAvailableRooms(context.Background(), "2026-04-26T12:00:00Z", "2026-04-26T10:00:00Z")
	if !errors.Is(err, bookingdomain.ErrInvalidTimeRange) {
		t.Fatalf("expected ErrInvalidTimeRange, got %v", err)
	}
}
