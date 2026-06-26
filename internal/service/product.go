package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Ivantime-Kai/ecommerce-api/internal/cache"
	"github.com/Ivantime-Kai/ecommerce-api/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

type ProductService struct {
	repository     repository.Querier // Primary — write
	readRepository repository.Querier // Replica — read
	redis          *redis.Client
	stockCache     *cache.StockCache
	productCache   *cache.ProductCache
	sfGroup        singleflight.Group
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

type SearchProductsParams struct {
	Query      string
	CategoryID uuid.UUID
	MinPrice   float64
	MaxPrice   float64
	Cursor     uuid.UUID
	Limit      int32
}

type SearchProductsResponse struct {
	Products   []repository.SearchProductsRow `json:"products"`
	NextCursor *uuid.UUID                     `json:"next_cursor`
}

func NewProductService(repository repository.Querier, readRepository repository.Querier, redis *redis.Client, stockCache *cache.StockCache, productCache *cache.ProductCache) *ProductService {
	return &ProductService{
		repository:     repository,
		readRepository: readRepository,
		redis:          redis,
		stockCache:     stockCache,
		productCache:   productCache,
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

	s.stockCache.SetStock(ctx, id.String(), int(req.Stock))
	if err := s.invalidateProductCache(ctx); err != nil {
		slog.Warn("failed to invalidate product cache", "error", err)
	}

	return &product, nil
}

func (s *ProductService) GetProductByID(ctx context.Context, id uuid.UUID) (repository.Product, error) {

	productCacheString, err := s.redis.Get(ctx, productCacheKey(id)).Result()

	if err != nil {
		product, err := s.readRepository.GetProductByID(ctx, id)

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
	return s.readRepository.GetProductsByShopID(ctx, id)
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
	if err := s.invalidateProductCache(ctx); err != nil {
		slog.Warn("failed to invalidate product cache", "error", err)
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

	err = s.repository.DeleteProduct(ctx, req.ID)

	if err != nil {
		return err
	}

	s.redis.Del(ctx, productCacheKey(req.ID))
	if err := s.invalidateProductCache(ctx); err != nil {
		slog.Warn("failed to invalidate product cache", "error", err)
	}

	return nil
}

func (s *ProductService) SearchProducts(ctx context.Context, req SearchProductsParams) (*SearchProductsResponse, error) {

	if req.Limit == 0 {
		req.Limit = 20
	}

	cacheKey := fmt.Sprintf("products:q=%s:cat=%s:min=%s:max=%s:cursor=%s:limit=%d",
		req.Query,
		req.CategoryID.String(),
		fmt.Sprintf("%.0f", req.MinPrice),
		fmt.Sprintf("%.0f", req.MaxPrice),
		req.Cursor.String(),
		req.Limit,
	)

	// check cache
	cached, err := s.redis.Get(ctx, cacheKey).Bytes()

	if err == nil {
		var resp SearchProductsResponse
		if err := json.Unmarshal(cached, &resp); err == nil {
			return &resp, nil
		}
	} else if err != redis.Nil {
		slog.Warn("redis unavailable, falling back to DB", "error", err)

		return s.queryProducts(ctx, SearchProductsParams{
			Query:      req.Query,
			CategoryID: req.CategoryID,
			MinPrice:   req.MinPrice,
			MaxPrice:   req.MaxPrice,
			Cursor:     req.Cursor,
			Limit:      req.Limit,
		})
	}

	// cache miss => query DB
	result, err, _ := s.sfGroup.Do(cacheKey, func() (any, error) {
		resp, err := s.queryProducts(ctx, SearchProductsParams{
			Query:      req.Query,
			CategoryID: req.CategoryID,
			MinPrice:   req.MinPrice,
			MaxPrice:   req.MaxPrice,
			Cursor:     req.Cursor,
			Limit:      req.Limit,
		})

		if err != nil {
			return nil, err
		}

		// set cache TTL 5 minutes
		if data, err := json.Marshal(resp); err == nil {
			s.redis.Set(ctx, cacheKey, data, 5*time.Minute)
		}

		return resp, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*SearchProductsResponse), nil
}

func (s *ProductService) invalidateProductCache(ctx context.Context) error {
	return s.productCache.DeleteByPattern(ctx, "products:*")
}

func (s *ProductService) queryProducts(ctx context.Context, req SearchProductsParams) (*SearchProductsResponse, error) {
	products, err := s.readRepository.SearchProducts(ctx, repository.SearchProductsParams{
		Column1: req.Query,
		Column2: req.CategoryID,
		Column3: floatToNumeric(req.MinPrice),
		Column4: floatToNumeric(req.MaxPrice),
		Column5: req.Cursor,
		Limit:   req.Limit,
	})

	if err != nil {
		return nil, err
	}

	var nextCursor *uuid.UUID

	if len(products) == int(req.Limit) {
		last := products[len(products)-1].ID
		nextCursor = &last
	}

	resp := &SearchProductsResponse{
		Products:   products,
		NextCursor: nextCursor,
	}

	return resp, nil
}
