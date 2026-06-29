package handler

import (
	"net/http"

	"github.com/Ivantime-Kai/ecommerce-api/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type CategoryHandler struct {
	service *service.CategoryService
}

func NewCategoryHandler(service *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		service: service,
	}
}

func (h *CategoryHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.service.GetCategories(r.Context())

	if err != nil {
		handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": categories})
}

func (h *CategoryHandler) GetSubCategories(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))

	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid category id")
		return
	}

	categories, err := h.service.GetSubCategories(r.Context(), id)

	if err != nil {
		handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": categories})
}
