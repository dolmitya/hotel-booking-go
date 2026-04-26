package event

import (
	"time"

	"github.com/google/uuid"
)

type GuestEvent struct {
	EventType EventType `json:"eventType"`
	GuestID   uuid.UUID `json:"guestId"`
	Timestamp time.Time `json:"timestamp"`
}
