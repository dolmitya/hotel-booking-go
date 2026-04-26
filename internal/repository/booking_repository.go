package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"hotelService/internal/model"
)

var ErrBookingNotFound = errors.New("booking not found")

type BookingRepository struct {
	db *sql.DB
}

func NewBookingRepository(db *sql.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

func (r *BookingRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Booking, error) {
	const query = `
		SELECT id, room_id, start_time, end_time, created_at
		FROM booking
		WHERE id = $1
	`

	entity, err := scanBooking(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find booking by id %s: %w", id, err)
	}

	return &entity, nil
}

func (r *BookingRepository) Insert(ctx context.Context, entity model.Booking) error {
	const query = `
		INSERT INTO booking (id, room_id, start_time, end_time, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.ExecContext(ctx, query, entity.ID, entity.RoomID, entity.StartTime, entity.EndTime, entity.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert booking %s: %w", entity.ID, err)
	}

	return nil
}

func (r *BookingRepository) Update(ctx context.Context, entity model.Booking) error {
	const query = `
		UPDATE booking
		SET room_id = $1, start_time = $2, end_time = $3
		WHERE id = $4
	`

	result, err := r.db.ExecContext(ctx, query, entity.RoomID, entity.StartTime, entity.EndTime, entity.ID)
	if err != nil {
		return fmt.Errorf("update booking %s: %w", entity.ID, err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read updated rows for booking %s: %w", entity.ID, err)
	}

	if affected == 0 {
		return ErrBookingNotFound
	}

	return nil
}

func (r *BookingRepository) AddGuests(ctx context.Context, bookingID uuid.UUID, guestIDs []uuid.UUID) error {
	const query = `
		INSERT INTO booking_guest (booking_id, guest_id)
		VALUES ($1, $2)
	`

	for _, guestID := range guestIDs {
		if _, err := r.db.ExecContext(ctx, query, bookingID, guestID); err != nil {
			return fmt.Errorf("insert booking_guest booking=%s guest=%s: %w", bookingID, guestID, err)
		}
	}

	return nil
}

func (r *BookingRepository) DeleteGuestsByBookingID(ctx context.Context, bookingID uuid.UUID) error {
	const query = `DELETE FROM booking_guest WHERE booking_id = $1`
	if _, err := r.db.ExecContext(ctx, query, bookingID); err != nil {
		return fmt.Errorf("delete booking guests for booking %s: %w", bookingID, err)
	}

	return nil
}

func (r *BookingRepository) FindGuestIDsByBookingID(ctx context.Context, bookingID uuid.UUID) ([]uuid.UUID, error) {
	const query = `
		SELECT guest_id
		FROM booking_guest
		WHERE booking_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, bookingID)
	if err != nil {
		return nil, fmt.Errorf("find guest ids by booking id %s: %w", bookingID, err)
	}
	defer rows.Close()

	ids := make([]uuid.UUID, 0)
	for rows.Next() {
		var id uuid.UUID
		if scanErr := rows.Scan(&id); scanErr != nil {
			return nil, fmt.Errorf("scan guest id by booking id %s: %w", bookingID, scanErr)
		}
		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate guest ids by booking id %s: %w", bookingID, err)
	}

	return ids, nil
}

func (r *BookingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const query = `DELETE FROM booking WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete booking %s: %w", id, err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read deleted rows for booking %s: %w", id, err)
	}

	if affected == 0 {
		return ErrBookingNotFound
	}

	return nil
}

func (r *BookingRepository) FindAvailableRooms(ctx context.Context, start, end time.Time) ([]model.Room, error) {
	const query = `
		SELECT r.id, r.floor, r.number, r.capacity, r.created_at
		FROM room r
		WHERE r.id NOT IN (
			SELECT room_id FROM booking
			WHERE ($1 < end_time AND $2 > start_time)
		)
	`

	rows, err := r.db.QueryContext(ctx, query, start, end)
	if err != nil {
		return nil, fmt.Errorf("find available rooms: %w", err)
	}
	defer rows.Close()

	rooms := make([]model.Room, 0)
	for rows.Next() {
		room, scanErr := scanRoom(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan available room: %w", scanErr)
		}
		rooms = append(rooms, room)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate available rooms: %w", err)
	}

	return rooms, nil
}

func (r *BookingRepository) IsRoomAvailable(ctx context.Context, roomID uuid.UUID, start, end time.Time) (bool, error) {
	const query = `
		SELECT NOT EXISTS (
			SELECT 1
			FROM booking b
			WHERE b.room_id = $1
			  AND ($2 < b.end_time AND $3 > b.start_time)
		)
	`

	var available bool
	if err := r.db.QueryRowContext(ctx, query, roomID, start, end).Scan(&available); err != nil {
		return false, fmt.Errorf("check room availability room=%s: %w", roomID, err)
	}

	return available, nil
}

func (r *BookingRepository) IsRoomAvailableForUpdate(
	ctx context.Context,
	bookingID, roomID uuid.UUID,
	start, end time.Time,
) (bool, error) {
	const query = `
		SELECT NOT EXISTS (
			SELECT 1
			FROM booking b
			WHERE b.room_id = $1
			  AND b.id <> $2
			  AND ($3 < b.end_time AND $4 > b.start_time)
		)
	`

	var available bool
	if err := r.db.QueryRowContext(ctx, query, roomID, bookingID, start, end).Scan(&available); err != nil {
		return false, fmt.Errorf("check room availability for update booking=%s room=%s: %w", bookingID, roomID, err)
	}

	return available, nil
}

func (r *BookingRepository) CountActiveBookingsNow(ctx context.Context) (int64, error) {
	const query = `
		SELECT COUNT(*)
		FROM booking
		WHERE NOW() >= start_time
		  AND NOW() < end_time
	`

	var count int64
	if err := r.db.QueryRowContext(ctx, query).Scan(&count); err != nil {
		return 0, fmt.Errorf("count active bookings now: %w", err)
	}

	return count, nil
}

func scanBooking(scanner interface{ Scan(dest ...any) error }) (model.Booking, error) {
	var booking model.Booking

	if err := scanner.Scan(
		&booking.ID,
		&booking.RoomID,
		&booking.StartTime,
		&booking.EndTime,
		&booking.CreatedAt,
	); err != nil {
		return model.Booking{}, err
	}

	return booking, nil
}
