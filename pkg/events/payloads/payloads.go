package payloads

// PaymentProcessedPayload represents the payload for a payment-processed event
type PaymentProcessedPayload struct {
	PaymentID string  `json:"payment_id"`
	OrderID   string  `json:"order_id"`
	Amount    float64 `json:"amount"`
	Status    string  `json:"status"`
}
