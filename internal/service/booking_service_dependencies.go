package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"hotelService/internal/model"
)

type bookingRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.Booking, error)
	Insert(ctx context.Context, entity model.Booking) error
	Update(ctx context.Context, entity model.Booking) error
	AddGuests(ctx context.Context, bookingID uuid.UUID, guestIDs []uuid.UUID) error
	DeleteGuestsByBookingID(ctx context.Context, bookingID uuid.UUID) error
	FindGuestIDsByBookingID(ctx context.Context, bookingID uuid.UUID) ([]uuid.UUID, error)
	Delete(ctx context.Context, id uuid.UUID) error
	FindAvailableRooms(ctx context.Context, start, end time.Time) ([]model.Room, error)
	IsRoomAvailable(ctx context.Context, roomID uuid.UUID, start, end time.Time) (bool, error)
	IsRoomAvailableForUpdate(ctx context.Context, bookingID, roomID uuid.UUID, start, end time.Time) (bool, error)
	CountActiveBookingsNow(ctx context.Context) (int64, error)
}

type bookingGuestRepository interface {
	FindAllByIDs(ctx context.Context, ids []uuid.UUID) ([]model.Guest, error)
}

type bookingRoomRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.Room, error)
}
