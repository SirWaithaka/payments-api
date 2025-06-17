package services

import (
	"context"
	"fmt"
	"net/url"

	"github.com/rs/zerolog"

	"github.com/SirWaithaka/payments-api/clients/daraja"
	"github.com/SirWaithaka/payments-api/internal/domains/payments"
	"github.com/SirWaithaka/payments-api/internal/pkg/logger"
	"github.com/SirWaithaka/payments-api/request"
)

// given a base url and action, the function returns an url in the following format
// http[s]://<baseurl>?type=<action>
func webhook(baseUrl string, action string) string {
	if baseUrl == "" {
		return ""
	}

	u, err := url.Parse(baseUrl)
	if err != nil {
		return ""
	}

	q := u.Query()
	q.Add("type", action)
	u.RawQuery = q.Encode()
	return u.String()
}

func NewMpesaService(daraja *daraja.Client) MpesaService {
	return MpesaService{daraja: daraja}
}

// MpesaService provides an interface to the mpesa wallet
// through either daraja or quikk apis.
type MpesaService struct {
	daraja *daraja.Client
	//quikk  *quikk.Client
}

func (service MpesaService) C2B(ctx context.Context, shortcode payments.ShortCodeConfig, payment payments.Payment) error {
	l := zerolog.Ctx(ctx)
	l.Debug().Msg("handling c2b payment")

	timestamp := daraja.NewTimestamp()
	l.Debug().Str(logger.LData, timestamp.String()).Msg("timestamp")
	password := daraja.PasswordEncode(shortcode.ShortCode, shortcode.Passphrase, timestamp.String())

	payload := daraja.RequestC2BExpress{
		BusinessShortCode: shortcode.ShortCode,
		Password:          password,
		Timestamp:         timestamp,
		TransactionType:   daraja.TypeCustomerPayBillOnline,
		Amount:            payment.Amount,
		PartyA:            payment.ExternalAccountNumber,
		PartyB:            shortcode.ShortCode,
		PhoneNumber:       payment.ExternalAccountNumber,
		CallBackURL:       shortcode.CallbackURL,
		AccountReference:  payment.Reference,
		TransactionDesc:   fmt.Sprintf("C2B REF %s ID %s", payment.Reference, payment.ExternalID),
	}

	res, err := service.daraja.C2BExpress(ctx, payload)
	if err != nil {
		l.Error().Err(err).Msg("client error")
		return err
	}
	l.Debug().Any(logger.LData, res).Msg("c2b response")

	return nil

}

func (service MpesaService) B2C(ctx context.Context, shortcode payments.ShortCodeConfig, payment payments.Payment) error {
	l := zerolog.Ctx(ctx)
	l.Debug().Msg("handling b2c payment")

	credential, err := daraja.OpenSSLEncrypt(shortcode.InitiatorPassword, daraja.SandboxCertificate)
	if err != nil {
		l.Error().Err(err).Msg("error encrypting password")
		return err
	}

	payload := daraja.RequestB2C{
		OriginatorConversationID: payment.Reference,
		InitiatorName:            shortcode.InitiatorName,
		SecurityCredential:       credential,
		CommandID:                daraja.CommandBusinessPayment,
		Amount:                   payment.Amount,
		PartyA:                   shortcode.ShortCode,
		PartyB:                   payment.ExternalAccountNumber,
		Remarks:                  fmt.Sprintf("B2C REF %s ID %s", payment.Reference, payment.ExternalID),
		QueueTimeOutURL:          shortcode.CallbackURL,
		ResultURL:                shortcode.CallbackURL,
		Occasion:                 fmt.Sprintf("B2C REF %s ID %s", payment.Reference, payment.ExternalID),
	}

	res, err := service.daraja.B2C(ctx, payload)
	if err != nil {
		l.Error().Err(err).Msg("client error")
		return err
	}
	l.Debug().Any(logger.LData, res).Msg("b2c response")

	return nil
}

func (service MpesaService) B2B(ctx context.Context, shortcode payments.ShortCodeConfig, payment payments.Payment) error {
	l := zerolog.Ctx(ctx)
	l.Debug().Msg("handling b2b payment")

	credential, err := daraja.OpenSSLEncrypt(shortcode.InitiatorPassword, daraja.SandboxCertificate)
	if err != nil {
		l.Error().Err(err).Msg("error encrypting password")
		return err
	}

	payload := daraja.RequestB2B{
		Initiator:              shortcode.InitiatorName,
		SecurityCredential:     credential,
		CommandID:              daraja.CommandBusinessPayBill,
		SenderIdentifierType:   daraja.IdentifierOrgShortCode,
		RecieverIdentifierType: daraja.IdentifierOrgShortCode,
		Amount:                 payment.Amount,
		PartyA:                 shortcode.ShortCode,
		PartyB:                 payment.ExternalAccountNumber,
		AccountReference:       payment.BeneficiaryAccountNumber,
		Remarks:                fmt.Sprintf("B2B REF %s ID %s", payment.Reference, payment.ExternalID),
		QueueTimeOutURL:        shortcode.CallbackURL,
		ResultURL:              shortcode.CallbackURL,
	}
	res, err := service.daraja.B2B(ctx, payload, request.WithLogger(request.NewDefaultLogger()), request.WithLogLevel(request.LogDebugWithRequestErrors))
	if err != nil {
		l.Error().Err(err).Msg("client error")
		return err
	}
	l.Debug().Any(logger.LData, res).Msg("b2b response")

	return nil
}

