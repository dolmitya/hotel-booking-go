package model

import (
	"time"

	"github.com/google/uuid"
)

type Booking struct {
	ID        uuid.UUID
	RoomID    uuid.UUID
	StartTime time.Time
	EndTime   time.Time
	CreatedAt time.Time
}
