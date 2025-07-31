package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/rs/xid"
	"github.com/rs/zerolog"

	"github.com/SirWaithaka/payments-api/clients/daraja"
	"github.com/SirWaithaka/payments-api/internal/domains/payments"
	"github.com/SirWaithaka/payments-api/internal/domains/requests"
	"github.com/SirWaithaka/payments-api/internal/pkg/logger"
	"github.com/SirWaithaka/payments-api/request"
)

const (
	serviceName = "daraja"
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

type ResponseDefault daraja.ResponseDefault

func (response ResponseDefault) ExternalID() string { return response.OriginatorConversationID }

type ResponseC2BExpressQuery daraja.ResponseC2BExpressQuery

func (response ResponseC2BExpressQuery) ExternalID() string {
	return response.MerchantRequestID
}

type ResponseC2BExpress daraja.ResponseC2BExpress

func (response ResponseC2BExpress) ExternalID() string {
	return response.MerchantRequestID
}

type ResponseOrgNameCheck daraja.ResponseOrgInfoQuery

func (response ResponseOrgNameCheck) ExternalID() string {
	return response.ConversationID
}

type OrgNameCheckResponse daraja.ResponseOrgInfoQuery

func (response OrgNameCheckResponse) ExternalID() string {
	return response.ConversationID
}

// WEBHOOK REQUEST MODELS

type PaymentAttributes struct {
	SenderNo        string `json:"senderNo"`
	SenderName      string `json:"senderName"`
	RecipientNo     string `json:"recipientNo"`
	RecipientName   string `json:"recipientName"`
	Amount          string `json:"amount"`
	MpesaReceiptID  string `json:"mpesaReceiptId"`
	TransactionDate string `json:"transactionDate"`
}

type WebhookResult struct {
	Type           string            `json:"type"`
	ResultCode     daraja.ResultCode `json:"resultCode"`
	ResultMessage  string            `json:"resultMessage"`
	ConversationID string            `json:"conversationID"`
	OriginationID  string            `json:"originatorID"`
	Status         Status            `json:"status"`
	Attributes     any               `json:"attributes"`
}

func (result WebhookResult) ExternalID() string {
	return result.ConversationID
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

func NewDarajaApi(client *daraja.Client, shortcode ShortCodeConfig, repo requests.Repository) DarajaApi {
	return DarajaApi{client: client, shortcode: shortcode, requestRepo: repo}
}

// DarajaApi provides an interface to the mpesa wallet
// through the daraja platform
type DarajaApi struct {
	client      *daraja.Client
	shortcode   ShortCodeConfig
	requestRepo requests.Repository
}

func (api DarajaApi) C2B(ctx context.Context, payment payments.WalletPayment) error {
	l := zerolog.Ctx(ctx)
	l.Debug().Msg("handling c2b payment")

	timestamp := daraja.NewTimestamp()
	l.Debug().Str(logger.LData, timestamp.String()).Msg("timestamp")
	password := daraja.PasswordEncode(api.shortcode.ShortCode, api.shortcode.Passphrase, timestamp.String())

	payload := daraja.RequestC2BExpress{
		BusinessShortCode: api.shortcode.ShortCode,
		Password:          password,
		Timestamp:         timestamp,
		TransactionType:   daraja.TypeCustomerPayBillOnline,
		Amount:            payment.Amount,
		PartyA:            payment.DestinationAccountNumber,
		PartyB:            api.shortcode.ShortCode,
		PhoneNumber:       payment.DestinationAccountNumber,
		CallBackURL:       webhook(api.shortcode.CallbackURL, daraja.OperationC2BExpress),
		AccountReference:  payment.ClientTransactionID,
		TransactionDesc:   payment.Description,
		//TransactionDesc:   fmt.Sprintf("C2B REF %s ID %s", payment.TransactionID, payment.PaymentID),
	}
	l.Debug().Any(logger.LData, payload).Msg("request payload")

	// initialize request recorder
	recorder := NewRequestRecorder(api.requestRepo)

	// create an instance of request and add a request recorder hook
	requestID := xid.New().String()
	out := &ResponseC2BExpress{}
	req, _ := api.client.C2BExpressRequest(payload, request.WithServiceName(serviceName))
	req.WithContext(ctx)
	req.Data = out
	req.Hooks.Send.PushFrontHook(recorder.RecordRequest(payment.PaymentID, requestID))
	req.Hooks.Complete.PushFrontHook(recorder.UpdateRequestResponse(requestID))

	if err := req.Send(); err != nil {
		l.Error().Err(err).Msg("client error")
		return err
	}
	l.Debug().Any(logger.LData, out).Msg("c2b response")

	return nil

}

func (api DarajaApi) B2C(ctx context.Context, payment payments.WalletPayment) error {
	l := zerolog.Ctx(ctx)
	l.Debug().Msg("handling b2c payment")

	credential, err := daraja.OpenSSLEncrypt(api.shortcode.InitiatorPassword, daraja.SandboxCertificate)
	if err != nil {
		l.Error().Err(err).Msg("error encrypting password")
		return err
	}

	payload := daraja.RequestB2C{
		OriginatorConversationID: payment.ClientTransactionID,
		InitiatorName:            api.shortcode.InitiatorName,
		SecurityCredential:       credential,
		CommandID:                daraja.CommandBusinessPayment,
		Amount:                   payment.Amount,
		PartyA:                   api.shortcode.ShortCode,
		PartyB:                   payment.DestinationAccountNumber,
		QueueTimeOutURL:          webhook(api.shortcode.CallbackURL, daraja.OperationB2C),
		ResultURL:                webhook(api.shortcode.CallbackURL, daraja.OperationB2C),
		Remarks:                  payment.Description,
		Occasion:                 "OK",
		//Remarks:                  fmt.Sprintf("B2C REF %s ID %s", payment.TransactionID, payment.PaymentID),
		//Occasion:                 fmt.Sprintf("B2C REF %s ID %s", payment.TransactionID, payment.PaymentID),
	}

	res, err := api.client.B2C(ctx, payload)
	if err != nil {
		l.Error().Err(err).Msg("client error")
		return err
	}
	l.Debug().Any(logger.LData, res).Msg("b2c response")

	return nil
}

func (api DarajaApi) B2B(ctx context.Context, payment payments.WalletPayment) error {
	l := zerolog.Ctx(ctx)
	l.Debug().Msg("handling b2b payment")

	credential, err := daraja.OpenSSLEncrypt(api.shortcode.InitiatorPassword, daraja.SandboxCertificate)
	if err != nil {
		l.Error().Err(err).Msg("error encrypting password")
		return err
	}

	payload := daraja.RequestB2B{
		Initiator:              api.shortcode.InitiatorName,
		SecurityCredential:     credential,
		CommandID:              daraja.CommandBusinessPayBill,
		SenderIdentifierType:   daraja.IdentifierOrgShortCode,
		RecieverIdentifierType: daraja.IdentifierOrgShortCode,
		Amount:                 payment.Amount,
		PartyA:                 api.shortcode.ShortCode,
		PartyB:                 payment.DestinationAccountNumber,
		AccountReference:       payment.Beneficiary,
		QueueTimeOutURL:        webhook(api.shortcode.CallbackURL, daraja.OperationB2B),
		ResultURL:              webhook(api.shortcode.CallbackURL, daraja.OperationB2B),
		Remarks:                payment.Description,
	}
	res, err := api.client.B2B(ctx, payload, request.WithLogger(request.NewDefaultLogger()), request.WithLogLevel(request.LogError))
	if err != nil {
		l.Error().Err(err).Msg("client error")
		return err
	}
	l.Debug().Any(logger.LData, res).Msg("b2b response")

	return nil
}

func (api DarajaApi) Reversal(ctx context.Context, payment requests.Payment) error {
	l := zerolog.Ctx(ctx)
	l.Debug().Msg("handling reversal")

	credential, err := daraja.OpenSSLEncrypt(api.shortcode.InitiatorPassword, daraja.SandboxCertificate)
	if err != nil {
		l.Error().Err(err).Msg("error encrypting password")
		return nil
	}

	payload := daraja.RequestReversal{
		Initiator:              api.shortcode.InitiatorName,
		SecurityCredential:     credential,
		CommandID:              daraja.CommandTransactionReversal,
		TransactionID:          payment.PaymentReference,
		ReceiverParty:          api.shortcode.ShortCode,
		ReceiverIdentifierType: daraja.IdentifierOrgOperatorUsername,
		Amount:                 payment.Amount,
		ResultURL:              webhook(api.shortcode.CallbackURL, daraja.OperationReversal),
		QueueTimeOutURL:        webhook(api.shortcode.CallbackURL, daraja.OperationReversal),
		Remarks:                fmt.Sprintf("REVERSAL REF %s ID %s", payment.ClientTransactionID, payment.PaymentID),
	}

	res, err := api.client.Reverse(ctx, payload)
	if err != nil {
		l.Error().Err(err).Msg("client error")
		return err
	}
	l.Debug().Any(logger.LData, res).Msg("reverse response")

	return nil

}

// Status calls the api to check transaction status
func (api DarajaApi) Status(ctx context.Context, payment requests.Payment) error {
	l := zerolog.Ctx(ctx)
	l.Debug().Msg("handling transaction status")

	credential, err := daraja.OpenSSLEncrypt(api.shortcode.InitiatorPassword, daraja.SandboxCertificate)
	if err != nil {
		l.Error().Err(err).Msg("error encrypting password")
		return err
	}

	payload := daraja.RequestTransactionStatus{
		Initiator:          api.shortcode.InitiatorName,
		SecurityCredential: credential,
		CommandID:          daraja.CommandTransactionStatus,
		PartyA:             api.shortcode.ShortCode,
		IdentifierType:     daraja.IdentifierOrgShortCode,
		ResultURL:          webhook(api.shortcode.CallbackURL, daraja.OperationTransactionStatus),
		QueueTimeOutURL:    webhook(api.shortcode.CallbackURL, daraja.OperationTransactionStatus),
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

func (api DarajaApi) Balance(ctx context.Context, shortcode ShortCodeConfig) error {
	l := zerolog.Ctx(ctx)
	l.Debug().Msg("handling balance")

	credential, err := daraja.OpenSSLEncrypt(api.shortcode.InitiatorPassword, daraja.SandboxCertificate)
	if err != nil {
		l.Error().Err(err).Msg("error encrypting password")
		return err
	}

	payload := daraja.RequestBalance{
		Initiator:          api.shortcode.InitiatorName,
		SecurityCredential: credential,
		CommandID:          daraja.CommandAccountBalance,
		PartyA:             api.shortcode.ShortCode,
		IdentifierType:     daraja.IdentifierOrgShortCode,
		Remarks:            "Wallet Balance",
		QueueTimeOutURL:    webhook(api.shortcode.CallbackURL, daraja.OperationBalance),
		ResultURL:          webhook(api.shortcode.CallbackURL, daraja.OperationBalance),
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

// TRANSFORMER FUNCTIONS

func c2bWebHookResult(body io.Reader) (WebhookResult, error) {

	var c2bResult daraja.WebhookRequestC2BExpress
	if err := jsoniter.NewDecoder(body).Decode(&c2bResult); err != nil {
		return WebhookResult{}, err
	}

	var wb WebhookResult

	wb.ResultCode = c2bResult.Body.StkCallback.ResultCode
	wb.ResultMessage = c2bResult.Body.StkCallback.ResultDesc
	wb.OriginationID = c2bResult.Body.StkCallback.MerchantRequestID
	wb.ConversationID = c2bResult.Body.StkCallback.CheckoutRequestID

	// check if result code is success
	if c2bResult.Body.StkCallback.ResultCode != daraja.ResultCodeSuccess {
		wb.Status = StatusFailed
		return wb, nil
	}

	// otherwise if result code was success, retrieve payment information

	var attributes PaymentAttributes
	for _, item := range c2bResult.Body.StkCallback.CallbackMetadata.Item {
		if item.Name == "Amount" {
			amount := item.Value.(float64)
			attributes.Amount = strconv.FormatFloat(amount, 'f', 0, 64)
		}
		if item.Name == "MpesaReceiptNumber" {
			attributes.MpesaReceiptID = item.Value.(string)
		}
		if item.Name == "PhoneNumber" {
			senderNo := item.Value.(float64)
			attributes.SenderNo = strconv.FormatFloat(senderNo, 'f', 0, 64)
		}
		if item.Name == "TransactionDate" {
			transactionDate := item.Value.(float64)
			attributes.TransactionDate = strconv.FormatFloat(transactionDate, 'f', 0, 64)
		}
	}

	// update attributes field in variable
	wb.Attributes = attributes
	wb.Status = StatusCompleted

	return wb, nil
}

func b2cWebhookResult(body io.Reader) (WebhookResult, error) {

	var b2cResult daraja.WebhookRequestB2C
	if err := jsoniter.NewDecoder(body).Decode(&b2cResult); err != nil {
		return WebhookResult{}, err
	}

	var wb WebhookResult

	wb.ResultCode = b2cResult.Result.ResultCode
	wb.ResultMessage = b2cResult.Result.ResultDesc
	wb.OriginationID = b2cResult.Result.OriginatorConversationID
	wb.ConversationID = b2cResult.Result.ConversationID

	// check if result code is success
	if b2cResult.Result.ResultCode != daraja.ResultCodeSuccess {
		wb.Status = StatusFailed
		return wb, nil
	}

	// loop through params
	var attributes PaymentAttributes
	for _, param := range b2cResult.Result.ResultParameters.ResultParameter {
		if param.Key == "TransactionAmount" {
			amount := param.Value.(float64)
			attributes.Amount = strconv.FormatFloat(amount, 'f', 2, 64)
		}
		if param.Key == "TransactionReceipt" {
			attributes.MpesaReceiptID = param.Value.(string)
		}
		if param.Key == "ReceiverPartyPublicName" {
			attributes.RecipientName = param.Value.(string)
		}
		if param.Key == "TransactionCompletedDateTime" {
			attributes.TransactionDate = param.Value.(string)
		}
	}

	// update attributes field
	wb.Attributes = attributes
	wb.Status = StatusCompleted

	return wb, nil
}

func b2bWebhookResult(body io.Reader) (WebhookResult, error) {

	var b2bResult daraja.WebhookRequestB2B
	if err := jsoniter.NewDecoder(body).Decode(&b2bResult); err != nil {
		return WebhookResult{}, err
	}

	var wb WebhookResult

	wb.ResultCode = b2bResult.Result.ResultCode
	wb.ResultMessage = b2bResult.Result.ResultDesc
	wb.OriginationID = b2bResult.Result.OriginatorConversationID
	wb.ConversationID = b2bResult.Result.ConversationID

	// check if result code is success
	if b2bResult.Result.ResultCode != daraja.ResultCodeSuccess {
		wb.Status = StatusFailed
		return wb, nil
	}

	// loop through params
	var attributes PaymentAttributes
	for _, param := range b2bResult.Result.ResultParameters.ResultParameter {
		if param.Key == "Amount" {
			amount := param.Value.(float64)
			attributes.Amount = strconv.FormatFloat(amount, 'f', 2, 64)
		}
		if param.Key == "TransactionReceipt" {
			attributes.MpesaReceiptID = param.Value.(string)
		}
		if param.Key == "ReceiverPartyPublicName" {
			attributes.RecipientName = param.Value.(string)
		}
		if param.Key == "TransCompletedTime" {
			transactionDate := param.Value.(float64)
			attributes.TransactionDate = strconv.FormatFloat(transactionDate, 'f', 2, 64)
		}
	}

	// check receipt id is not empty
	if attributes.MpesaReceiptID == "" {
		attributes.MpesaReceiptID = b2bResult.Result.TransactionID
	}

	// update attributes field
	wb.Attributes = attributes
	wb.Status = StatusCompleted

	return wb, nil
}

func transactionStatusWebhookResult(body io.Reader) (WebhookResult, error) {

	var searchResult daraja.WebhookRequestTransactionStatus
	if err := jsoniter.NewDecoder(body).Decode(&searchResult); err != nil {
		return WebhookResult{}, err
	}

	var wb WebhookResult

	wb.ResultCode = searchResult.Result.ResultCode
	wb.ResultMessage = searchResult.Result.ResultDesc
	wb.OriginationID = searchResult.Result.OriginatorConversationID
	wb.ConversationID = searchResult.Result.ConversationID

	// check if result code is success
	// unlike other webhooks, non-success result code does not mean the payment
	// request failed.
	if searchResult.Result.ResultCode != daraja.ResultCodeSuccess {
		return wb, nil
	}

	// loop through params
	var attributes PaymentAttributes
	for _, param := range searchResult.Result.ResultParameters.ResultParameter {
		if param.Key == "Amount" {
			amount := param.Value.(float64)
			attributes.Amount = strconv.FormatFloat(amount, 'f', 2, 64)
		}
		if param.Key == "InitiatedTime" {
			transactionDate := param.Value.(float64)
			attributes.TransactionDate = strconv.FormatFloat(transactionDate, 'f', 0, 64)
		}
		if param.Key == "ReceiptNo" {
			attributes.MpesaReceiptID = param.Value.(string)
		}
		if param.Key == "TransactionStatus" {
			status := param.Value.(string)
			wb.Status = ToStatus(status)
		}
		if param.Key == "DebitPartyName" {
			attributes.SenderName = param.Value.(string)
		}
		if param.Key == "CreditPartyName" {
			attributes.RecipientName = param.Value.(string)
		}
		if param.Key == "OriginatorConversationID" {
			wb.OriginationID = param.Value.(string)
		}
	}

	wb.Attributes = attributes

	return wb, nil
}

func NewWebhookProcessor() WebhookProcessor {
	return WebhookProcessor{}
}

type WebhookProcessor struct{}

// Process takes in the webhook data as io.Reader, parses into a struct and sets the requests.WebhookResult Data field
func (processor WebhookProcessor) Process(ctx context.Context, result *requests.WebhookResult) (requests.OptionsUpdatePayment, error) {
	l := zerolog.Ctx(ctx)

	var wb WebhookResult
	var err error

	r := bytes.NewReader(result.Bytes())
	switch result.Action {
	case string(daraja.OperationC2BExpress):
		wb, err = c2bWebHookResult(r)
	case string(daraja.OperationB2C):
		wb, err = b2cWebhookResult(r)
	case string(daraja.OperationB2B):
		wb, err = b2bWebhookResult(r)
	case string(daraja.OperationTransactionStatus):
		wb, err = transactionStatusWebhookResult(r)
	case string(daraja.OperationReversal):
		//TODO: Create for reversal
		return requests.OptionsUpdatePayment{}, errors.New("webhook processor for reversal not created")
	default:
		return requests.OptionsUpdatePayment{}, errors.New("action processor not defined")
	}
	if err != nil {
		l.Error().Err(err).Msg("error processing webhook")
		return requests.OptionsUpdatePayment{}, err
	}
	l.Debug().Any(logger.LData, wb).Msg("webhook result")

	// assign to Data field
	result.Data = wb

	// set payment update options depending on status
	options := requests.OptionsUpdatePayment{}
	if wb.Status == StatusFailed {
		status := requests.StatusFailed
		options.Status = &status

		return options, nil
	}

	// update payment status
	status := requests.StatusSucceeded
	options.Status = &status

	// retrieve payment details
	var attributes PaymentAttributes
	var ok bool
	if attributes, ok = wb.Attributes.(PaymentAttributes); !ok {
		return requests.OptionsUpdatePayment{}, nil
	}

	options.PaymentReference = &attributes.MpesaReceiptID
	//options.SenderAccountName = &attributes.SenderName
	//options.SenderAccountNo = &attributes.SenderNo
	//options.RecipientAccountName = &attributes.RecipientName
	//options.RecipientAccountNo = &attributes.RecipientNo

	return options, nil

}
