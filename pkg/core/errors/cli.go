package errors

import "errors"

// ErrInvalidPassword is returned when a password does not meet the requirements
var ErrInvalidPassword = errors.New("invalid password")
