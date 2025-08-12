package quikk

import (
	"bytes"
	"context"
	"errors"

	jsoniter "github.com/json-iterator/go"

	"github.com/SirWaithaka/payments-api/clients/quikk"
	"github.com/SirWaithaka/payments-api/internal/domains/requests"
	"github.com/SirWaithaka/payments-api/internal/pkg/types"
)

func NewWebhookProcessor() WebhookProcessor {
	return WebhookProcessor{}
}

type WebhookProcessor struct{}

func (processor WebhookProcessor) Process(ctx context.Context, result *requests.WebhookResult, out any) error {

	options, ok := (out).(*requests.OptionsUpdatePayment)
	if !ok {
		return errors.New("invalid type for options")
	}

	r := bytes.NewReader(result.Bytes())
	switch result.Action {
	case quikk.OperationCharge:
		wb := quikk.WebhookResult[quikk.WebhookAttributesCharge]{}
		if err := jsoniter.NewDecoder(r).Decode(&wb); err != nil {
			return err
		}

		// check for failed status
		if wb.Meta != nil && wb.Meta.Code != quikk.ResultCodeSuccess {
			options.Status = types.Pointer(requests.StatusFailed)
		} else {
			options.Status = types.Pointer(requests.StatusSucceeded)
			options.PaymentReference = &wb.Data.Attributes.TxnID
		}

	case quikk.OperationPayout:
		wb := quikk.WebhookResult[quikk.WebhookAttributesPayout]{}
		if err := jsoniter.NewDecoder(r).Decode(&wb); err != nil {
			return err
		}

		// check for failed status
		if wb.Meta != nil && wb.Meta.Code != quikk.ResultCodeSuccess {
			options.Status = types.Pointer(requests.StatusFailed)
		} else {
			options.Status = types.Pointer(requests.StatusSucceeded)
			options.PaymentReference = &wb.Data.Attributes.TxnID
		}

	case quikk.OperationTransfer:
		wb := quikk.WebhookResult[quikk.WebhookAttributesTransfer]{}
		if err := jsoniter.NewDecoder(r).Decode(&wb); err != nil {
			return err
		}

		// check for failed status
		if wb.Meta != nil && wb.Meta.Code != quikk.ResultCodeSuccess {
			options.Status = types.Pointer(requests.StatusFailed)
		} else {
			options.Status = types.Pointer(requests.StatusSucceeded)
			options.PaymentReference = &wb.Data.Attributes.TxnID
		}

	case quikk.OperationSearch:
		// quikk.OperationSearch supports both transaction search and balance search
		// here the concern is only transaction search
		wb := quikk.WebhookResult[quikk.WebhookAttributesTransactionSearch]{}
		if err := jsoniter.NewDecoder(r).Decode(&wb); err != nil {
			return err
		}

		// if the webhook has Meta field, safely ignore the webhook
		if wb.Meta != nil && wb.Meta.Code != quikk.ResultCodeSuccess {
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
