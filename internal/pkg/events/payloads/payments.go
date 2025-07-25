package payloads

type PaymentCompleted struct {
	Reference  string `json:"reference"`  // original payment reference
	ReceiptRef string `json:"receiptRef"` // payment receipt ref from provider
	Amount     string `json:"amount"`     // amount
	Status     string `json:"status"`
	Sender     struct {
		AccountNo string `json:"accountNo"`
		Name      string `json:"name"`
	}
	Recipient struct {
		AccountNo string `json:"accountNo"`
		Name      string `json:"name"`
	}
}

// WebhookReceived payload for received webhook requests
type WebhookReceived[T any] struct {
	Content T `json:"content"`
}
