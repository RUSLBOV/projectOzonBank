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
func (h *Handler) Shorten(w http.ResponseWriter, r *http.Request) {
	var req ShortenRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "неправильное тело запроса")
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

func mapError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		writeError(w, http.StatusNotFound, "короткая ссылка не найдена")

	case errors.Is(err, domain.ErrInvalidURL):
		writeError(w, http.StatusBadRequest, "неккоректный url")

	case errors.Is(err, domain.ErrGenerationFailed):
		writeError(w, http.StatusServiceUnavailable, "попробуйте позже")

	default:
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, ErrorResponse{
		Error: msg,
	})
}
