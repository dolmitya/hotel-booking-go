package guest

import "errors"

var (
	ErrNotFound         = errors.New("guest not found")
	ErrInvalidGuestID   = errors.New("invalid guest id")
	ErrInvalidBirthDate = errors.New("invalid birth date")
)
