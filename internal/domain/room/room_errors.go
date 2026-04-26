package room

import "errors"

var (
	ErrNotFound        = errors.New("room not found")
	ErrInvalidRoomID   = errors.New("invalid room id")
	ErrInvalidFloor    = errors.New("invalid floor")
	ErrInvalidCapacity = errors.New("invalid capacity")
)
