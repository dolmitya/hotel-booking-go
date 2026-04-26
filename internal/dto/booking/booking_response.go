package booking

import (
	"time"

	"hotelService/internal/dto/guest"
	"hotelService/internal/dto/room"

	"github.com/google/uuid"
)

type Response struct {
	ID        uuid.UUID        `json:"id" example:"9a6c1f90-4d3b-4e7c-8e8a-1f23a1e7a123"`
	Room      room.Response    `json:"room"`
	Guests    []guest.Response `json:"guests"`
	StartTime time.Time        `json:"startTime" example:"2026-03-02T10:00:00Z"`
	EndTime   time.Time        `json:"endTime" example:"2026-03-04T11:00:00Z"`
}
