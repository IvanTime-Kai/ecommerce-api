package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Ivantime-Kai/ecommerce-api/internal/middleware"
	"github.com/Ivantime-Kai/ecommerce-api/internal/service"
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

func handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrOutOfStock):
		writeError(w, http.StatusConflict, "OUT_OF_STOCK", err.Error())
	case errors.Is(err, service.ErrEmptyItems):
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
	case errors.Is(err, service.ErrInvalidProducts):
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
	case errors.Is(err, service.ErrInvalidTransition):
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
	case errors.Is(err, service.ErrNotFound):
		writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
	case errors.Is(err, service.ErrForbidden):
		writeError(w, http.StatusForbidden, "FORBIDDEN", err.Error())
	case errors.Is(err, service.ErrAlreadyReviewed):
		writeError(w, http.StatusConflict, "ALREADY_REVIEWED", err.Error())
	case errors.Is(err, service.ErrOrderNotDelivered):
		writeError(w, http.StatusBadRequest, "ORDER_NOT_DELIVERED", err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
}
