package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Ivantime-Kai/ecommerce-api/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrEmptyItems = fmt.Errorf("items cannot be empty")
var ErrInvalidProducts = fmt.Errorf("one or more products are invalid")
var ErrOutOfStock = fmt.Errorf("out of stock")
var ErrInvalidTransition = fmt.Errorf("invalid order transition")

type OrderService struct {
	repository repository.Querier
	db         *pgxpool.Pool
}

type OrderItemInput struct {
	ProductID uuid.UUID
	Quantity  int
}

type OrderDetail struct {
	Order repository.Order       `json:"order"`
	Items []repository.OrderItem `json:"items"`
}

type CreateOrderParams struct {
	UserID           uuid.UUID
	ShopID           uuid.UUID
	ShippingFullName string
	ShippingPhone    string
	ShippingProvince string
	ShippingDistrict string
	ShippingWard     string
	ShippingStreet   string
	Items            []OrderItemInput
}

type OrderActionParams struct {
	ID     uuid.UUID
	UserID uuid.UUID
}

func NewOrderService(repository repository.Querier, db *pgxpool.Pool) *OrderService {
	return &OrderService{
		repository: repository,
		db:         db,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, req CreateOrderParams) (*repository.Order, error) {

	itemsLength := len(req.Items)

	if itemsLength == 0 {
		return nil, ErrEmptyItems
	}

	tx, err := s.db.Begin(ctx)

	if err != nil {
		return nil, err
	}

	defer tx.Rollback(ctx)

	qtx := repository.New(tx)

	productIDs := make([]uuid.UUID, len(req.Items))
	for i, item := range req.Items {
		productIDs[i] = item.ProductID
	}

	products, err := qtx.GetProductsForOrder(ctx, repository.GetProductsForOrderParams{
		Column1: productIDs,
		ShopID:  req.ShopID,
	})

	if err != nil {
		return nil, err
	}

	if len(products) != len(req.Items) {
		return nil, ErrInvalidProducts
	}

	productMap := make(map[uuid.UUID]repository.GetProductsForOrderRow, len(products))
	for _, p := range products {
		productMap[p.ID] = p
	}

	var total float64
	for _, item := range req.Items {
		product := productMap[item.ProductID]
		total += numericToFloat(product.Price) * float64(item.Quantity)
	}

	id, err := uuid.NewV7()

	if err != nil {
		return nil, err
	}

	order, err := qtx.CreateOrder(ctx, repository.CreateOrderParams{
		ID:               id,
		UserID:           req.UserID,
		ShopID:           req.ShopID,
		TotalAmount:      floatToNumeric(total),
		ShippingFullName: req.ShippingFullName,
		ShippingPhone:    req.ShippingPhone,
		ShippingProvince: req.ShippingProvince,
		ShippingDistrict: req.ShippingDistrict,
		ShippingWard:     req.ShippingWard,
		ShippingStreet:   req.ShippingStreet,
	})

	if err != nil {
		return nil, err
	}

	for _, item := range req.Items {
		product := productMap[item.ProductID]

		id, err := uuid.NewV7()

		if err != nil {
			return nil, err
		}

		_, err = qtx.CreateOrderItem(ctx, repository.CreateOrderItemParams{
			ID:          id,
			OrderID:     order.ID,
			ProductID:   product.ID,
			ProductName: product.Name,
			Quantity:    int32(item.Quantity),
			Price:       product.Price,
		})

		if err != nil {
			return nil, err
		}

		_, err = qtx.DeductProductStock(ctx, repository.DeductProductStockParams{
			ID:    product.ID,
			Stock: int32(item.Quantity),
		})

		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: product %s", ErrOutOfStock, product.ID)
		}

		if err != nil {
			return nil, err
		}

	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &order, nil
}

func (s *OrderService) GetOrderByID(ctx context.Context, id uuid.UUID) (*OrderDetail, error) {
	order, err := s.repository.GetOrderByID(ctx, id)
	if err != nil {
		return nil, err
	}

	items, err := s.repository.GetOrderItemsByOrderID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &OrderDetail{Order: order, Items: items}, nil
}

func (s *OrderService) GetOrdersByUserID(ctx context.Context, id uuid.UUID) ([]repository.Order, error) {
	return s.repository.GetOrdersByUserID(ctx, id)
}

func (s *OrderService) GetOrdersByShopID(ctx context.Context, userID uuid.UUID) ([]repository.Order, error) {
	shop, err := s.repository.GetShopByOwnerID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.repository.GetOrdersByShopID(ctx, shop.ID)
}

func (s *OrderService) ConfirmOrder(ctx context.Context, req OrderActionParams) (*repository.Order, error) {
	order, err := s.repository.GetOrderByID(ctx, req.ID)

	if err != nil {
		return nil, err
	}

	shop, err := s.repository.GetShopByOwnerID(ctx, req.UserID)

	if err != nil {
		return nil, err
	}

	if order.ShopID != shop.ID {
		return nil, fmt.Errorf("forbidden")
	}

	if order.Status != repository.OrderStatusPending {
		return nil, fmt.Errorf("%w: order is not in shipping status", ErrInvalidTransition)
	}

	order, err = s.repository.ConfirmOrder(ctx, req.ID)

	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (s *OrderService) ShipOrder(ctx context.Context, req OrderActionParams) (*repository.Order, error) {
	order, err := s.repository.GetOrderByID(ctx, req.ID)

	if err != nil {
		return nil, err
	}

	shop, err := s.repository.GetShopByOwnerID(ctx, req.UserID)

	if err != nil {
		return nil, err
	}

	if order.ShopID != shop.ID {
		return nil, fmt.Errorf("forbidden")
	}

	if order.Status != repository.OrderStatusConfirmed {
		return nil, fmt.Errorf("%w: order is not in shipping status", ErrInvalidTransition)
	}

	order, err = s.repository.ShipOrder(ctx, req.ID)

	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (s *OrderService) DeliverOrder(ctx context.Context, req OrderActionParams) (*repository.Order, error) {
	order, err := s.repository.GetOrderByID(ctx, req.ID)

	if err != nil {
		return nil, err
	}

	shop, err := s.repository.GetShopByOwnerID(ctx, req.UserID)

	if err != nil {
		return nil, err
	}

	if order.ShopID != shop.ID {
		return nil, fmt.Errorf("forbidden")
	}

	if order.Status != repository.OrderStatusShipping {
		return nil, fmt.Errorf("%w: order is not in shipping status", ErrInvalidTransition)
	}

	order, err = s.repository.DeliverOrder(ctx, req.ID)

	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (s *OrderService) CancelOrder(ctx context.Context, req OrderActionParams) (*repository.Order, error) {

	order, err := s.repository.GetOrderByID(ctx, req.ID)

	if err != nil {
		return nil, err
	}

	if order.UserID != req.UserID {
		return nil, fmt.Errorf("forbidden")
	}

	if order.Status != repository.OrderStatusPending && order.Status != repository.OrderStatusConfirmed {
		return nil, fmt.Errorf("order cannot be cancelled status")
	}

	order, err = s.repository.CancelOrder(ctx, req.ID)

	if err != nil {
		return nil, err
	}

	return &order, nil
}
