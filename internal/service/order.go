package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Ivantime-Kai/ecommerce-api/internal/cache"
	"github.com/Ivantime-Kai/ecommerce-api/internal/kafka"
	"github.com/Ivantime-Kai/ecommerce-api/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrEmptyItems = fmt.Errorf("items cannot be empty")
var ErrInvalidProducts = fmt.Errorf("one or more products are invalid")
var ErrOutOfStock = fmt.Errorf("out of stock")
var ErrInvalidTransition = fmt.Errorf("invalid order transition")

type OrderService struct {
	repository repository.Querier
	db         *pgxpool.Pool
	kafka      kafka.KafkaProducer
	cacheStock *cache.StockCache
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
	IdempotencyKey   string
}

type OrderActionParams struct {
	ID     uuid.UUID
	UserID uuid.UUID
}

type GetOrderParams struct {
	UserID uuid.UUID
	Cursor *uuid.UUID
	Limit  int32
}

type GetOrdersResponse struct {
	Orders     []repository.Order `json:"orders"`
	NextCursor *uuid.UUID         `json:"next_cursor"`
}

type GetRevenueParams struct {
	UserID   uuid.UUID
	FromDate *time.Time
	ToDate   *time.Time
}

func NewOrderService(repository repository.Querier, db *pgxpool.Pool, kafka kafka.KafkaProducer, cacheStock *cache.StockCache) *OrderService {
	return &OrderService{
		repository: repository,
		db:         db,
		kafka:      kafka,
		cacheStock: cacheStock,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, req CreateOrderParams) (*repository.Order, error) {

	// Check Idempotency key from header
	if req.IdempotencyKey != "" {
		idempotencyKey, err := s.repository.GetIdempotencyKey(ctx, req.IdempotencyKey)

		if err == nil {
			var order repository.Order
			json.Unmarshal(idempotencyKey.Response, &order)
			return &order, nil
		}
	}

	// Handler func timeout after 10s
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	itemsLength := len(req.Items)

	if itemsLength == 0 {
		return nil, ErrEmptyItems
	}

	var deductedItems []OrderItemInput
	for _, item := range req.Items {
		ok, err := s.cacheStock.DeductStock(ctx, item.ProductID.String(), item.Quantity)
		if err != nil || !ok {
			for _, d := range deductedItems {
				s.cacheStock.RestoreStock(ctx, d.ProductID.String(), d.Quantity)
			}
			return nil, fmt.Errorf("%w: product %s", ErrOutOfStock, item.ProductID)
		}
		deductedItems = append(deductedItems, item)
	}

	user, err := s.repository.GetUserByID(ctx, req.UserID)

	if err != nil {
		return nil, err
	}

	tx, err := s.db.Begin(ctx)

	if err != nil {
		return nil, err
	}

	defer tx.Rollback(ctx)

	restoreStock := func() {
		for _, item := range req.Items {
			s.cacheStock.RestoreStock(ctx, item.ProductID.String(), item.Quantity)
		}
	}

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
		restoreStock()
		return nil, err
	}

	if len(products) != len(req.Items) {
		restoreStock()
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
		restoreStock()
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
		restoreStock()
		return nil, err
	}

	for _, item := range req.Items {
		product := productMap[item.ProductID]

		id, err := uuid.NewV7()

		if err != nil {
			restoreStock()
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
			restoreStock()
			return nil, err
		}

		_, err = qtx.DeductProductStock(ctx, repository.DeductProductStockParams{
			ID:    product.ID,
			Stock: int32(item.Quantity),
		})

		if errors.Is(err, pgx.ErrNoRows) {
			restoreStock()
			return nil, fmt.Errorf("%w: product %s", ErrOutOfStock, product.ID)
		}

		if err != nil {
			restoreStock()
			return nil, err
		}
	}

	event := kafka.OrderCreatedEvent{
		UserID:      req.UserID.String(),
		OrderID:     order.ID.String(),
		BuyerEmail:  user.Email.String,
		TotalAmount: total,
	}

	outboxID, _ := uuid.NewV7()
	eventBytes, _ := json.Marshal(event)

	err = qtx.CreateOutboxEvent(ctx, repository.CreateOutboxEventParams{
		ID:        outboxID,
		EventType: kafka.TopicOrderCreated,
		Payload:   eventBytes,
	})

	if err != nil {
		restoreStock()
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		restoreStock()
		return nil, err
	}

	// go func() {
	// 	s.kafka.Publish(context.Background(), []byte(order.ID.String()), eventBytes)
	// }()

	responseBytes, _ := json.Marshal(order)

	if req.IdempotencyKey != "" {

		s.repository.CreateIdempotencyKey(ctx, repository.CreateIdempotencyKeyParams{
			Key:       req.IdempotencyKey,
			Response:  responseBytes,
			ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(24 * time.Hour), Valid: true},
		})
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

func (s *OrderService) GetOrdersByUserID(ctx context.Context, req GetOrderParams) (*GetOrdersResponse, error) {
	if req.Limit == 0 {
		req.Limit = 10
	}

	var cursor uuid.UUID

	if req.Cursor != nil {
		cursor = *req.Cursor
	}

	orders, err := s.repository.GetOrdersByUserIDWithCursor(ctx, repository.GetOrdersByUserIDWithCursorParams{
		UserID:  req.UserID,
		Column2: cursor,
		Limit:   req.Limit,
	})

	if err != nil {
		return nil, err
	}

	var nextCursor *uuid.UUID

	ordersLength := len(orders)

	if ordersLength == int(req.Limit) {
		last := orders[ordersLength-1].ID
		nextCursor = &last
	}

	return &GetOrdersResponse{
		Orders:     orders,
		NextCursor: nextCursor,
	}, nil
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

	tx, err := s.db.Begin(ctx)

	if err != nil {
		return nil, err
	}

	defer tx.Rollback(ctx)

	qtx := repository.New(tx)

	order, err = qtx.DeliverOrder(ctx, req.ID)

	if err != nil {
		return nil, err
	}

	err = qtx.UpsertRevenueSummary(ctx, repository.UpsertRevenueSummaryParams{
		ShopID: order.ShopID,
		Date:   pgtype.Date{Time: time.Now(), Valid: true},
		Total:  order.TotalAmount,
	})

	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
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

func (s *OrderService) GetRevenueSummary(ctx context.Context, req GetRevenueParams) ([]repository.GetRevenueSummaryRow, error) {
	shop, err := s.repository.GetShopByOwnerID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	fromDate := pgtype.Date{Time: time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC), Valid: true}
	toDate := pgtype.Date{Time: time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC), Valid: true}

	if req.FromDate != nil {
		fromDate = pgtype.Date{Time: *req.FromDate, Valid: true}
	}
	if req.ToDate != nil {
		toDate = pgtype.Date{Time: *req.ToDate, Valid: true}
	}

	return s.repository.GetRevenueSummary(ctx, repository.GetRevenueSummaryParams{
		ShopID:   shop.ID,
		FromDate: fromDate,
		ToDate:   toDate,
	})
}

func (s *OrderService) HandlePaymentCompleted(ctx context.Context, event kafka.PaymentCompletedEvent) error {
	orderID, err := uuid.Parse(event.OrderID)

	if err != nil {
		return err
	}

	_, err = s.repository.ConfirmOrder(ctx, orderID)

	return err
}

func (s *OrderService) HandlePaymentFailed(ctx context.Context, event kafka.PaymentFailedEvent) error {
	orderID, err := uuid.Parse(event.OrderID)

	if err != nil {
		return err
	}

	_, err = s.repository.CancelOrder(ctx, orderID)

	return err
}
