package quikk

const (
	OperationAuthCheck         = "auth_check"
	OperationCharge            = "charge"
	OperationPayout            = "payout"
	OperationTransfer          = "transfer"
	OperationBalance           = "balance"
	OperationTransactionSearch = "transaction_search"
	OperationSearch            = "search"
)

const (
	ProductionUrl = "https://api.quikk.dev"
	SandboxUrl    = "https://tryapi.quikk.dev"

	EndpointAuthCheck         = "/v1/auth-check"
	EndpointCharge            = "/v1/mpesa/charge"
	EndpointPayout            = "/v1/mpesa/payouts"
	EndpointTransfer          = "/v1/mpesa/transfers"
	EndpointBalance           = "/v1/mpesa/searches/balance"
	EndpointTransactionSearch = "/v1/mpesa/searches/transaction"
)
