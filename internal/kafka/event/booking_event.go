package event

import (
	"time"

	"github.com/google/uuid"
)

type BookingEvent struct {
	EventType EventType   `json:"eventType"`
	BookingID uuid.UUID   `json:"bookingId"`
	GuestIDs  []uuid.UUID `json:"guestIds"`
	RoomID    uuid.UUID   `json:"roomId"`
	Timestamp time.Time   `json:"timestamp"`
}
