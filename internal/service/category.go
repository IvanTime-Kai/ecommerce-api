package service

import (
	"context"

	"github.com/Ivantime-Kai/ecommerce-api/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type CategoryService struct {
	repository repository.Querier
}

func NewCategoryService(repository repository.Querier) *CategoryService {
	return &CategoryService{
		repository: repository,
	}
}

func (s *CategoryService) GetCategories(ctx context.Context) ([]repository.Category, error) {
	return s.repository.GetCategories(ctx)
}

func (s *CategoryService) GetSubCategories(ctx context.Context, parentID uuid.UUID) ([]repository.Category, error) {
	return s.repository.GetSubcategories(ctx, pgtype.UUID{
		Bytes: parentID,
		Valid: true,
	})
}
