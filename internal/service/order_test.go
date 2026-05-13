package service

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"

	"github.com/Ivantime-Kai/ecommerce-api/internal/cache"
	"github.com/Ivantime-Kai/ecommerce-api/internal/db"
	"github.com/Ivantime-Kai/ecommerce-api/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

var (
	testDB      *pgxpool.Pool
	testRepo    *repository.Queries
	cacheStock  *cache.StockCache
	redisClient *redis.Client
)

type mockKafkaProducer struct{}

func (m *mockKafkaProducer) Publish(ctx context.Context, key, value []byte) error {
	return nil
}

func (m *mockKafkaProducer) Close() error {
	return nil
}

func TestMain(m *testing.M) {
	pool, err := db.Connect(os.Getenv("DATABASE_URL"))

	if err != nil {
		panic(err)
	}

	defer pool.Close()

	testDB = pool
	testRepo = repository.New(pool)

	redisClient = redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_URL"),
	})

	cacheStock = cache.NewStockCache(redisClient)

	os.Exit(m.Run())
}

func TestCreateOrder_RaceCondition(t *testing.T) {
	ctx := context.Background()

	productID := uuid.MustParse("019e015f-845b-7d0f-9734-b259d5453c4c")

	err := cacheStock.SetStock(ctx, productID.String(), 1)
	if err != nil {
		t.Fatalf("failed to seed redis stock: %v", err)
	}

	_, err = testDB.Exec(ctx, "UPDATE products SET stock = 1 WHERE id = $1", productID)
	if err != nil {
		t.Fatalf("failed to reset db stock: %v", err)
	}

	orderService := NewOrderService(testRepo, testDB, &mockKafkaProducer{}, cacheStock)

	req := CreateOrderParams{
		UserID: uuid.MustParse("019e015e-d6db-70ea-9d36-a0e375666278"),
		ShopID: uuid.MustParse("019e015f-49e0-753b-b6e5-5707f87f49b9"),
		Items: []OrderItemInput{
			{ProductID: productID, Quantity: 1},
		},
		ShippingFullName: "Test User",
		ShippingPhone:    "0901234567",
		ShippingProvince: "Ho Chi Minh",
		ShippingDistrict: "Quan 1",
		ShippingWard:     "Phuong Ben Nghe",
		ShippingStreet:   "123 Test St",
	}

	var wg sync.WaitGroup

	results := make([]error, 2)

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, results[idx] = orderService.CreateOrder(ctx, req)
		}(i)
	}

	wg.Wait()

	// Assert: 1 success (nil), 1 ErrOutOfStock
	successCount := 0
	outOfStockCount := 0
	for i, err := range results {
		t.Logf("result[%d]: %v", i, err)
		if err == nil {
			successCount++
		} else if errors.Is(err, ErrOutOfStock) {
			outOfStockCount++
		}
	}

	assert.Equal(t, 1, successCount)
	assert.Equal(t, 1, outOfStockCount)
}
