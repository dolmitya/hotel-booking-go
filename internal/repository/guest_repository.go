package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"hotelService/internal/model"
)

var ErrGuestNotFound = errors.New("guest not found")

type GuestRepository struct {
	db *sql.DB
}

func NewGuestRepository(db *sql.DB) *GuestRepository {
	return &GuestRepository{db: db}
}

func (r *GuestRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Guest, error) {
	const query = `
		SELECT id, last_name, first_name, middle_name, birth_date, phone_number, created_at
		FROM guest
		WHERE id = $1
	`

	entity, err := scanGuest(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find guest by id %s: %w", id, err)
	}

	return &entity, nil
}

func (r *GuestRepository) FindAllByIDs(ctx context.Context, ids []uuid.UUID) ([]model.Guest, error) {
	if len(ids) == 0 {
		return []model.Guest{}, nil
	}

	placeholders := make([]string, 0, len(ids))
	args := make([]any, 0, len(ids))
	for i, id := range ids {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
		args = append(args, id)
	}

	query := fmt.Sprintf(`
		SELECT id, last_name, first_name, middle_name, birth_date, phone_number, created_at
		FROM guest
		WHERE id IN (%s)
	`, strings.Join(placeholders, ", "))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("find guests by ids: %w", err)
	}
	defer rows.Close()

	guests := make([]model.Guest, 0, len(ids))
	for rows.Next() {
		entity, scanErr := scanGuest(rows)
		if scanErr != nil {
			return nil, scanErr
		}

		guests = append(guests, entity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate guests by ids: %w", err)
	}

	return guests, nil
}

func (r *GuestRepository) Insert(ctx context.Context, entity model.Guest) error {
	const query = `
		INSERT INTO guest (
			id, last_name, first_name, middle_name,
			birth_date, phone_number, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		entity.ID,
		entity.LastName,
		entity.FirstName,
		entity.MiddleName,
		entity.BirthDate,
		entity.PhoneNumber,
		entity.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert guest %s: %w", entity.ID, err)
	}

	return nil
}

func (r *GuestRepository) Update(ctx context.Context, entity model.Guest) error {
	const query = `
		UPDATE guest
		SET last_name = $1,
			first_name = $2,
			middle_name = $3,
			birth_date = $4,
			phone_number = $5
		WHERE id = $6
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		entity.LastName,
		entity.FirstName,
		entity.MiddleName,
		entity.BirthDate,
		entity.PhoneNumber,
		entity.ID,
	)
	if err != nil {
		return fmt.Errorf("update guest %s: %w", entity.ID, err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read updated rows for guest %s: %w", entity.ID, err)
	}

	if affected == 0 {
		return ErrGuestNotFound
	}

	return nil
}

func scanGuest(scanner interface{ Scan(dest ...any) error }) (model.Guest, error) {
	var (
		guest      model.Guest
		middleName sql.NullString
	)

	if err := scanner.Scan(
		&guest.ID,
		&guest.LastName,
		&guest.FirstName,
		&middleName,
		&guest.BirthDate,
		&guest.PhoneNumber,
		&guest.CreatedAt,
	); err != nil {
		return model.Guest{}, err
	}

	if middleName.Valid {
		guest.MiddleName = &middleName.String
	}

	return guest, nil
}
