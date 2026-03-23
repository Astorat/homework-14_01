package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"blog-api/auth"
	"blog-api/logger"
	"blog-api/storage"
)

type Handler struct {
	store  *storage.FileStorage
	auth   *auth.TokenAuth
	logger *logger.EventLogger
}

func NewHandler(store *storage.FileStorage, auth *auth.TokenAuth, logger *logger.EventLogger) *Handler {
	return &Handler{store: store, auth: auth, logger: logger}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, ErrorResponse{Error: message})
}

func (h *Handler) getUserIDFromRequest(r *http.Request) (int, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return 0, errors.New("authorization header required")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return 0, errors.New("invalid authorization header format, use: Bearer <token>")
	}

	return h.auth.ValidateToken(parts[1])
}
