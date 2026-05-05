package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Ivantime-Kai/ecommerce-api/internal/service"
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
	ShippingPhone    string           `json:"shipping_full_phone"`
	ShippingProvince string           `json:"shipping_full_province"`
	ShippingDistrict string           `json:"shipping_full_district"`
	ShippingWard     string           `json:"shipping_full_ward"`
	ShippingStreet   string           `json:"shipping_full_street"`
	Items            []OrderItemInput `json:"items"`
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromContext(r)

	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

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
	})

	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"data": order})
}
