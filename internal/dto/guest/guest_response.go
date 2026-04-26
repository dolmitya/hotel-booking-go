package guest

import (
	"github.com/google/uuid"
)

type Response struct {
	ID          uuid.UUID `json:"id" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	LastName    string    `json:"lastName" example:"Ivanov"`
	FirstName   string    `json:"firstName" example:"Ivan"`
	MiddleName  *string   `json:"middleName,omitempty" example:"Ivanovich"`
	BirthDate   string    `json:"birthDate" example:"1990-05-12"`
	PhoneNumber string    `json:"phoneNumber" example:"+79991234567"`
}
