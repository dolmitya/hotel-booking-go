package model

import (
	"time"

	"github.com/google/uuid"
)

type Guest struct {
	ID          uuid.UUID
	LastName    string
	FirstName   string
	MiddleName  *string
	BirthDate   time.Time
	PhoneNumber string
	CreatedAt   time.Time
}
