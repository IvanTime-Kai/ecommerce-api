package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Ivantime-Kai/ecommerce-api/internal/config"
	"github.com/Ivantime-Kai/ecommerce-api/internal/db"
	"github.com/Ivantime-Kai/ecommerce-api/internal/handler"
	"github.com/Ivantime-Kai/ecommerce-api/internal/repository"
	"github.com/Ivantime-Kai/ecommerce-api/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg, err := config.LoadConfig()

	if err != nil {
		log.Fatal(err)
	}

	pool, err := db.Connect(cfg.DB.Url)

	if err != nil {
		log.Fatal(err)
	}

	defer pool.Close()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	q := repository.New(pool)
	userService := service.NewUserService(q, pool)
	userHandler := handler.NewUserHandler(userService)

	r.Post("/v1/users/register", userHandler.Register)

	log.Printf("Server running on port %s", cfg.Server.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", cfg.Server.Port), r))
}
