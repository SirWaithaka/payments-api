package requests

type RequestPayment struct {
	Type string `json:"type" binding:"required"`

	//External identifier for the transfer which can be used for reconciliation. Need not be unique
	ExternalID string `json:"external_id" binding:"required"`
	//Unique idempotency identifier. Duplicates are rejected
	ExternalUID string `json:"external_uid" binding:"required"`

	Amount string `json:"amount" binding:"required"`

	ExternalWalletID string `json:"external_wallet_id" binding:"required"`
}
