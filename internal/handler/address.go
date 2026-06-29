package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Ivantime-Kai/ecommerce-api/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type AddressHandler struct {
	service *service.AddressService
}

type CreateAddressRequest struct {
	FullName  string `json:"full_name"`
	Phone     string `json:"phone"`
	Province  string `json:"province"`
	District  string `json:"district"`
	Ward      string `json:"ward"`
	Street    string `json:"street"`
	IsDefault bool   `json:"is_default"`
}

func NewAddressHandler(service *service.AddressService) *AddressHandler {
	return &AddressHandler{
		service: service,
	}
}

func (h *AddressHandler) CreateAddress(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromContext(r)

	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	var req CreateAddressRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	address, err := h.service.CreateAddress(r.Context(), service.CreateAddressParams{
		UserID:    userID,
		FullName:  req.FullName,
		Phone:     req.Phone,
		Province:  req.Province,
		District:  req.District,
		Ward:      req.Ward,
		Street:    req.Street,
		IsDefault: req.IsDefault,
	})

	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"data": address})
}

func (h *AddressHandler) GetAddressByID(w http.ResponseWriter, r *http.Request) {
	paramId := chi.URLParam(r, "id")
	id, err := uuid.Parse(paramId)

	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid address id")
		return
	}

	address, err := h.service.GetAddressByID(r.Context(), id)

	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": address})
}

func (h *AddressHandler) SetDefaultAddress(w http.ResponseWriter, r *http.Request) {

	paramId := chi.URLParam(r, "id")
	id, err := uuid.Parse(paramId)

	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid address id")
		return
	}

	userID, ok := getUserIDFromContext(r)

	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	err = h.service.SetDefaultAddress(r.Context(), service.SetDefaultAddressParams{
		ID:     id,
		UserID: userID,
	})

	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": "set default addresses successfully"})
}

func (h *AddressHandler) DeleteAddress(w http.ResponseWriter, r *http.Request) {
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

	err = h.service.DeleteAddress(r.Context(), service.DeleteAddressParams{
		ID:     id,
		UserID: userID,
	})

	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": "address deleted successfully"})
}
