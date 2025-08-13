package daraja

const (
	ProductionUrl = "https://api.safaricom.co.ke"
	SandboxUrl    = "https://sandbox.safaricom.co.ke"

	EndpointAuthentication    = "/oauth/v1/generate"
	EndpointC2bExpress        = "/mpesa/stkpush/v1/processrequest"
	EndpointAccountBalance    = "/mpesa/accountbalance/v1/query"
	EndpointReversal          = "/mpesa/reversal/v1/request"
	EndpointC2bExpressQuery   = "/mpesa/stkpushquery/v1/query"
	EndpointTransactionStatus = "/mpesa/transactionstatus/v1/query"
	EndpointB2cPayment        = "/mpesa/b2c/v3/paymentrequest"
	EndpointB2bPayment        = "/mpesa/b2b/v1/paymentrequest"
	EndpointQueryOrgInfo      = "/sfcverify/v1/query/info"
)

const (
	OperationC2BExpress        = "express"
	OperationC2BQuery          = "stk_query"
	OperationReversal          = "reversal"
	OperationB2C               = "b2c"
	OperationB2B               = "b2b"
	OperationBalance           = "balance"
	OperationTransactionStatus = "search"
	OperationQueryOrgInfo      = "org_info_query"
)
