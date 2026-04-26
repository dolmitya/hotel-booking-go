package event

import (
	"time"

	"github.com/google/uuid"
)

type RoomEvent struct {
	EventType EventType `json:"eventType"`
	RoomID    uuid.UUID `json:"roomId"`
	Timestamp time.Time `json:"timestamp"`
}
