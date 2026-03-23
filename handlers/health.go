package handlers

import "net/http"

func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
