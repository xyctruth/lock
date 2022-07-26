package lock

import "errors"

var (
	ErrAlreadyLocked    = errors.New("ErrAlreadyLocked")
	ErrDeadlineExceeded = errors.New("ErrDeadlineExceeded")
	ErrNotFoundLocked   = errors.New("ErrNotFoundLocked")
)
