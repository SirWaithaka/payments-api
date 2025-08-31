package tanda

const (
	EndpointAuthentication    = "/v1/oauth2/token"
	EndpointPayments          = "/io/v3/organizations/%s/request"    // "/io/v3/organizations/{{organizationId}}/request"
	EndpointTransactionStatus = "/io/v3/organizations/%s/request/%s" // "/io/v3/organizations/{{organizationId}}/request/{{trackingId}}"
)

const (
	OperationAuthenticate      = "authenticate"
	OperationPayment           = "payment"
	OperationTransactionStatus = "transaction_status"
)
