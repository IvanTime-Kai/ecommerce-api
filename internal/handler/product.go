package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Ivantime-Kai/ecommerce-api/internal/repository"
	"github.com/Ivantime-Kai/ecommerce-api/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ProductHandler struct {
	service *service.ProductService
}

func NewProductHandler(service *service.ProductService) *ProductHandler {
	return &ProductHandler{
		service: service,
	}
}

type CreateProductRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Price       float64 `json:"price"`
	Stock       int32   `json:"stock"`
}

type UpdateProductRequest struct {
	Name        string                   `json:"name"`
	Description *string                  `json:"description"`
	Status      repository.ProductStatus `json:"status"`
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromContext(r)

	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	var req CreateProductRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	product, err := h.service.CreateProduct(r.Context(), service.CreateProductParams{
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
	})

	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"data": product})
}

func (h *ProductHandler) GetProductByID(w http.ResponseWriter, r *http.Request) {
	paramId := chi.URLParam(r, "id")
	id, err := uuid.Parse(paramId)

	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid product id")
		return
	}

	product, err := h.service.GetProductByID(r.Context(), id)

	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": product})
}

func (h *ProductHandler) GetProductsByShopID(w http.ResponseWriter, r *http.Request) {
	paramId := chi.URLParam(r, "id")
	id, err := uuid.Parse(paramId)

	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid shop id")
		return
	}

	products, err := h.service.GetProductsByShopID(r.Context(), id)

	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": products})
}

func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	paramId := chi.URLParam(r, "id")
	id, err := uuid.Parse(paramId)

	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid product id")
		return
	}

	userID, ok := getUserIDFromContext(r)

	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	var req UpdateProductRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	product, err := h.service.UpdateProduct(r.Context(), service.UpdateProductParams{
		ID:          id,
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Status:      req.Status,
	})

	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": product})
}

func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	paramId := chi.URLParam(r, "id")
	id, err := uuid.Parse(paramId)

	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid product id")
		return
	}

	userID, ok := getUserIDFromContext(r)

	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	err = h.service.DeleteProduct(r.Context(), service.DeleteProductParams{
		ID:     id,
		UserID: userID,
	})

	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": "product deleted successfully"})
}

func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	query := q.Get("q")

	var categoryID uuid.UUID
	if s := q.Get("category_id"); s != "" {
		categoryID, _ = uuid.Parse(s)
	}

	var minPrice, maxPrice float64
	if s := q.Get("min_price"); s != "" {
		minPrice, _ = strconv.ParseFloat(s, 64)
	}
	if s := q.Get("max_price"); s != "" {
		maxPrice, _ = strconv.ParseFloat(s, 64)
	}

	var cursor uuid.UUID
	if s := q.Get("cursor"); s != "" {
		cursor, _ = uuid.Parse(s)
	}

	limit := int32(20)
	if s := q.Get("limit"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v > 0 && v <= 100 {
			limit = int32(v)
		}
	}

	result, err := h.service.SearchProducts(r.Context(), service.SearchProductsParams{
		Query:      query,
		CategoryID: categoryID,
		MinPrice:   minPrice,
		MaxPrice:   maxPrice,
		Cursor:     cursor,
		Limit:      limit,
	})

	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": result})
}
