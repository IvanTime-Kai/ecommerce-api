package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Ivantime-Kai/ecommerce-api/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ReviewHandler struct {
	service *service.ReviewService
}

func NewReviewHandler(service *service.ReviewService) *ReviewHandler {
	return &ReviewHandler{
		service: service,
	}
}

type CreateReviewRequest struct {
	OrderID   uuid.UUID `json:"order_id"`
	ProductID uuid.UUID `json:"product_id"`
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment"`
}

func (h ReviewHandler) CreateReview(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromContext(r)

	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	var req CreateReviewRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	review, err := h.service.CreateReview(r.Context(), service.CreateReviewParams{
		OrderID:   req.OrderID,
		ProductID: req.ProductID,
		UserID:    userID,
		Rating:    req.Rating,
		Comment:   req.Comment,
	})

	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"data": review})
}

func (h *ReviewHandler) GetReviewsByProductID(w http.ResponseWriter, r *http.Request) {
	productID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid product id")
		return
	}

	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	var cursor *uuid.UUID
	if cursorStr := r.URL.Query().Get("cursor"); cursorStr != "" {
		id, err := uuid.Parse(cursorStr)
		if err == nil {
			cursor = &id
		}
	}

	reviews, err := h.service.GetReviewsByProductID(r.Context(), service.GetReviewsParams{
		ProductID: productID,
		Limit:     int32(limit),
		Cursor:    cursor,
	})
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": reviews})
}
