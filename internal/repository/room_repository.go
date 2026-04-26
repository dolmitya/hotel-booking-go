package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"hotelService/internal/model"
)

var ErrRoomNotFound = errors.New("room not found")

type RoomRepository struct {
	db *sql.DB
}

func NewRoomRepository(db *sql.DB) *RoomRepository {
	return &RoomRepository{db: db}
}

func (r *RoomRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Room, error) {
	const query = `
		SELECT id, floor, number, capacity, created_at
		FROM room
		WHERE id = $1
	`

	room, err := scanRoom(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("find room by id %s: %w", id, err)
	}

	return &room, nil
}

func (r *RoomRepository) Insert(ctx context.Context, room model.Room) error {
	const query = `
		INSERT INTO room (id, floor, number, capacity, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		room.ID,
		room.Floor,
		room.Number,
		room.Capacity,
		room.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert room %s: %w", room.ID, err)
	}

	return nil
}

func (r *RoomRepository) Update(ctx context.Context, room model.Room) error {
	const query = `
		UPDATE room
		SET floor = $1,
			number = $2,
			capacity = $3
		WHERE id = $4
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		room.Floor,
		room.Number,
		room.Capacity,
		room.ID,
	)
	if err != nil {
		return fmt.Errorf("update room %s: %w", room.ID, err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read updated rows for room %s: %w", room.ID, err)
	}

	if affected == 0 {
		return ErrRoomNotFound
	}

	return nil
}

func (r *RoomRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const query = `DELETE FROM room WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete room %s: %w", id, err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read deleted rows for room %s: %w", id, err)
	}

	if affected == 0 {
		return ErrRoomNotFound
	}

	return nil
}

func (r *RoomRepository) CountAll(ctx context.Context) (int64, error) {
	const query = `SELECT COUNT(*) FROM room`

	var count int64
	if err := r.db.QueryRowContext(ctx, query).Scan(&count); err != nil {
		return 0, fmt.Errorf("count rooms: %w", err)
	}

	return count, nil
}

func scanRoom(scanner interface{ Scan(dest ...any) error }) (model.Room, error) {
	var room model.Room

	if err := scanner.Scan(
		&room.ID,
		&room.Floor,
		&room.Number,
		&room.Capacity,
		&room.CreatedAt,
	); err != nil {
		return model.Room{}, err
	}

	return room, nil
}
