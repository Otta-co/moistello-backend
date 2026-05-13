package user

import "errors"

var (
	ErrUserNotFound = errors.New("user not found")
	ErrInvalidUUID  = errors.New("invalid user ID format")
)