func (service MpesaService) Reversal(ctx context.Context, shortcode payments.ShortCodeConfig, payment payments.Payment) error {
	l := zerolog.Ctx(ctx)
	l.Debug().Msg("handling reversal")

	credential, err := daraja.OpenSSLEncrypt(shortcode.InitiatorPassword, daraja.SandboxCertificate)
	if err != nil {
		l.Error().Err(err).Msg("error encrypting password")
		return nil
	}

	payload := daraja.RequestReversal{
		Initiator:              shortcode.InitiatorName,
		SecurityCredential:     credential,
		CommandID:              daraja.CommandTransactionReversal,
		TransactionID:          payment.PaymentReference,
		ReceiverParty:          shortcode.ShortCode,
		ReceiverIdentifierType: daraja.IdentifierOrgOperatorUsername,
		Amount:                 payment.Amount,
		ResultURL:              shortcode.CallbackURL,
		QueueTimeOutURL:        shortcode.CallbackURL,
		Remarks:                fmt.Sprintf("REVERSAL REF %s ID %s", payment.Reference, payment.ExternalID),
	}

	res, err := service.daraja.Reverse(ctx, payload)
	if err != nil {
		l.Error().Err(err).Msg("client error")
		return err
	}
	l.Debug().Any(logger.LData, res).Msg("reverse response")

	return nil

}

func (service MpesaService) TransactionStatus(ctx context.Context, shortcode payments.ShortCodeConfig, payment payments.Payment) error {
	l := zerolog.Ctx(ctx)
	l.Debug().Msg("handling transaction status")

	credential, err := daraja.OpenSSLEncrypt(shortcode.InitiatorPassword, daraja.SandboxCertificate)
	if err != nil {
		l.Error().Err(err).Msg("error encrypting password")
		return err
	}

	payload := daraja.RequestTransactionStatus{
		Initiator:          shortcode.InitiatorName,
		SecurityCredential: credential,
		CommandID:          daraja.CommandTransactionStatus,
		PartyA:             shortcode.ShortCode,
		IdentifierType:     daraja.IdentifierOrgShortCode,
		ResultURL:          shortcode.CallbackURL,
		QueueTimeOutURL:    shortcode.CallbackURL,
		Remarks:            "OK",
		Occasion:           "OK",
	}

	// set payload parameters according to which values have been passed
	if payment.PaymentReference != "" {
		payload.TransactionID = &payment.PaymentReference
	} else {
		payload.OriginatorConversationID = &payment.Reference
	}

	res, err := service.daraja.TransactionStatus(ctx, payload)
	if err != nil {
		l.Error().Err(err).Msg("client error")
		return err
	}
	l.Debug().Any(logger.LData, res).Msg("transaction status")

	return nil

}

func (service MpesaService) Balance(ctx context.Context, shortcode payments.ShortCodeConfig) error {
	l := zerolog.Ctx(ctx)
	l.Debug().Msg("handling balance")

	credential, err := daraja.OpenSSLEncrypt(shortcode.InitiatorPassword, daraja.SandboxCertificate)
	if err != nil {
		l.Error().Err(err).Msg("error encrypting password")
		return err
	}

	payload := daraja.RequestBalance{
		Initiator:          shortcode.InitiatorName,
		SecurityCredential: credential,
		CommandID:          daraja.CommandAccountBalance,
		PartyA:             shortcode.ShortCode,
		IdentifierType:     daraja.IdentifierOrgShortCode,
		Remarks:            "Wallet Balance",
		QueueTimeOutURL:    webhook(shortcode.CallbackURL, daraja.OperationBalance),
		ResultURL:          webhook(shortcode.CallbackURL, daraja.OperationBalance),
	}

	res, err := service.daraja.Balance(ctx, payload)
	if err != nil {
		l.Error().Err(err).Msg("client error")
		return err
	}
	l.Debug().Any(logger.LData, res).Msg("balance response")

	return nil

}

//func (service MpesaService) Namecheck(ctx context.Context, req RequestOrgNameCheck) (OrgNameCheckResponse, error) {
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
