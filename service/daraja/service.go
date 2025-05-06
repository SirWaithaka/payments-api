package daraja

import (
	"context"
	"net/http"

	"github.com/SirWaithaka/payments-api/request"
)

const (
	OperationC2BExpress        = "express"
	OperationC2BQuery          = "stk_query"
	OperationReversal          = "reversal"
	OperationB2C               = "b2c"
	OperationB2B               = "b2b"
	OperationTransactionStatus = "search"
	OperationBalance           = "balance"
	OperationQueryOrgInfo      = "org_info_query"
)

// Daraja provides the API operation methods for making requests
// to MPESA Daraja service.
type Daraja struct {
	hooks request.Hooks
}

func New(hooks request.Hooks) Daraja {
	return Daraja{hooks: hooks}
}

func (daraja Daraja) C2BExpressRequest(input RequestC2BExpress) (*request.Request, ResponseC2BExpress) {
	op := &request.Operation{
		Name:   OperationC2BExpress,
		Method: http.MethodPost,
		Path:   EndpointC2bExpress,
	}

	output := &ResponseC2BExpress{}
	return request.New(daraja.hooks, op, input, output), *output
}

func (daraja Daraja) C2BExpress(ctx context.Context, request RequestC2BExpress) (ResponseC2BExpress, error) {

	req, out := daraja.C2BExpressRequest(request)
	req.WithContext(ctx)

	if err := req.Send(); err != nil {
		return ResponseC2BExpress{}, err
	}

	return out, nil
}
