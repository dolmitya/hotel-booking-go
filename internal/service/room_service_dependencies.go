package service

import (
	"context"

	"github.com/google/uuid"

	"hotelService/internal/model"
)

type roomRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.Room, error)
	Insert(ctx context.Context, room model.Room) error
	Update(ctx context.Context, room model.Room) error
	Delete(ctx context.Context, id uuid.UUID) error
}
