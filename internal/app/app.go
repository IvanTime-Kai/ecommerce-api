package app

import (
	"context"
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
	"github.com/Ivantime-Kai/ecommerce-api/internal/search"
	"github.com/Ivantime-Kai/ecommerce-api/internal/service"
	"github.com/Ivantime-Kai/ecommerce-api/internal/worker"
)

type App struct {
	cfg       *config.Config
	server    *http.Server
	producers *Producers
	consumers *Consumers
	outbox    *worker.OutboxWorker
}

func New(cfg *config.Config) (*App, error) {
	ctx := context.Background()

	// DB
	pool, err := db.Connect(cfg.DB.Url)
	if err != nil {
		return nil, err
	}
	replicaPool, err := db.ConnectReplica(cfg.DB.ReplicaUrl)
	if err != nil {
		return nil, err
	}

	// Redis
	redisClient, err := cache.NewRedisClient(cfg.Redis.Url)
	if err != nil {
		return nil, err
	}

	// Kafka topics
	for _, topic := range []string{kafka.TopicOrderCreated, kafka.TopicPaymentCompleted} {
		if err := kafka.CreateTopic(ctx, cfg.Kafka.Broker, topic, 3); err != nil {
			slog.Warn("create topic", "error", err)
		}
	}

	// Producers
	producers, err := setupProducers(cfg)
	if err != nil {
		return nil, err
	}

	// Repository
	q := repository.New(pool)
	replicaQuery := repository.New(replicaPool)

	// Cache
	stockCache := cache.NewStockCache(redisClient)
	productCache := cache.NewProductCache(redisClient)
	lockCache := cache.NewLockCache(redisClient)
	idempotencyCache := cache.NewIdempotencyCache(redisClient, 24*time.Hour)
	rateLimiter := middleware.NewRateLimiter(redisClient, 5, time.Minute)

	// Load stock vào Redis
	products, err := q.GetAllProducts(ctx)
	if err != nil {
		return nil, err
	}
	for _, p := range products {
		stockCache.SetStock(ctx, p.ID.String(), int(p.Stock))
	}

	// Services
	userService := service.NewUserService(q, pool, &cfg.JWT)
	shopService := service.NewShopService(q)
	categoryService := service.NewCategoryService(q)
	addressService := service.NewAddressService(q, pool)
	productService := service.NewProductService(q, replicaQuery, redisClient, stockCache, productCache)
	orderService := service.NewOrderService(q, pool, producers.CB, stockCache)
	paymentService := service.NewPaymentService(producers.Payment, producers.Refund)
	inventoryService := service.NewInventoryService(q, pool, producers.Inventory, stockCache)
	reviewService := service.NewReviewService(q)
	notificationService := service.NewNotificationService(&cfg.SMTP)

	// Outbox worker
	outboxWorker := worker.NewOutboxWorker(cfg.DB.Url, q, map[string]kafka.KafkaProducer{
		kafka.TopicOrderCreated: producers.CB,
	}, lockCache)
	outboxWorker.Start(ctx)

	// Consumers
	consumers := setupConsumers(ctx, cfg, paymentService, orderService, inventoryService, idempotencyCache, notificationService)

	// Handlers
	esClient := search.NewClient(cfg.Elasticsearch.URL)
	healthHandler := handler.NewHealthHandler(pool, redisClient, cfg.Kafka.Broker)

	// Routes
	router := setupRoutes(
		cfg,
		rateLimiter,
		handler.NewUserHandler(userService),
		handler.NewShopHandler(shopService),
		handler.NewCategoryHandler(categoryService),
		handler.NewAddressHandler(addressService),
		handler.NewProductHandler(productService),
		handler.NewSearchHandler(esClient),
		handler.NewOrderHandler(orderService),
		handler.NewReviewHandler(reviewService),
		healthHandler,
	)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Server.Port),
		Handler: router,
	}

	return &App{
		cfg:       cfg,
		server:    srv,
		producers: producers,
		consumers: consumers,
		outbox:    outboxWorker,
	}, nil
}

func (a *App) Run() {
	go func() {
		log.Printf("Server running on port %s", a.cfg.Server.Port)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Duration(a.cfg.Server.RequestTimeout)*time.Second)
	defer cancel()

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	a.consumers.Close()
	a.consumers.Wait()
	a.producers.Close()
	a.outbox.Wait()

	log.Println("Server exited")
}
