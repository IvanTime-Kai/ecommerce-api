package service

import (
	"context"
	"fmt"

	"github.com/Ivantime-Kai/ecommerce-api/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserService struct {
	repository repository.Querier

	db *pgxpool.Pool
}

type CreateUserParams struct {
	FullName string
	Email    *string
	Phone    *string
	Password string
}

func NewUserService(repository repository.Querier, db *pgxpool.Pool) *UserService {
	return &UserService{
		repository: repository,
		db:         db,
	}
}

func (s *UserService) CreateUser(ctx context.Context, req *CreateUserParams) (*repository.User, error) {

	if req.Email == nil && req.Phone == nil {
		return nil, fmt.Errorf("Bad request")
	}

	var email pgtype.Text
	var phone pgtype.Text

	if req.Email != nil {
		email = toNullString(req.Email)
		isExistEmail, err := s.repository.CheckUserEmailExists(ctx, email)
		if err != nil {
			return nil, err
		}

		if isExistEmail {
			return nil, fmt.Errorf("Email exist")
		}
	}

	if req.Phone != nil {
		phone = toNullString(req.Phone)
		isExistPhone, err := s.repository.CheckUserPhoneExists(ctx, phone)
		if err != nil {
			return nil, err
		}

		if isExistPhone {
			return nil, fmt.Errorf("Phone exist")
		}
	}

	hashedPassword, err := hashPassword(req.Password)

	if err != nil {
		return nil, err
	}

	password := toText(hashedPassword)

	id, err := uuid.NewV7()

	if err != nil {
		return nil, err
	}

	tx, err := s.db.Begin(ctx)

	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	qtx := repository.New(tx)

	user, err := qtx.CreateUser(ctx, repository.CreateUserParams{
		ID:       id,
		FullName: req.FullName,
		Email:    email,
		Phone:    phone,
	})

	if err != nil {
		return nil, err
	}

	_, err = qtx.CreateUserAuth(ctx, repository.CreateUserAuthParams{
		UserID:       user.ID,
		Provider:     "local",
		PasswordHash: password,
	})

	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &user, nil
}
