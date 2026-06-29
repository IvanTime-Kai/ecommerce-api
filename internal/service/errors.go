package service

import "fmt"

var (
	ErrNotFound     = fmt.Errorf("not found")
	ErrForbidden    = fmt.Errorf("forbidden")
	ErrUnauthorized = fmt.Errorf("unauthorized")
)
