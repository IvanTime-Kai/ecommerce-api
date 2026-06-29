package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Ivantime-Kai/ecommerce-api/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ShopHandler struct {
	service *service.ShopService
}

func NewShopHandler(service *service.ShopService) *ShopHandler {
	return &ShopHandler{
		service: service,
	}
}

type CreateShopRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	LogoURL     *string `json:"logo_url"`
}

func (h *ShopHandler) CreateShop(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromContext(r)

	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	var req CreateShopRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	shop, err := h.service.CreateShop(r.Context(), &service.CreateShopParams{
		OwnerID:     userID,
		Name:        req.Name,
		Description: req.Description,
		LogoURL:     req.LogoURL,
	})

	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"data": shop})
}

func (h *ShopHandler) GetMyShop(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromContext(r)

	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	shop, err := h.service.GetShopByOwnerID(r.Context(), userID)

	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": shop})
}

func (h *ShopHandler) GetShopByID(w http.ResponseWriter, r *http.Request) {
	paramId := chi.URLParam(r, "id")
	id, err := uuid.Parse(paramId)

	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid shop id")
		return
	}

	shop, err := h.service.GetShopByID(r.Context(), id)

	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": shop})
}
