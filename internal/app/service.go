package app

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"projectOzonBank/internal/domain"
	"projectOzonBank/internal/shortener"
)

type LinkService struct {
	storage    domain.Storage
	maxRetries int
}

func New(storage domain.Storage, maxRetries int) *LinkService {
	return &LinkService{
		storage:    storage,
		maxRetries: maxRetries,
	}
}

func (s *LinkService) Shorten(ctx context.Context, originalURL string) (string, error) {
	parsed, err := url.ParseRequestURI(originalURL)
	if err != nil {
		return "", domain.ErrInvalidURL
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", domain.ErrInvalidURL
	}

	if parsed.Host == "" {
		return "", domain.ErrInvalidURL
	}

	for i := 0; i < s.maxRetries; i++ {
		code, err := shortener.Generate()
		if err != nil {
			return "", fmt.Errorf("generate code: %w", err)
		}

		err = s.storage.Save(ctx, code, originalURL)
		if err == nil {
			return code, nil
		}

		var existsErr *domain.AlreadyExistsError
		if errors.As(err, &existsErr) {
			return existsErr.ExistingCode, nil
		}

		if errors.Is(err, domain.ErrCodeAlreadyTaken) {
			continue
		}

		return "", err
	}

	return "", fmt.Errorf("%w: exhausted %d attempts", domain.ErrGenerationFailed, s.maxRetries)
}

func (s *LinkService) Resolve(ctx context.Context, code string) (string, error) {
	if len(code) != shortener.CodeLength {
		return "", domain.ErrNotFound
	}

	return s.storage.Get(ctx, code)
}
