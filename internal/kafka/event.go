package kafka

type OrderCreatedEvent struct {
	OrderID     string  `json:"order_id"`
	UserID      string  `json:"user_id"`
	BuyerEmail  string  `json:"buyer_email"`
	TotalAmount float64 `json:"total_amount"`
}
