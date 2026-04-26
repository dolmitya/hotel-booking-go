package model

import (
	"time"

	"github.com/google/uuid"
)

type Room struct {
	ID        uuid.UUID
	Floor     int
	Number    string
	Capacity  int
	CreatedAt time.Time
}
