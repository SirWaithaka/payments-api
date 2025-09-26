package quikk

import (
	"bytes"
	"context"
	"errors"

	jsoniter "github.com/json-iterator/go"

	"github.com/SirWaithaka/payments-api/pkg/types"
	"github.com/SirWaithaka/payments-api/src/domains/mpesa"
	"github.com/SirWaithaka/payments-api/src/domains/requests"
	quikk2 "github.com/SirWaithaka/payments/quikk"
)

// WEBHOOK REQUEST MODELS

type ChargeWebhook quikk2.WebhookResult[quikk2.WebhookAttributesCharge]

// ExternalID should match the id returned by the quikk api during the
// initial payment request. For charge api, this is the TxnChargeID field.
func (webhook ChargeWebhook) ExternalID() string {
	return webhook.Data.Attributes.TxnChargeID
}

type PayoutWebhook quikk2.WebhookResult[quikk2.WebhookAttributesPayout]

// ExternalID should match the id returned by the quikk api during the
// initial payment request. For payout api, this is the ResponseID field.
func (webhook PayoutWebhook) ExternalID() string {
	return webhook.Data.Attributes.ResponseID
}

type TransferWebhook quikk2.WebhookResult[quikk2.WebhookAttributesTransfer]

// ExternalID should match the id returned by the quikk api during the
// initial payment request. For payout api, this is the ResponseID field.
func (webhook TransferWebhook) ExternalID() string {
	return webhook.Data.Attributes.ResponseID
}

type TransactionSearchWebhook quikk2.WebhookResult[quikk2.WebhookAttributesTransactionSearch]

// ExternalID should match the id returned during the initial payment request.
// For search api, check the TxnType field, if the field equals "payin", externalID
// should be set to the ResourceID, otherwise, externalID should be set to ResponseID
func (webhook TransactionSearchWebhook) ExternalID() string {
	// TODO: confirm this is correct
	if webhook.Data.Attributes.TxnType == "payin" {
		return webhook.Data.Attributes.ResourceID
	}
	return webhook.Data.Attributes.ResponseID
}

func NewWebhookProcessor() WebhookProcessor {
	return WebhookProcessor{}
}

type WebhookProcessor struct{}

func (processor WebhookProcessor) Process(ctx context.Context, result *requests.WebhookResult, out any) error {

	options, ok := (out).(*mpesa.OptionsUpdatePayment)
	if !ok {
		return errors.New("invalid type for options")
	}

	r := bytes.NewReader(result.Bytes())
	switch result.Action {
	case quikk2.OperationCharge:
		wb := ChargeWebhook{}
		if err := jsoniter.NewDecoder(r).Decode(&wb); err != nil {
			return err
		}
		result.Data = wb

		// check for failed status
		if wb.Meta != nil && wb.Meta.Code != quikk2.ResultCodeSuccess {
			options.Status = types.Pointer(requests.StatusFailed)
		} else {
			options.Status = types.Pointer(requests.StatusSucceeded)
			options.PaymentReference = &wb.Data.Attributes.TxnID
		}

	case quikk2.OperationPayout:
		wb := PayoutWebhook{}
		if err := jsoniter.NewDecoder(r).Decode(&wb); err != nil {
			return err
		}
		result.Data = wb

		// check for failed status
		if wb.Meta != nil && wb.Meta.Code != quikk2.ResultCodeSuccess {
			options.Status = types.Pointer(requests.StatusFailed)
		} else {
			options.Status = types.Pointer(requests.StatusSucceeded)
			options.PaymentReference = &wb.Data.Attributes.TxnID
		}

	case quikk2.OperationTransfer:
		wb := TransferWebhook{}
		if err := jsoniter.NewDecoder(r).Decode(&wb); err != nil {
			return err
		}
		result.Data = wb

		// check for failed status
		if wb.Meta != nil && wb.Meta.Code != quikk2.ResultCodeSuccess {
			options.Status = types.Pointer(requests.StatusFailed)
		} else {
			options.Status = types.Pointer(requests.StatusSucceeded)
			options.PaymentReference = &wb.Data.Attributes.TxnID
		}

	case quikk2.OperationSearch:
		// quikk.OperationSearch supports both transaction search and balance search
		// here the concern is only transaction search
		wb := TransactionSearchWebhook{}
		if err := jsoniter.NewDecoder(r).Decode(&wb); err != nil {
			return err
		}
		result.Data = wb

		// if the webhook has Meta field, safely ignore the webhook
		if wb.Meta != nil && wb.Meta.Code != quikk2.ResultCodeSuccess {
			return nil
		}

		// confirm the webhook has certain fields
		if wb.Data.Attributes.TxnType == "" {
			// safely ignore the webhook
			return nil
		}

		options.PaymentReference = &wb.Data.Attributes.TxnID
		options.Status = types.Pointer(requests.StatusSucceeded)

		return nil

	}

	return nil
}
