package model

import "errors"

var (
	ErrNotFound      = errors.New("avatar not found")
	ErrForbidden     = errors.New("forbidden")
	ErrInvalidFormat = errors.New("invalid file format")
	ErrFileTooLarge  = errors.New("file too large")
	ErrMissingUserID = errors.New("missing user id")
	ErrUnavailable   = errors.New("dependency unavailable")
)
