package handler

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type HealthHandler struct {
	db          *pgxpool.Pool
	redis       *redis.Client
	kafkaBroker string
}

type response struct {
	Status   string            `json:"status"`
	Services map[string]string `json:"services"`
}

func NewHealthHandler(db *pgxpool.Pool, redis *redis.Client, kafkaBroker string) *HealthHandler {
	return &HealthHandler{
		db:          db,
		redis:       redis,
		kafkaBroker: kafkaBroker,
	}
}

func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	services := make(map[string]string)
	status := "ok"

	// Check DB
	if err := h.db.Ping(ctx); err != nil {
		services["database"] = "error: " + err.Error()
		status = "degraded"
	} else {
		services["database"] = "ok"
	}

	// Check Redis
	if err := h.redis.Ping(ctx).Err(); err != nil {
		services["redis"] = "error: " + err.Error()
		status = "degraded"
	} else {
		services["redis"] = "ok"
	}

	// Check Kafka
	if err := checkKafka(h.kafkaBroker); err != nil {
		services["kafka"] = "error: " + err.Error()
		status = "degraded"
	} else {
		services["kafka"] = "ok"
	}

	w.Header().Set("Content-Type", "application/json")
	if status == "degraded" {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	json.NewEncoder(w).Encode(response{
		Status:   status,
		Services: services,
	})
}

func checkKafka(broker string) error {
	conn, err := net.DialTimeout("tcp", broker, 2*time.Second)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}
