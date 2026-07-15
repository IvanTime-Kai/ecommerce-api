package app

import (
	"time"

	"github.com/Ivantime-Kai/ecommerce-api/internal/config"
	"github.com/Ivantime-Kai/ecommerce-api/internal/handler"
	"github.com/Ivantime-Kai/ecommerce-api/internal/middleware"
	"github.com/go-chi/chi/v5"
	middlewareChi "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func setupRoutes(
	cfg *config.Config,
	rateLimiter *middleware.RateLimiter,
	userHandler *handler.UserHandler,
	shopHandler *handler.ShopHandler,
	categoryHandler *handler.CategoryHandler,
	addressHandler *handler.AddressHandler,
	productHandler *handler.ProductHandler,
	searchHandler *handler.SearchHandler,
	orderHandler *handler.OrderHandler,
	reviewHandler *handler.ReviewHandler,
	healthHandler *handler.HealthHandler,
) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middlewareChi.Logger)
	r.Use(middlewareChi.Recoverer)
	r.Use(middlewareChi.Timeout(time.Duration(cfg.Server.RequestTimeout) * time.Second))
	r.Use(middleware.PrometheusMiddleware)
	r.Use(middleware.RequestID)
	r.Use(middleware.RequestLogger)

	r.Get("/health", healthHandler.Check)
	r.Handle("/metrics", promhttp.Handler())

	r.Route("/api/v1", func(r chi.Router) {
		r.With(rateLimiter.Limit).Post("/auth/logout", userHandler.Logout)
		r.With(rateLimiter.Limit).Post("/auth/login", userHandler.Login)
		r.With(rateLimiter.Limit).Post("/auth/register", userHandler.Register)
		r.With(rateLimiter.Limit).Post("/auth/refresh-token", userHandler.RefreshToken)

		r.Get("/products", productHandler.ListProducts)
		r.Get("/products/search", searchHandler.SearchProducts)
		r.Get("/products/{id}/reviews", reviewHandler.GetReviewsByProductID)
		r.Get("/categories", categoryHandler.GetCategories)
		r.Get("/categories/{id}/subcategories", categoryHandler.GetSubCategories)

		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(cfg.JWT.ApiSecret))
			r.Use(rateLimiter.LimitByUser)

			r.Get("/user/profile", userHandler.GetProfile)
			r.Post("/user/mfa/enable", userHandler.EnableMFA)
			r.Post("/user/mfa/verify", userHandler.VerifyMFA)

			r.Post("/shops", shopHandler.CreateShop)
			r.Get("/shops/me", shopHandler.GetMyShop)
			r.Get("/shops/{id}", shopHandler.GetShopByID)

			r.Post("/products", productHandler.CreateProduct)
			r.Get("/products/{id}", productHandler.GetProductByID)
			r.Get("/shops/{id}/products", productHandler.GetProductsByShopID)
			r.Put("/products/{id}", productHandler.UpdateProduct)
			r.Delete("/products/{id}", productHandler.DeleteProduct)

			r.Post("/addresses", addressHandler.CreateAddress)
			r.Get("/addresses/{id}", addressHandler.GetAddressByID)
			r.Patch("/addresses/{id}/default", addressHandler.SetDefaultAddress)
			r.Delete("/addresses/{id}", addressHandler.DeleteAddress)

			r.Post("/orders", orderHandler.CreateOrder)
			r.Post("/orders/{id}/confirm", orderHandler.ConfirmOrder)
			r.Post("/orders/{id}/ship", orderHandler.ShipOrder)
			r.Post("/orders/{id}/deliver", orderHandler.DeliverOrder)
			r.Post("/orders/{id}/cancel", orderHandler.CancelOrder)
			r.Get("/orders/me", orderHandler.GetOrdersByUserID)
			r.Get("/orders/shop", orderHandler.GetOrdersByShopID)
			r.Get("/orders/{id}", orderHandler.GetOrderByID)
			r.Get("/orders/revenue-summary", orderHandler.GetRevenueSummary)

			r.Post("/reviews", reviewHandler.CreateReview)
		})
	})

	return r
}
