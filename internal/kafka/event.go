package kafka

type OrderCreatedEvent struct {
	OrderID     string  `json:"order_id"`
	UserID      string  `json:"user_id"`
	BuyerEmail  string  `json:"buyer_email"`
	TotalAmount float64 `json:"total_amount"`
}

type PaymentCompletedEvent struct {
	OrderID     string  `json:"order_id"`
	BuyerEmail  string  `json:"buyer_email"`
	TotalAmount float64 `json:"total_amount"`
}

type PaymentFailedEvent struct {
	OrderID string `json:"order_id"`
	Reason  string `json:"reason"`
}

type InventoryReservedEvent struct {
	OrderID string `json:"order_id"`
}

type InventoryFailedEvent struct {
	OrderID string `json:"order_id"`
	Reason  string `json:"reason"`
}

type PaymentRefundEvent struct {
	OrderID     string  `json:"order_id"`
	TotalAmount float64 `json:"total_amount"`
}
