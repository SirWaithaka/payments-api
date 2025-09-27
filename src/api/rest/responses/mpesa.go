package responses

type MpesaPaymentResponse struct {
	PaymentID     string `json:"payment_id"`
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
}
