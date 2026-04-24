package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(url string) (*pgxpool.Pool, error) {
	conn, err := pgxpool.New(context.Background(), url)

	if err != nil {
		return nil, err
	}

	return conn, nil
}
