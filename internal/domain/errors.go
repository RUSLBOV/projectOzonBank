package domain

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound         = errors.New("short link not found")
	ErrInvalidURL       = errors.New("invalid original url")
	ErrCodeAlreadyTaken = errors.New("short code already taken")
	ErrGenerationFailed = errors.New("failed to generate unique code")
)

type AlreadyExistsError struct {
	ExistingCode string
}

func (e *AlreadyExistsError) Error() string {
	return fmt.Sprintf("URL already taken, it's shorten version: %s", e.ExistingCode)
}
