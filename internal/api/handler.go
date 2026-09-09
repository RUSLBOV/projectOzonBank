package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"projectOzonBank/internal/domain"
)

type Handler struct {
	service Shortener
}

func NewHandler(service Shortener) *Handler {
	return &Handler{
		service: service,
	}
}

// Shorten обрабатывает POST-запрос на создание короткой ссылки.
// Принимает JSON {"url": "..."} и возвращает {"code": "..."} с кодом 201,
// либо существующий код, если URL уже был сокращён ранее.
func (h *Handler) Shorten(w http.ResponseWriter, r *http.Request) {
	var req ShortenRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	code, err := h.service.Shorten(r.Context(), req.URL)
	if err != nil {
		mapError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, ShortenResponse{
		Code: code,
	})
}

// Resolve обрабатывает GET-запрос по короткому коду и возвращает
// оригинальный URL в формате {"original_url": "..."}.
func (h *Handler) Resolve(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")

	originalURL, err := h.service.Resolve(r.Context(), code)
	if err != nil {
		mapError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, ResolveResponse{
		OriginalURL: originalURL,
	})
}

// mapError транслирует доменные ошибки в соответствующие HTTP-статусы
// и записывает JSON-ответ с сообщением об ошибке.
func mapError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		writeError(w, http.StatusNotFound, "short link not found")

	case errors.Is(err, domain.ErrInvalidURL):
		writeError(w, http.StatusBadRequest, "invalid url")

	case errors.Is(err, domain.ErrGenerationFailed):
		writeError(w, http.StatusServiceUnavailable, "try again later")

	default:
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}

// writeJSON сериализует v в JSON и записывает в ответ с указанным статусом.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(v)
}

// writeError - обёртка над writeJSON для унифицированного формата ошибок.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, ErrorResponse{
		Error: msg,
	})
}
