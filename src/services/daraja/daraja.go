package daraja

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/rs/xid"
	"github.com/rs/zerolog"

	"github.com/SirWaithaka/payments-api/pkg/logger"
	"github.com/SirWaithaka/payments-api/src/domains/mpesa"
	"github.com/SirWaithaka/payments-api/src/domains/requests"
	"github.com/SirWaithaka/payments-api/src/services/hooks"
	daraja2 "github.com/SirWaithaka/payments/daraja"

	"github.com/SirWaithaka/payments-api/request"
)

const (
	serviceName = requests.PartnerDaraja
)

type Status string

const (
	StatusFailed    Status = "failed"
	StatusCompleted Status = "completed"
)

func ToStatus(status string) Status {
	switch strings.ToLower(status) {
	case "completed":
		return StatusCompleted
	case "failed":
		return StatusFailed
	case "cancelled":
		return StatusFailed
	default:
		return StatusFailed
	}
}

// RESPONSE MODELS

type ResponseDefault daraja2.ResponseDefault

func (response ResponseDefault) ExternalID() string { return response.OriginatorConversationID }

type ResponseC2BExpress daraja2.ResponseC2BExpress

func (response ResponseC2BExpress) ExternalID() string {
	return response.MerchantRequestID
}

// adds action to the path of base url
// https://<baseurl>/:action
func webhook(baseUrl string, action string) string {
	if baseUrl == "" {
		return ""
	}

	u, err := url.Parse(baseUrl)
	if err != nil {
		return ""
	}

	u.Path, err = url.JoinPath(u.Path, action)
	if err != nil {
		return u.String()
	}

	return u.String()
}

func NewDarajaApi(client *daraja2.Client, certificate string, shortcode mpesa.ShortCode, repo requests.Repository) DarajaApi {
	return DarajaApi{client: client, certificate: certificate, shortcode: shortcode, requestRepo: repo}
}

// DarajaApi provides an interface to the mpesa wallet
// through the daraja platform
type DarajaApi struct {
	certificate string
	client      *daraja2.Client
	shortcode   mpesa.ShortCode
	requestRepo requests.Repository
}

func (api DarajaApi) C2B(ctx context.Context, paymentID string, payment mpesa.PaymentRequest) error {
	l := zerolog.Ctx(ctx)
	l.Debug().Msg("handling c2b payment")

	timestamp := daraja2.NewTimestamp()
	l.Debug().Str(logger.LData, timestamp.String()).Msg("timestamp")
	password := daraja2.PasswordEncode(api.shortcode.ShortCode, api.shortcode.Passphrase, timestamp.String())

	payload := daraja2.RequestC2BExpress{
		BusinessShortCode: api.shortcode.ShortCode,
		Password:          password,
		Timestamp:         timestamp,
		TransactionType:   daraja2.TypeCustomerPayBillOnline,
		Amount:            payment.Amount,
		PartyA:            payment.ExternalAccountNumber,
		PartyB:            api.shortcode.ShortCode,
		PhoneNumber:       payment.ExternalAccountNumber,
		CallBackURL:       webhook(api.shortcode.CallbackURL, daraja2.OperationC2BExpress),
		AccountReference:  payment.ClientTransactionID,
		TransactionDesc:   payment.Description,
		//TransactionDesc:   fmt.Sprintf("C2B REF %s ID %s", payment.TransactionID, payment.PaymentID),
	}
	l.Debug().Any(logger.LData, payload).Msg("request payload")

	// initialize request recorder
	recorder := hooks.NewRequestRecorder(api.requestRepo)

	// create an instance of request and add the request recorder hook
	requestID := xid.New().String()
	out := &ResponseC2BExpress{}
	req, _ := api.client.C2BExpressRequest(payload, request.WithServiceName(serviceName.String()))
	req.WithContext(ctx)
	req.Data = out
	req.Hooks.Send.PushFrontHook(recorder.RecordRequest(paymentID, requestID))
	req.Hooks.Complete.PushFrontHook(recorder.UpdateRequestResponse(requestID))

	if err := req.Send(); err != nil {
		l.Error().Err(err).Msg("client error")
		return err
	}
	l.Debug().Any(logger.LData, out).Msg("c2b response")

	return nil

}

