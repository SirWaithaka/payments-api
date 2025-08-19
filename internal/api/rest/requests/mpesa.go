package requests

type RequestMpesaPayment struct {
	//External identifier for the transfer which can be used for reconciliation. Need not be unique
	TransactionID string `json:"transaction_id" validate:"required"`
	//Unique idempotency identifier. Duplicates are rejected
	IdempotencyID string `json:"idempotency_id" validate:"required"`
	// payment amount
	Amount string `json:"amount" validate:"required,numeric,min=10"`
	// (optional) but required for transfers, this can be till or paybill
	ExternalAccountType string `json:"external_account_type" validate:"oneof=till paybill"`
	// customer account that will be charged
	ExternalAccountID string `json:"external_account_id" validate:"required"`
	// account number or name of the payment beneficiary
	Beneficiary string `json:"beneficiary"`
	// payment description
	Description string `json:"description"`
}
