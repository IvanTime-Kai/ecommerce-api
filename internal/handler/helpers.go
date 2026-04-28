package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Ivantime-Kai/ecommerce-api/internal/middleware"
	"github.com/google/uuid"
)

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, code string, msg string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": msg,
		},
	})
}

func getUserIDFromContext(r *http.Request) (uuid.UUID, bool) {
	userIDStr, ok := r.Context().Value(middleware.UserIDKey).(string)

	if !ok {
		return uuid.Nil, false

	}

	userID, err := uuid.Parse(userIDStr)

	if err != nil {
		return uuid.Nil, false
	}

	return userID, true
}