func (api DarajaApi) B2C(ctx context.Context, paymentID string, payment mpesa.PaymentRequest) error {
	l := zerolog.Ctx(ctx)
	l.Debug().Msg("handling b2c payment")

	credential, err := daraja2.OpenSSLEncrypt(api.shortcode.InitiatorPassword, api.certificate)
	if err != nil {
		l.Error().Err(err).Msg("error encrypting password")
		return err
	}

	payload := daraja2.RequestB2C{
		OriginatorConversationID: payment.ClientTransactionID,
		InitiatorName:            api.shortcode.InitiatorName,
		SecurityCredential:       credential,
		CommandID:                daraja2.CommandBusinessPayment,
		Amount:                   payment.Amount,
		PartyA:                   api.shortcode.ShortCode,
		PartyB:                   payment.ExternalAccountNumber,
		QueueTimeOutURL:          webhook(api.shortcode.CallbackURL, daraja2.OperationB2C),
		ResultURL:                webhook(api.shortcode.CallbackURL, daraja2.OperationB2C),
		Remarks:                  payment.Description,
		Occasion:                 "OK",
		//Remarks:                  fmt.Sprintf("B2C REF %s ID %s", payment.TransactionID, payment.PaymentID),
		//Occasion:                 fmt.Sprintf("B2C REF %s ID %s", payment.TransactionID, payment.PaymentID),
	}
	l.Debug().Any(logger.LData, payload).Msg("request payload")

	// configure and add a hook to record this request attempt
	recorder := hooks.NewRequestRecorder(api.requestRepo)
	// create an instance of request and add the request recorder
	requestID := xid.New().String()
	out := &ResponseDefault{}
	req, _ := api.client.B2CRequest(payload, request.WithServiceName(serviceName.String()))
	req.WithContext(ctx)
	req.Data = out
	req.Hooks.Send.PushFrontHook(recorder.RecordRequest(paymentID, requestID))
	req.Hooks.Complete.PushFrontHook(recorder.UpdateRequestResponse(requestID))

	if err = req.Send(); err != nil {
		l.Error().Err(err).Msg("client error")
		return err
	}
	l.Debug().Any(logger.LData, out).Msg("b2c response")

	return nil
}

func (api DarajaApi) B2B(ctx context.Context, paymentID string, payment mpesa.PaymentRequest) error {
	l := zerolog.Ctx(ctx)
	l.Debug().Msg("handling b2b payment")

	credential, err := daraja2.OpenSSLEncrypt(api.shortcode.InitiatorPassword, api.certificate)
	if err != nil {
		l.Error().Err(err).Msg("error encrypting password")
		return err
	}

	payload := daraja2.RequestB2B{
		Initiator:              api.shortcode.InitiatorName,
		SecurityCredential:     credential,
		CommandID:              daraja2.CommandBusinessPayBill,
		SenderIdentifierType:   daraja2.IdentifierOrgShortCode,
		RecieverIdentifierType: daraja2.IdentifierOrgShortCode,
		Amount:                 payment.Amount,
		PartyA:                 api.shortcode.ShortCode,
		PartyB:                 payment.ExternalAccountNumber,
		AccountReference:       payment.Beneficiary,
		QueueTimeOutURL:        webhook(api.shortcode.CallbackURL, daraja2.OperationB2B),
		ResultURL:              webhook(api.shortcode.CallbackURL, daraja2.OperationB2B),
		Remarks:                payment.Description,
	}
	if payment.ExternalAccountType == mpesa.AccountTypeTill {
		payload.CommandID = daraja2.CommandBusinessBuyGoods
	}

	l.Debug().Any(logger.LData, payload).Msg("request payload")

	// configure and add a hook to record this request attempt
	recorder := hooks.NewRequestRecorder(api.requestRepo)
	// create an instance of request and add the request recorder
	requestID := xid.New().String()
	out := &ResponseDefault{}
	req, _ := api.client.B2BRequest(payload, request.WithServiceName(serviceName.String()))
	req.WithContext(ctx)
	req.Data = out
	req.Hooks.Send.PushFrontHook(recorder.RecordRequest(paymentID, requestID))
	req.Hooks.Complete.PushFrontHook(recorder.UpdateRequestResponse(requestID))

	if err = req.Send(); err != nil {
		l.Error().Err(err).Msg("client error")
		return err
	}
	l.Debug().Any(logger.LData, out).Msg("b2b response")

	return nil
}

