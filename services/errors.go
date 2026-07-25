package services

import "errors"

// Sentinel errors the read/authz layer returns so controllers can map them to
// HTTP status codes (403 / 404) instead of leaking 500s.
var (
	ErrForbidden = errors.New("forbidden")
	ErrNotFound  = errors.New("not found")
)
