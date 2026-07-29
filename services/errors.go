package services

import "errors"

var (
	ErrForbidden    = errors.New("forbidden")
	ErrNotFound     = errors.New("not found")
	ErrInvalidInput = errors.New("invalid input")
	ErrUpstream     = errors.New("upstream unavailable")
)
