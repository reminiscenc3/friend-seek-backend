package models

import (
	"errors"
	"fmt"
)

const (
	MaxLoginLength = 32
	MaxNoteLength  = 4096
)

var (
	ErrInvalidLogin     = fmt.Errorf("login must be a non-empty string up to %d characters", MaxLoginLength)
	ErrNoteTooLong      = fmt.Errorf("note must be at most %d characters", MaxNoteLength)
	ErrUserNotFound     = errors.New("user not found")
	ErrLoginTaken       = errors.New("login already taken")
	ErrLocationNotFound = errors.New("location is not available yet")
	ErrNoCoordinates    = errors.New("lat and lon are required")
	ErrInvalidLocation  = errors.New("invalid coordinates")
	ErrInvalidAzimuth   = errors.New("azimuth must be a finite number of degrees")

	ErrBadJSON    = errors.New("malformed json body")
	ErrNoLogin    = errors.New("login is required")
	ErrNoDatabase = errors.New("database is unavailable")
)
