package service

import (
	"context"

	"github.com/Ivantime-Kai/ecommerce-api/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AddressService struct {
	repository repository.Querier
	db         *pgxpool.Pool
}

func NewAddressService(repository repository.Querier, db *pgxpool.Pool) *AddressService {
	return &AddressService{
		repository: repository,
		db:         db,
	}
}

type CreateAddressParams struct {
	UserID    uuid.UUID
	FullName  string
	Phone     string
	Province  string
	District  string
	Ward      string
	Street    string
	IsDefault bool
}

type SetDefaultAddressParams struct {
	ID     uuid.UUID
	UserID uuid.UUID
}

type DeleteAddressParams struct {
	ID     uuid.UUID
	UserID uuid.UUID
}

func (s *AddressService) CreateAddress(ctx context.Context, req CreateAddressParams) (*repository.Address, error) {
	tx, err := s.db.Begin(ctx)

	if err != nil {
		return nil, err
	}

	defer tx.Rollback(ctx)

	qtx := repository.New(tx)

	if req.IsDefault == true {
		err := qtx.SetDefaultAddress(ctx, req.UserID)

		if err != nil {
			return nil, err
		}
	}

	id, err := uuid.NewV7()

	if err != nil {
		return nil, err
	}

	address, err := qtx.CreateAddress(ctx, repository.CreateAddressParams{
		ID:        id,
		UserID:    req.UserID,
		FullName:  req.FullName,
		Phone:     req.Phone,
		Province:  req.Province,
		District:  req.District,
		Ward:      req.Ward,
		Street:    req.Street,
		IsDefault: req.IsDefault,
	})

	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &address, nil
}

func (s *AddressService) GetAddressByID(ctx context.Context, id uuid.UUID) (repository.Address, error) {
	return s.repository.GetAddressByID(ctx, id)
}

func (s *AddressService) SetDefaultAddress(ctx context.Context, req SetDefaultAddressParams) error {

	tx, err := s.db.Begin(ctx)

	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	qtx := repository.New(tx)

	err = qtx.SetDefaultAddress(ctx, req.UserID)

	if err != nil {
		return err
	}

	err = qtx.UpdateAddressDefault(ctx, req.ID)

	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (s *AddressService) DeleteAddress(ctx context.Context, req DeleteAddressParams) error {

	address, err := s.repository.GetAddressByID(ctx, req.ID)

	if err != nil {
		return err
	}

	if address.UserID != req.UserID {
		return ErrForbidden
	}

	return s.repository.DeleteAddress(ctx, req.ID)
}
