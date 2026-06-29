package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Ivantime-Kai/ecommerce-api/internal/repository"
	"github.com/Ivantime-Kai/ecommerce-api/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type OrderHandler struct {
	service *service.OrderService
}

func NewOrderHandler(service *service.OrderService) *OrderHandler {
	return &OrderHandler{
		service: service,
	}
}

type OrderItemInput struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
}

type CreateOrderRequest struct {
	ShopID           uuid.UUID        `json:"shop_id"`
	ShippingFullName string           `json:"shipping_full_name"`
	ShippingPhone    string           `json:"shipping_phone"`
	ShippingProvince string           `json:"shipping_province"`
	ShippingDistrict string           `json:"shipping_district"`
	ShippingWard     string           `json:"shipping_ward"`
	ShippingStreet   string           `json:"shipping_street"`
	Items            []OrderItemInput `json:"items"`
}

type GetOrdersByUserIDRequest struct {
	OffSet     int        `json:"off_set"`
	Limit      int        `json:"limit"`
	NextCursor *uuid.UUID `json:"next_cursor"`
}

type GetRevenueSummaryRequest struct {
	FromDate *time.Time
	ToDate   *time.Time
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromContext(r)

	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	idempotencyKey := r.Header.Get("X-Idempotency-Key")

	var req CreateOrderRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	items := make([]service.OrderItemInput, len(req.Items))
	for i, item := range req.Items {
		items[i] = service.OrderItemInput{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	order, err := h.service.CreateOrder(r.Context(), service.CreateOrderParams{
		UserID:           userID,
		ShopID:           req.ShopID,
		ShippingFullName: req.ShippingFullName,
		ShippingPhone:    req.ShippingPhone,
		ShippingProvince: req.ShippingProvince,
		ShippingDistrict: req.ShippingDistrict,
		ShippingWard:     req.ShippingWard,
		ShippingStreet:   req.ShippingStreet,
		Items:            items,
		IdempotencyKey:   idempotencyKey,
	})

	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"data": order})
}

func (h *OrderHandler) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	paramId := chi.URLParam(r, "id")
	id, err := uuid.Parse(paramId)

	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid order id")
		return
	}

	order, err := h.service.GetOrderByID(r.Context(), id)

	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": order})
}

func (h *OrderHandler) GetOrdersByUserID(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromContext(r)

	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit, _ := strconv.Atoi(limitStr)

	var cursor *uuid.UUID
	if cursorStr := r.URL.Query().Get("cursor"); cursorStr != "" {
		id, err := uuid.Parse(cursorStr)

		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_CURSOR", "invalid cursor")
			return
		}
		cursor = &id
	}

	orders, err := h.service.GetOrdersByUserID(r.Context(), service.GetOrderParams{
		UserID: userID,
		Cursor: cursor,
		Limit:  int32(limit),
	})

	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": orders})
}

func (h *OrderHandler) GetOrdersByShopID(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromContext(r)

	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	orders, err := h.service.GetOrdersByShopID(r.Context(), userID)

	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": orders})
}

func (h *OrderHandler) ConfirmOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromContext(r)

	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	h.handleOrderAction(w, r, func(id uuid.UUID) (*repository.Order, error) {
		return h.service.ConfirmOrder(r.Context(), service.OrderActionParams{ID: id, UserID: userID})
	})
}

func (h *OrderHandler) ShipOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromContext(r)

	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	h.handleOrderAction(w, r, func(id uuid.UUID) (*repository.Order, error) {
		return h.service.ShipOrder(r.Context(), service.OrderActionParams{ID: id, UserID: userID})
	})
}

func (h *OrderHandler) DeliverOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromContext(r)

	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	h.handleOrderAction(w, r, func(id uuid.UUID) (*repository.Order, error) {
		return h.service.DeliverOrder(r.Context(), service.OrderActionParams{ID: id, UserID: userID})
	})
}

func (h *OrderHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromContext(r)

	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	h.handleOrderAction(w, r, func(id uuid.UUID) (*repository.Order, error) {
		return h.service.CancelOrder(r.Context(), service.OrderActionParams{ID: id, UserID: userID})
	})
}

func (h *OrderHandler) handleOrderAction(
	w http.ResponseWriter,
	r *http.Request,
	action func(uuid.UUID) (*repository.Order, error),
) {
	paramID := chi.URLParam(r, "id")
	id, err := uuid.Parse(paramID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid order id")
		return
	}

	order, err := action(id)
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": order})
}

func (h *OrderHandler) GetRevenueSummary(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	var fromDate, toDate *time.Time

	if fromDateStr := r.URL.Query().Get("from_date"); fromDateStr != "" {
		t, err := time.Parse("2006-01-02", fromDateStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_DATE", "invalid from_date format, use YYYY-MM-DD")
			return
		}
		fromDate = &t
	}

	if toDateStr := r.URL.Query().Get("to_date"); toDateStr != "" {
		t, err := time.Parse("2006-01-02", toDateStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_DATE", "invalid to_date format, use YYYY-MM-DD")
			return
		}
		toDate = &t
	}

	revenue, err := h.service.GetRevenueSummary(r.Context(), service.GetRevenueParams{
		UserID:   userID,
		FromDate: fromDate,
		ToDate:   toDate,
	})

	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": revenue})
}
