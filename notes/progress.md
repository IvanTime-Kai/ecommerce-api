# Progress

## 2026-07-15

**Đã làm:**
- Khảo sát lại toàn bộ event-driven flow thật trong code: `CreateOrder` → outbox (cùng transaction) → Postgres trigger `pg_notify` → `OutboxWorker` (Redis lock + fan-out) → `CircuitBreakerProducer` → Kafka `order.created` → chuỗi consumer (`payment-service` → `inventory-service` → `order-service-inventory(-failed)` + `notification` + `payment-service-refund`).
- Vẽ diagram lưu tại `docs/event-flow.html` (mermaid, mở trực tiếp bằng browser).

**Vấn đề thật phát hiện được khi đọc code (chưa fix):**
1. `OutboxWorker` và `OrderService` dùng chung 1 producer (`producers.CB`) hard-bound vào topic `order.created`. Cột `event_type` trong bảng `outbox` hiện chưa được dùng để chọn topic — nếu sau này thêm event type khác qua outbox sẽ bị gửi nhầm topic.
   - File liên quan: `internal/app/kafka.go` (`setupProducers`), `internal/worker/outbox.go`.
2. `Consumer.Consume` (`internal/kafka/consumer.go`) chỉ log lỗi rồi đọc tiếp khi handler fail — không có dead-letter topic, không retry ở tầng consumer (khác với outbox, vốn có retry x3 + trạng thái `failed`).

## Bước tiếp theo

Chọn 1 trong 2 hướng (cả hai đều là vấn đề thật, chưa quyết định thứ tự):
- [ ] Route outbox event theo `event_type` thay vì 1 producer cố định — học các cách map event_type → topic/producer, trade-off.
- [ ] Thêm DLQ/retry cho consumer — học retry topic pattern, dead-letter queue, at-least-once vs exactly-once.