func (api DarajaApi) Reversal(ctx context.Context, payment mpesa.ReversalRequest) error {
	l := zerolog.Ctx(ctx)
	l.Debug().Msg("handling reversal")

	credential, err := daraja2.OpenSSLEncrypt(api.shortcode.InitiatorPassword, api.certificate)
	if err != nil {
		l.Error().Err(err).Msg("error encrypting password")
		return nil
	}

	payload := daraja2.RequestReversal{
		Initiator:              api.shortcode.InitiatorName,
		SecurityCredential:     credential,
		CommandID:              daraja2.CommandTransactionReversal,
		TransactionID:          payment.PaymentReference,
		ReceiverParty:          api.shortcode.ShortCode,
		ReceiverIdentifierType: daraja2.IdentifierOrgOperatorUsername,
		Amount:                 payment.Amount,
		ResultURL:              webhook(api.shortcode.CallbackURL, daraja2.OperationReversal),
		QueueTimeOutURL:        webhook(api.shortcode.CallbackURL, daraja2.OperationReversal),
		Remarks:                fmt.Sprintf("REVERSAL REF %s ID %s", payment.ClientTransactionID, payment.PaymentID),
	}
	l.Debug().Any(logger.LData, payload).Msg("request payload")

	// create an instance of request and add the request recorder
	out := &ResponseDefault{}
	req, _ := api.client.ReversalRequest(payload, request.WithServiceName(serviceName.String()))
	req.WithContext(ctx)
	req.Data = out
	// generate a unique request id
	requestID := xid.New().String()
	// create new instance of request record and add as hook
	recorder := hooks.NewRequestRecorder(api.requestRepo)
	req.Hooks.Send.PushFrontHook(recorder.RecordRequest(payment.PaymentID, requestID))
	req.Hooks.Complete.PushFrontHook(recorder.UpdateRequestResponse(requestID))

	if err = req.Send(); err != nil {
		l.Error().Err(err).Msg("client error")
		return err
	}
	l.Debug().Any(logger.LData, out).Msg("reverse response")

	return nil

}

// Status calls the api to check transaction status
func (api DarajaApi) Status(ctx context.Context, payment mpesa.Payment) error {
	l := zerolog.Ctx(ctx)
	l.Debug().Msg("handling transaction status")

	credential, err := daraja2.OpenSSLEncrypt(api.shortcode.InitiatorPassword, api.certificate)
	if err != nil {
		l.Error().Err(err).Msg("error encrypting password")
		return err
	}

	payload := daraja2.RequestTransactionStatus{
		Initiator:          api.shortcode.InitiatorName,
		SecurityCredential: credential,
		CommandID:          daraja2.CommandTransactionStatus,
		PartyA:             api.shortcode.ShortCode,
		IdentifierType:     daraja2.IdentifierOrgShortCode,
		ResultURL:          webhook(api.shortcode.CallbackURL, daraja2.OperationTransactionStatus),
		QueueTimeOutURL:    webhook(api.shortcode.CallbackURL, daraja2.OperationTransactionStatus),
		Remarks:            "OK",
		Occasion:           "OK",
	}

	// set payload parameters according to which values have been passed
	if payment.PaymentReference != "" {
		payload.TransactionID = &payment.PaymentReference
	} else {
		payload.OriginatorConversationID = &payment.ClientTransactionID
	}

	res, err := api.client.TransactionStatus(ctx, payload)
	if err != nil {
		l.Error().Err(err).Msg("client error")
		return err
	}
	l.Debug().Any(logger.LData, res).Msg("transaction status")

	return nil

}

func (api DarajaApi) Balance(ctx context.Context) error {
	l := zerolog.Ctx(ctx)
	l.Debug().Msg("handling balance")

	credential, err := daraja2.OpenSSLEncrypt(api.shortcode.InitiatorPassword, api.certificate)
	if err != nil {
		l.Error().Err(err).Msg("error encrypting password")
		return err
	}

	payload := daraja2.RequestBalance{
		Initiator:          api.shortcode.InitiatorName,
		SecurityCredential: credential,
		CommandID:          daraja2.CommandAccountBalance,
		PartyA:             api.shortcode.ShortCode,
		IdentifierType:     daraja2.IdentifierOrgShortCode,
		Remarks:            "Wallet Balance",
		QueueTimeOutURL:    webhook(api.shortcode.CallbackURL, daraja2.OperationBalance),
		ResultURL:          webhook(api.shortcode.CallbackURL, daraja2.OperationBalance),
	}

	res, err := api.client.Balance(ctx, payload)
	if err != nil {
		l.Error().Err(err).Msg("client error")
		return err
	}
	l.Debug().Any(logger.LData, res).Msg("balance response")

	return nil

}

//func (service DarajaApi) Namecheck(ctx context.Context, req RequestOrgNameCheck) (OrgNameCheckResponse, error) {
//	l := zerolog.Ctx(ctx)
//	l.Debug().Msg("handling namecheck")
//
//	payload := daraja.RequestOrgInfoQuery{
//		IdentifierType: daraja.IdentifierOrgShortCode,
//		Identifier:     req.OrgBusinessNo,
//	}
//
//	if req.OrgType == "buygoods" {
//		payload.IdentifierType = daraja.IdentifierTillNumber
//	}
//
//	var response OrgNameCheckResponse
//	err := service.b2c.Call(ctx, daraja.QueryOrgInfoRequest(payload), &response)
//	if err != nil {
//		return OrgNameCheckResponse{}, err
//	}
//
//	// check on response code
//	if response.ResponseCode == daraja.CheckSuccess || response.ResponseCode == daraja.SuccessSubmission {
//		response.ResponseCode = daraja.SuccessSubmission
//	}
//
//	return response, nil
//
//}
