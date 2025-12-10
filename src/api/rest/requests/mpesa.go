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

type RequestMpesaPaymentStatus struct {
	PaymentID        string `json:"payment_id"`
	TransactionID    string `json:"transaction_id"`
	PaymentReference string `json:"payment_reference"`
}

type RequestAddShortCode struct {
	Environment       string `json:"environment" validate:"required,oneof=sandbox production"`
	Service           string `json:"service" validate:"required,oneof=daraja quikk"`
	Type              string `json:"type" validate:"required,oneof=charge payout transfer"`
	ShortCode         string `json:"shortcode" validate:"required"`
	InitiatorName     string `json:"initiator_name" validate:"required"`
	InitiatorPassword string `json:"initiator_password" validate:"required"`
	Key               string `json:"key" validate:"required"`
	Secret            string `json:"secret" validate:"required"`
	Passphrase        string `json:"passphrase"`
}
