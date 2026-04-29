package service

import (
	"context"

	"github.com/google/uuid"

	"hotelService/internal/model"
)

type guestRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.Guest, error)
	Insert(ctx context.Context, entity model.Guest) error
	Update(ctx context.Context, entity model.Guest) error
}
