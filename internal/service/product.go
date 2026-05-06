package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Ivantime-Kai/ecommerce-api/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
)

type ProductService struct {
	repository repository.Querier
	redis      *redis.Client
}

type CreateProductParams struct {
	UserID      uuid.UUID
	Name        string
	Description *string
	Price       float64
	Stock       int32
}

type UpdateProductParams struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Name        string
	Description *string
	Status      repository.ProductStatus
}

type DeleteProductParams struct {
	ID     uuid.UUID
	UserID uuid.UUID
}

func NewProductService(repository repository.Querier, redis *redis.Client) *ProductService {
	return &ProductService{
		repository: repository,
		redis:      redis,
	}
}

func productCacheKey(id uuid.UUID) string {
	return fmt.Sprintf("product:%s", id)
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
		Price:       floatToNumeric(req.Price),
		Stock:       req.Stock,
	})

	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (s *ProductService) GetProductByID(ctx context.Context, id uuid.UUID) (repository.Product, error) {

	productCacheString, err := s.redis.Get(ctx, productCacheKey(id)).Result()

	if err != nil {
		product, err := s.repository.GetProductByID(ctx, id)

		if err != nil {
			return repository.Product{}, err
		}

		productBytes, _ := json.Marshal(product)
		s.redis.Set(ctx, productCacheKey(id), productBytes, time.Hour)

		return product, nil
	}
	var productCache repository.Product

	json.Unmarshal([]byte(productCacheString), &productCache)

	return productCache, nil
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
		Status:      req.Status,
	})

	if err != nil {
		return nil, err
	}

	s.redis.Del(ctx, productCacheKey(req.ID))

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

	err = s.repository.DeleteProduct(ctx, req.ID)

	if err != nil {
		return err
	}

	s.redis.Del(ctx, productCacheKey(req.ID))

	return nil
}
