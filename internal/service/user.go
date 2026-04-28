package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Ivantime-Kai/ecommerce-api/internal/config"
	"github.com/Ivantime-Kai/ecommerce-api/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repository repository.Querier
	db         *pgxpool.Pool
	cfg        *config.JWTConfig
}

type CreateUserParams struct {
	FullName string
	Email    *string
	Phone    *string
	Password string
}

type LoginParams struct {
	Email      *string
	Phone      *string
	Password   string
	OTP        *string
	IP         string
	UserAgent  string
	IsRemember bool
}

type LoginResponse struct {
	AccessToken  string           `json:"access_token"`
	RefreshToken string           `json:"refresh_token"`
	User         *repository.User `json:"user"`
}

type RefreshTokenParams struct {
	RefreshToken string
}

type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type EnableMFAResponse struct {
	Secret string `json:"secret"`
	QRCode string `json:"qr_code"`
}

func NewUserService(repository repository.Querier, db *pgxpool.Pool, cfg *config.JWTConfig) *UserService {
	return &UserService{
		repository: repository,
		db:         db,
		cfg:        cfg,
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

func (s *UserService) Login(ctx context.Context, req *LoginParams) (*LoginResponse, error) {
	if req.Email == nil && req.Phone == nil {
		return nil, fmt.Errorf("Bad request")
	}

	var user repository.User

	if req.Email != nil {
		email := toNullString(req.Email)
		res, err := s.repository.GetUserByEmail(ctx, email)

		if err != nil {
			return nil, err
		}

		user = res
	} else if req.Phone != nil {
		phone := toNullString(req.Phone)
		res, err := s.repository.GetUserByPhone(ctx, phone)

		if err != nil {
			return nil, err
		}

		user = res
	}

	passwordHashRes, err := s.repository.GetUserPasswordHash(ctx, repository.GetUserPasswordHashParams{
		UserID:   user.ID,
		Provider: "local",
	})

	if err != nil {
		return nil, err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(passwordHashRes.String), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	isEnabledMFA, err := s.repository.CheckUserEnabledMFA(ctx, user.ID)

	if err != nil {
		return nil, err
	}

	if isEnabledMFA == true {
		if req.OTP == nil {
			return nil, fmt.Errorf("Bad request")
		}

		secretKey, err := s.repository.GetUserTOTPSecret(ctx, user.ID)

		if err != nil {
			return nil, fmt.Errorf("")
		}

		isOTP := totp.Validate(*req.OTP, secretKey)

		if !isOTP {
			return nil, fmt.Errorf("invalid OTP")
		}
	}

	accessToken, err := generateAccessToken(user.ID, s.cfg.ApiSecret, s.cfg.AccessTokenTTL)

	if err != nil {
		return nil, err
	}

	rawRefresh, hashedRefresh, err := generateRefreshToken(s.cfg.ApiSecret)

	if err != nil {
		return nil, err
	}

	sessionID, err := uuid.NewV7()

	if err != nil {
		return nil, err
	}

	expiresAt := pgtype.Timestamptz{
		Time:  time.Now().Add(time.Duration(s.cfg.RefreshTokenTTL) * time.Minute),
		Valid: true,
	}

	_, err = s.repository.CreateUserSession(ctx, repository.CreateUserSessionParams{
		ID:               sessionID,
		UserID:           user.ID,
		Ip:               toText(req.IP),
		UserAgent:        toText(req.UserAgent),
		RefreshTokenHash: hashedRefresh,
		ExpiresAt:        expiresAt,
		IsRemember:       req.IsRemember,
	})

	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		User:         &user,
	}, nil
}

func (s *UserService) GetProfile(ctx context.Context, userID uuid.UUID) (repository.User, error) {
	return s.repository.GetUserByID(ctx, userID)
}

func (s *UserService) RefreshToken(ctx context.Context, req *RefreshTokenParams) (*RefreshTokenResponse, error) {
	if req.RefreshToken == "" {
		return nil, fmt.Errorf("bad request")
	}

	hashedToken := hashToken(req.RefreshToken, s.cfg.ApiSecret)

	session, err := s.repository.GetSessionByRefreshTokenHash(ctx, hashedToken)

	if err != nil {
		return nil, err
	}

	if session.ExpiresAt.Time.Before(time.Now()) {
		return nil, fmt.Errorf("refresh token expired")
	}

	accessToken, err := generateAccessToken(session.UserID, s.cfg.ApiSecret, s.cfg.AccessTokenTTL)

	if err != nil {
		return nil, err
	}

	return &RefreshTokenResponse{
		AccessToken: accessToken,
	}, nil
}

func (s *UserService) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return fmt.Errorf("Bad request")
	}
	hashedToken := hashToken(refreshToken, s.cfg.ApiSecret)
	return s.repository.DeleteSessionByRefreshToken(ctx, hashedToken)
}

func (s *UserService) EnableMFA(ctx context.Context, userID uuid.UUID) (*EnableMFAResponse, error) {

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "ecommerce-api",
		AccountName: userID.String(),
	})

	if err != nil {
		return nil, err
	}

	secretKey := key.Secret()
	id, err := uuid.NewV7()

	if err != nil {
		return nil, err
	}

	userMFA, err := s.repository.CreateUserMFA(ctx, repository.CreateUserMFAParams{
		ID:        id,
		UserID:    userID,
		Method:    "totp",
		SecretKey: secretKey,
	})

	if err != nil {
		return nil, err
	}

	return &EnableMFAResponse{
		Secret: userMFA.SecretKey,
		QRCode: key.URL(),
	}, nil
}

func (s *UserService) VerifyMFA(ctx context.Context, userID uuid.UUID, otp string) error {
	userMFA, err := s.repository.GetUserMFAByUserID(ctx, userID)

	if err != nil {
		return err
	}

	if !totp.Validate(otp, userMFA.SecretKey) {
		return fmt.Errorf("invalid OTP")
	}

	if err := s.repository.EnableUserMFA(ctx, userID); err != nil {
		return err
	}

	return nil
}
