package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type DBConfig struct {
	Url string
}

type RedisConfig struct {
	Url string
}

type ServerConfig struct {
	Port           string
	Mode           string
	RequestTimeout int
}

type JWTConfig struct {
	ApiSecret       string
	AccessTokenTTL  int
	RefreshTokenTTL int
}

type KafkaConfig struct {
	Broker string
}

type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
}

type Config struct {
	DB     DBConfig
	Redis  RedisConfig
	Server ServerConfig
	JWT    JWTConfig
	Kafka  KafkaConfig
	SMTP   SMTPConfig
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

	requestTimeout, err := strconv.Atoi(os.Getenv("SERVER_REQUEST_TIMEOUT"))
	if err != nil {
		return nil, err
	}

	return &Config{
		DB: DBConfig{
			Url: os.Getenv("DB_URL"),
		},
		Redis: RedisConfig{
			Url: os.Getenv("REDIS_URL"),
		},
		Server: ServerConfig{
			Port:           os.Getenv("SERVER_PORT"),
			Mode:           os.Getenv("SERVER_MODE"),
			RequestTimeout: requestTimeout,
		},
		JWT: JWTConfig{
			ApiSecret:       os.Getenv("JWT_SECRET"),
			AccessTokenTTL:  accessTokenTTL,
			RefreshTokenTTL: refreshTokenTTL,
		},
		Kafka: KafkaConfig{
			Broker: os.Getenv("KAFKA_BROKER"),
		},
		SMTP: SMTPConfig{
			Host:     os.Getenv("SMTP_HOST"),
			Port:     os.Getenv("SMTP_PORT"),
			Username: os.Getenv("SMTP_USERNAME"),
			Password: os.Getenv("SMTP_PASSWORD"),
		},
	}, nil
}
