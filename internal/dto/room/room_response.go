package room

import "github.com/google/uuid"

type Response struct {
	ID       uuid.UUID `json:"id" example:"9a6c1f90-4d3b-4e7c-8e8a-1f23a1e7a123"`
	Floor    int       `json:"floor" example:"3"`
	Number   string    `json:"number" example:"305"`
	Capacity int       `json:"capacity" example:"2"`
}
