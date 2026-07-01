package service

import (
	"context"

	"github.com/Ivantime-Kai/ecommerce-api/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type ReviewService struct {
	repository repository.Querier
}

func NewReviewService(repository repository.Querier) *ReviewService {
	return &ReviewService{
		repository: repository,
	}
}

type CreateReviewParams struct {
	OrderID   uuid.UUID
	ProductID uuid.UUID
	UserID    uuid.UUID
	Rating    int
	Comment   string
}

type GetReviewsParams struct {
	ProductID uuid.UUID
	Cursor    *uuid.UUID
	Limit     int32
}

func (s *ReviewService) CreateReview(ctx context.Context, req CreateReviewParams) (*repository.Review, error) {
	order, err := s.repository.GetOrderByID(ctx, req.OrderID)

	if err != nil {
		return nil, ErrNotFound
	}

	if order.UserID != req.UserID {
		return nil, ErrForbidden
	}

	if order.Status != repository.OrderStatusDelivered {
		return nil, ErrOrderNotDelivered
	}

	_, err = s.repository.GetReviewByOrderAndProduct(ctx, repository.GetReviewByOrderAndProductParams{
		OrderID:   req.OrderID,
		ProductID: req.ProductID,
	})

	if err == nil {
		return nil, ErrAlreadyReviewed
	}

	id, _ := uuid.NewV7()

	review, err := s.repository.CreateReview(ctx, repository.CreateReviewParams{
		ID:        id,
		OrderID:   req.OrderID,
		ProductID: req.ProductID,
		UserID:    req.UserID,
		Rating:    int32(req.Rating),
		Comment:   pgtype.Text{String: req.Comment, Valid: req.Comment != ""},
	})

	if err != nil {
		return nil, err
	}

	return &review, nil
}

func (s *ReviewService) GetReviewsByProductID(ctx context.Context, req GetReviewsParams) ([]repository.GetReviewsByProductIDRow, error) {
	cursor := uuid.UUID{}
	if req.Cursor != nil {
		cursor = *req.Cursor
	}
	return s.repository.GetReviewsByProductID(ctx, repository.GetReviewsByProductIDParams{
		ProductID: req.ProductID,
		Column2:   cursor,
		Limit:     req.Limit,
	})
}
