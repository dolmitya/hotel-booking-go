package booking

import "errors"

var (
	ErrNotFound             = errors.New("booking not found")
	ErrInvalidBookingID     = errors.New("invalid booking id")
	ErrInvalidRoomID        = errors.New("invalid room id")
	ErrInvalidTimeRange     = errors.New("end time must be after start time")
	ErrRoomNotFound         = errors.New("room not found")
	ErrGuestNotFound        = errors.New("one or more guests not found")
	ErrRoomCapacityExceeded = errors.New("room capacity exceeded")
	ErrRoomNotAvailable     = errors.New("room is not available for selected time range")
)
