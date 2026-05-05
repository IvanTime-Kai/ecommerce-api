package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Ivantime-Kai/ecommerce-api/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ProductService struct {
	repository repository.Querier
}

type CreateProductParams struct {
	UserID      uuid.UUID
	Name        string
	Description *string
}

type UpdateProductParams struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Name        string
	Description *string
	IsActive    bool
}

type DeleteProductParams struct {
	ID     uuid.UUID
	UserID uuid.UUID
}

func NewProductService(repository repository.Querier) *ProductService {
	return &ProductService{
		repository: repository,
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, req CreateProductParams) (*repository.Product, error) {
	shop, err := s.repository.GetShopByOwnerID(ctx, req.UserID)

	if err != nil {
		return nil, err
	}

	id, err := uuid.NewV7()

	if err != nil {
		return nil, err
	}

	product, err := s.repository.CreateProduct(ctx, repository.CreateProductParams{
		ID:          id,
		ShopID:      shop.ID,
		Name:        req.Name,
		Description: toNullString(req.Description),
	})

	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (s *ProductService) GetProductByID(ctx context.Context, id uuid.UUID) (repository.Product, error) {
	return s.repository.GetProductByID(ctx, id)
}

func (s *ProductService) GetProductsByShopID(ctx context.Context, id uuid.UUID) ([]repository.Product, error) {
	return s.repository.GetProductsByShopID(ctx, id)
}

func (s *ProductService) UpdateProduct(ctx context.Context, req UpdateProductParams) (*repository.Product, error) {

	_, err := s.repository.GetProductByIDAndShopOwner(ctx, repository.GetProductByIDAndShopOwnerParams{
		ID:      req.ID,
		OwnerID: req.UserID,
	})

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("forbidden")
	}

	if err != nil {
		return nil, err
	}

	product, err := s.repository.UpdateProduct(ctx, repository.UpdateProductParams{
		ID:          req.ID,
		Name:        req.Name,
		Description: toNullString(req.Description),
		IsActive:    req.IsActive,
	})

	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (s *ProductService) DeleteProduct(ctx context.Context, req DeleteProductParams) error {

	_, err := s.repository.GetProductByIDAndShopOwner(ctx, repository.GetProductByIDAndShopOwnerParams{
		ID:      req.ID,
		OwnerID: req.UserID,
	})

	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("forbidden")
	}

	if err != nil {
		return err
	}

	return s.repository.DeleteProduct(ctx, req.ID)
}