package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ivantime-Kai/ecommerce-api/internal/cache"
	"github.com/Ivantime-Kai/ecommerce-api/internal/config"
	"github.com/Ivantime-Kai/ecommerce-api/internal/db"
	"github.com/Ivantime-Kai/ecommerce-api/internal/handler"
	"github.com/Ivantime-Kai/ecommerce-api/internal/kafka"
	"github.com/Ivantime-Kai/ecommerce-api/internal/middleware"
	"github.com/Ivantime-Kai/ecommerce-api/internal/repository"
	"github.com/Ivantime-Kai/ecommerce-api/internal/service"
	"github.com/go-chi/chi/v5"
	middlewareChi "github.com/go-chi/chi/v5/middleware"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.LoadConfig()

	if err != nil {
		log.Fatal(err)
	}

	pool, err := db.Connect(cfg.DB.Url)

	if err != nil {
		log.Fatal(err)
	}

	defer pool.Close()

	kafkaProducer, err := kafka.NewProducer(cfg.Kafka.Broker, kafka.TopicOrderCreated)

	if err != nil {
		log.Fatal(err)
	}

	defer kafkaProducer.Close()

	kafkaConsumer := kafka.NewConsumer(cfg.Kafka.Broker, kafka.TopicOrderCreated, kafka.GroupID)

	notificationService := service.NewNotificationService(&cfg.SMTP)

	go kafkaConsumer.Consume(ctx, func(key, value []byte) error {
		slog.Info("received kafka message", "key", string(key))
		var event kafka.OrderCreatedEvent
		if err := json.Unmarshal(value, &event); err != nil {
			return err
		}
		return notificationService.SendOrderConfirmation(event)
	})

	defer kafkaConsumer.Close()

	redis, err := cache.NewRedisClient(cfg.Redis.Url)

	if err != nil {
		log.Fatal(err)
	}

	defer redis.Close()

	rateLimiter := middleware.NewRateLimiter(redis, 5, time.Minute)

	r := chi.NewRouter()
	r.Use(middlewareChi.Logger)
	r.Use(middlewareChi.Recoverer)
	r.Use(middlewareChi.Timeout(time.Duration(cfg.Server.RequestTimeout) * time.Second))

	q := repository.New(pool)

	userService := service.NewUserService(q, pool, &cfg.JWT)
	userHandler := handler.NewUserHandler(userService)

	shopService := service.NewShopService(q)
	shopHandler := handler.NewShopHandler(shopService)

	productService := service.NewProductService(q, redis)
	productHandler := handler.NewProductHandler(productService)

	addressService := service.NewAddressService(q, pool)
	addressHandler := handler.NewAddressHandler(addressService)

	orderService := service.NewOrderService(q, pool, kafkaProducer)
	orderHandler := handler.NewOrderHandler(orderService)

	r.Route("/api/v1", func(r chi.Router) {
		// AUTH
		r.Post("/auth/logout", userHandler.Logout)
		r.Post("/auth/register", userHandler.Register)
		r.With(rateLimiter.Limit).Post("/auth/login", userHandler.Login)
		r.Post("/auth/refresh-token", userHandler.RefreshToken)

		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(cfg.JWT.ApiSecret))

			// USER
			r.Get("/user/profile", userHandler.GetProfile)
			r.Post("/user/mfa/enable", userHandler.EnableMFA)
			r.Post("/user/mfa/verify", userHandler.VerifyMFA)

			// SHOP
			r.Post("/shops", shopHandler.CreateShop)
			r.Get("/shops/me", shopHandler.GetMyShop)
			r.Get("/shops/{id}", shopHandler.GetShopByID)

			// PRODUCT
			r.Post("/products", productHandler.CreateProduct)
			r.Get("/products/{id}", productHandler.GetProductByID)
			r.Get("/shops/{id}/products", productHandler.GetProductsByShopID)
			r.Put("/products/{id}", productHandler.UpdateProduct)
			r.Delete("/products/{id}", productHandler.DeleteProduct)

			// ADDRESS
			r.Post("/addresses", addressHandler.CreateAddress)
			r.Get("/addresses/{id}", addressHandler.GetAddressByID)
			r.Patch("/addresses/{id}/default", addressHandler.SetDefaultAddress)
			r.Delete("/addresses/{id}", addressHandler.DeleteAddress)

			// ORDER
			r.Post("/orders", orderHandler.CreateOrder)
			r.Post("/orders/{id}/confirm", orderHandler.ConfirmOrder)
			r.Post("/orders/{id}/ship", orderHandler.ShipOrder)
			r.Post("/orders/{id}/deliver", orderHandler.DeliverOrder)
			r.Post("/orders/{id}/cancel", orderHandler.CancelOrder)
			r.Get("/orders/me", orderHandler.GetOrdersByUserID)
			r.Get("/orders/shop", orderHandler.GetOrdersByShopID)
			r.Get("/orders/{id}", orderHandler.GetOrderByID)
		})
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Server.Port),
		Handler: r,
	}

	go func() {
		log.Printf("Server running on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Duration(cfg.Server.RequestTimeout)*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
