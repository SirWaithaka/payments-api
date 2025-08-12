package daraja

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strconv"

	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog"

	"github.com/SirWaithaka/payments-api/clients/daraja"
	"github.com/SirWaithaka/payments-api/internal/domains/mpesa"
	"github.com/SirWaithaka/payments-api/internal/domains/requests"
	"github.com/SirWaithaka/payments-api/internal/pkg/logger"
)

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
	return result.OriginationID
}

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
func (processor WebhookProcessor) Process(ctx context.Context, result *requests.WebhookResult, out any) error {
	l := zerolog.Ctx(ctx)

	options, ok := (out).(*mpesa.OptionsUpdatePayment)
	if !ok {
		return errors.New("invalid type for options")
	}

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
		return errors.New("webhook processor for reversal not created")
	default:
		return errors.New("action processor not defined")
	}
	if err != nil {
		l.Error().Err(err).Msg("error processing webhook")
		return err
	}
	l.Debug().Any(logger.LData, wb).Msg("webhook result")

	// assign to Data field
	result.Data = wb

	// set payment update options depending on status
	if wb.Status == StatusFailed {
		status := requests.StatusFailed
		options.Status = &status

		return nil
	}

	// update payment status
	status := requests.StatusSucceeded
	options.Status = &status

	// retrieve payment details
	var attributes PaymentAttributes
	if attributes, ok = wb.Attributes.(PaymentAttributes); !ok {
		return nil
	}

	options.PaymentReference = &attributes.MpesaReceiptID
	//options.SenderAccountName = &attributes.SenderName
	//options.SenderAccountNo = &attributes.SenderNo
	//options.RecipientAccountName = &attributes.RecipientName
	//options.RecipientAccountNo = &attributes.RecipientNo

	return nil

}
