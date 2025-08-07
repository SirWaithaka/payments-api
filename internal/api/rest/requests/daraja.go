package requests

type RequestPayment struct {
	Type string `json:"type" binding:"required"`

	//External identifier for the transfer which can be used for reconciliation. Need not be unique
	TransactionID string `json:"transaction_id" binding:"required"`
	//Unique idempotency identifier. Duplicates are rejected
	IdempotencyID string `json:"idempotency_id" binding:"required"`

	Amount string `json:"amount" binding:"required"`

	ExternalAccountID string `json:"external_account_id" binding:"required"`

	Description string `json:"description"`
}

type RequestWalletPayment struct {
	// code identifying the wallet provider
	BankCode string `json:"bank" binding:"required"`
	//External identifier for the transfer which can be used for reconciliation. Need not be unique
	TransactionID string `json:"transaction_id" binding:"required"`
	//Unique idempotency identifier. Duplicates are rejected
	IdempotencyID string `json:"idempotency_id" binding:"required"`
	// amount to be charged
	Amount string `json:"amount" binding:"required"`
	// customer account that will be charged
	ExternalAccountID string `json:"external_account_id" binding:"required"`
	// account number or name of the payment beneficiary
	Beneficiary string `json:"beneficiary"`
	// payment description
	Description string `json:"description"`
}

type RequestPaymentStatus struct {
	PaymentID        string `json:"payment_id"`
	TransactionID    string `json:"transaction_id"`
	PaymentReference string `json:"payment_reference"`
}
