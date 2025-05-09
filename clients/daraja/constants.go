package daraja

const (
	ENV_PRODUCTION = "PRODUCTION"
	ENV_SANDBOX    = "SANDBOX"

	productionUrl = "https://api.safaricom.co.ke"
	sandboxUrl    = "https://sandbox.safaricom.co.ke"

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
