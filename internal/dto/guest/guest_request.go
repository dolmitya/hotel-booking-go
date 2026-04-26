package guest

type Request struct {
	LastName    string  `json:"lastName" binding:"required" example:"Ivanov"`
	FirstName   string  `json:"firstName" binding:"required" example:"Ivan"`
	MiddleName  *string `json:"middleName,omitempty" example:"Ivanovich"`
	BirthDate   string  `json:"birthDate" binding:"required" example:"1990-05-12"`
	PhoneNumber string  `json:"phoneNumber" binding:"required" example:"+79991234567"`
}
