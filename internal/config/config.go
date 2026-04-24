package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type DBConfig struct {
	Url string
}

type ServerConfig struct {
	Port string
	Mode string
}

type JWTConfig struct {
	ApiSecret       string
	AccessTokenTTL  int
	RefreshTokenTTL int
}

type Config struct {
	DB     DBConfig
	Server ServerConfig
	JWT    JWTConfig
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()

	if err != nil {
		return nil, err
	}

	accessTokenTTL, err := strconv.Atoi(os.Getenv("JWT_ACCESS_TOKEN_TTL"))
	if err != nil {
		return nil, err
	}

	refreshTokenTTL, err := strconv.Atoi(os.Getenv("JWT_REFRESH_TOKEN_TTL"))
	if err != nil {
		return nil, err
	}

	return &Config{
		DB: DBConfig{
			Url: os.Getenv("DB_URL"),
		},
		Server: ServerConfig{
			Port: os.Getenv("SERVER_PORT"),
			Mode: os.Getenv("SERVER_MODE"),
		},
		JWT: JWTConfig{
			ApiSecret:       os.Getenv("JWT_SECRET"),
			AccessTokenTTL:  accessTokenTTL,
			RefreshTokenTTL: refreshTokenTTL,
		},
	}, nil
}
