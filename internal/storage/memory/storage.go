package memory

import (
	"context"
	"sync"

	"projectOzonBank/internal/domain"
)

type Storage struct {
	mu        sync.RWMutex
	codeToURL map[string]string
	urlToCode map[string]string
}

func New() *Storage {
	return &Storage{
		codeToURL: make(map[string]string),
		urlToCode: make(map[string]string),
	}
}

func (s *Storage) Save(ctx context.Context, code, originalURL string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existingCode, ok := s.urlToCode[originalURL]; ok {
		return &domain.AlreadyExistsError{ExistingCode: existingCode}
	}

	if _, ok := s.codeToURL[code]; ok {
		return domain.ErrCodeAlreadyTaken
	}

	s.urlToCode[originalURL] = code
	s.codeToURL[code] = originalURL
	return nil
}

func (s *Storage) Get(ctx context.Context, code string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, ok := s.codeToURL[code]
	if ok {
		return data, nil
	}
	return "", domain.ErrNotFound
}
