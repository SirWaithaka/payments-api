package tanda

import (
	"errors"
	"fmt"
	"iter"
	"slices"

	jsoniter "github.com/json-iterator/go"

	"github.com/SirWaithaka/payments-api/request"
)

// getRequiredParametersForCommand returns the required parameter IDs for each command
func getRequiredParametersForCommand(commandID Command) []ParameterID {
	switch commandID {
	case CommandCustomerToMerchantMobileMoneyPayment, CommandMerchantToCustomerMobileMoneyPayment:
		return []ParameterID{
			ParameterIDAmount,
			ParameterIDShortCode,
			ParameterIDAccountNumber,
			ParameterIDNarration,
			ParameterIDIpnUrl,
		}

	case CommandMerchantToCustomerBankPayment:
		return []ParameterID{
			ParameterIDAmount,
			ParameterIDShortCode,
			ParameterIDAccountNumber,
			ParameterIDAccountName,
			ParameterIDBankCode,
			ParameterIDNarration,
			ParameterIDIpnUrl,
		}

	case CommandMerchantTo3rdPartyMerchantPayment, CommandMerchantToMerchantTandaPayment:
		return []ParameterID{
			ParameterIDAmount,
			ParameterIDPartyA,
			ParameterIDPartyB,
			ParameterIDNarration,
			ParameterIDIpnUrl,
		}

	case CommandMerchantTo3rdPartyBusinessPayment:
		return []ParameterID{
			ParameterIDAmount,
			ParameterIDShortCode,
			ParameterIDBusinessNumber,
			ParameterIDAccountReference,
			ParameterIDNarration,
			ParameterIDIpnUrl,
		}

	case CommandInternationalMoneyTransferBank:
		return []ParameterID{
			ParameterIDAmount,
			ParameterIDCurrency,
			ParameterIDMobileNumber,
			ParameterIDAccountName,
			ParameterIDAccountNumber,
			ParameterIDBankCode,
			ParameterIDSenderType,
			ParameterIDBeneficiaryType,
			ParameterIDBeneficiaryAddress,
			ParameterIDBeneficiaryActivity,
			ParameterIDBeneficiaryCountry,
			ParameterIDBeneficiaryEmailAddress,
			ParameterIDDocumentType,
			ParameterIDDocumentNumber,
			ParameterIDNarration,
			ParameterIDSenderName,
			ParameterIDSenderAddress,
			ParameterIDSenderPhoneNumber,
			ParameterIDSenderDocumentType,
			ParameterIDSenderDocumentNumber,
			ParameterIDSenderCountry,
			ParameterIDSenderCurrency,
			ParameterIDSenderSourceOfFunds,
			ParameterIDSenderPrincipalActivity,
			ParameterIDSenderBankCode,
			ParameterIDSenderEmailAddress,
			ParameterIDSenderPrimaryAccountNumber,
			ParameterIDSenderDateOfBirth,
			ParameterIDIpnUrl,
			ParameterIDShortCode,
		}

	case CommandInternationalMoneyTransferMobile:
		return []ParameterID{
			ParameterIDAmount,
			ParameterIDCurrency,
			ParameterIDMobileNumber,
			ParameterIDAccountName,
			ParameterIDAccountNumber,
			ParameterIDSenderType,
			ParameterIDSenderCompanyName,
			ParameterIDBeneficiaryType,
			ParameterIDBeneficiaryActivity,
			ParameterIDBeneficiaryCountry,
			ParameterIDDocumentType,
			ParameterIDDocumentNumber,
			ParameterIDNarration,
			ParameterIDSenderName,
			ParameterIDSenderPhoneNumber,
			ParameterIDSenderDocumentType,
			ParameterIDSenderDocumentNumber,
			ParameterIDSenderCountry,
			ParameterIDSenderCurrency,
			ParameterIDSenderSourceOfFunds,
			ParameterIDSenderPrincipalActivity,
			ParameterIDIpnUrl,
			ParameterIDShortCode,
		}

	default:
		return nil
	}
}

// Custom iterator that yields only the ID field
func parameterIDs(params []PaymentRequestParameter) iter.Seq[ParameterID] {
	return func(yield func(ParameterID) bool) {
		for _, param := range params {
			if !yield(param.ID) {
				return
			}
		}
	}
}

// PaymentParametersValidator is a build hook that checks that all required parameters
// in RequestPayment.Request are present for a given command id.
var PaymentParametersValidator = request.Hook{
	Name: "tanda.PaymentParametersValidator",
	Fn: func(r *request.Request) {
		// get and cast request payload
		payload, ok := r.Params.(RequestPayment)
		if !ok {
			r.Error = errors.New("invalid payload")
			return
		}

		params := payload.Request
		if params == nil || len(params) == 0 || payload.CommandID == "" {
			r.Error = errors.New("invalid payload")
			return
		}

		requiredParams := getRequiredParametersForCommand(payload.CommandID)
		if requiredParams == nil || len(requiredParams) == 0 {
			r.Error = fmt.Errorf("invalid command id: %s", payload.CommandID)
			return
		}

		// check that all required parameters are present in the payload parameters
		for _, param := range requiredParams {
			if !slices.Contains(slices.Collect(parameterIDs(params)), param) {
				r.Error = fmt.Errorf("missing required parameter: %s", param)
				return
			}
		}
	},
}

type errResponse struct {
	ErrorResponse
}

func (r errResponse) Error() string {
	return fmt.Sprintf("<%s> %s: %s", r.Status, r.ErrorResponse.Error, r.Description)
}

// ResponseDecoder parse the http.Response body into the property
// request.Request.Data, if the status code is successful
// Otherwise for failed requests, it will parse the error response
// into the property request.Request.Error
var ResponseDecoder = request.Hook{
	Name: "tanda.ResponseDecoder",
	Fn: func(r *request.Request) {
		// response formats for non-2xx status codes follow the same format
		if r.Response.StatusCode < 200 || r.Response.StatusCode >= 300 {
			response := &errResponse{}
			if err := jsoniter.NewDecoder(r.Response.Body).Decode(response.ErrorResponse); err != nil {
				r.Error = err
				return
			}
			r.Error = response
			return
		}

		if err := jsoniter.NewDecoder(r.Response.Body).Decode(r.Data); err != nil {
			r.Error = err
		}
	},
}
