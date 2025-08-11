package requests

type RequestMpesaPayment struct {
	//External identifier for the transfer which can be used for reconciliation. Need not be unique
	TransactionID string `json:"transaction_id"`
	//Unique idempotency identifier. Duplicates are rejected
	IdempotencyID string `json:"idempotency_id"`
	// payment amount
	Amount string `json:"amount"`
	// (optional) for transfers, this can be till_number or pay_bill
	ExternalAccountType string `json:"external_account_type"`
	// customer account that will be charged
	ExternalAccountID string `json:"external_account_id"`
	// account number or name of the payment beneficiary
	Beneficiary string `json:"beneficiary"`
	// payment description
	Description string `json:"description"`
}
