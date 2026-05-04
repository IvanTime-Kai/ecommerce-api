package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Ivantime-Kai/ecommerce-api/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ShopService struct {
	repository repository.Querier
}

type CreateShopParams struct {
	OwnerID     uuid.UUID
	Name        string
	Description *string
	LogoURL     *string
}

func NewShopService(repository repository.Querier) *ShopService {
	return &ShopService{
		repository: repository,
	}
}

func (s *ShopService) CreateShop(ctx context.Context, req *CreateShopParams) (*repository.Shop, error) {

	_, err := s.repository.GetShopByOwnerID(ctx, req.OwnerID)

	if err == nil {
		return nil, fmt.Errorf("shop already exists")
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	id, err := uuid.NewV7()

	if err != nil {
		return nil, err
	}

	slug := generateSlug(req.Name)

	shop, err := s.repository.CreateShop(ctx, repository.CreateShopParams{
		ID:          id,
		OwnerID:     req.OwnerID,
		Name:        req.Name,
		Slug:        slug,
		Description: toNullString(req.Description),
		LogoUrl:     toNullString(req.LogoURL),
	})

	if err != nil {
		return nil, err
	}

	return &shop, nil
}

func (s *ShopService) GetShopByOwnerID(ctx context.Context, ownerID uuid.UUID) (repository.Shop, error) {
	return s.repository.GetShopByOwnerID(ctx, ownerID)
}

func (s *ShopService) GetShopByID(ctx context.Context, id uuid.UUID) (repository.Shop, error) {
	return s.repository.GeShopByID(ctx, id)
}
