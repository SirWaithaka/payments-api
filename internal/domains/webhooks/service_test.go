package webhooks_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/go-playground/assert/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/oklog/ulid/v2"

	"github.com/SirWaithaka/payments-api/internal/domains/mpesa"
	"github.com/SirWaithaka/payments-api/internal/domains/requests"
	"github.com/SirWaithaka/payments-api/internal/domains/webhooks"
	"github.com/SirWaithaka/payments-api/internal/pkg/events"
	"github.com/SirWaithaka/payments-api/internal/repositories/postgres"
	"github.com/SirWaithaka/payments-api/internal/testdata"
)

type FakeWebhookBody struct {
	ResultCode    string `json:"ResultCode"`
	OriginationID string `json:"OriginationID"`
	Amount        string `json:"Amount"`
	ReceiptID     string `json:"ReceiptID"`
}

func (f FakeWebhookBody) ExternalID() string {
	return f.OriginationID
}

type MockPublisher struct{}

func (m *MockPublisher) Publish(ctx context.Context, event events.EventType) error {
	return nil
}

type MockWebhookProcessor struct{}

func (m MockWebhookProcessor) Process(ctx context.Context, result *requests.WebhookResult, out any) error {
	var body FakeWebhookBody
	if err := jsoniter.Unmarshal(result.Bytes(), &body); err != nil {
		return err
	}

	result.Data = body

	opts := out.(*mpesa.OptionsUpdatePayment)

	var status requests.Status
	if body.ResultCode == "0" {
		status = requests.StatusSucceeded
	} else {
		status = requests.StatusFailed
	}

	// update options
	opts.PaymentReference = &body.ReceiptID
	opts.Status = &status

	return nil
}

type MockProvider struct{}

func (m MockProvider) GetWebhookClient(service requests.Partner) requests.WebhookProcessor {
	return &MockWebhookProcessor{}
}

func TestWebhookService_Process(t *testing.T) {
	defer testdata.ResetTables(inf)

	requestsRepo := postgres.NewRequestRepository(inf.Storage.PG)
	paymentsRepo := postgres.NewMpesaPaymentsRepository(inf.Storage.PG)
	webhooksRepo := postgres.NewWebhookRepository(inf.Storage.PG)

	// save a payment
	payment := mpesa.Payment{
		PaymentID:           ulid.Make().String(),
		ClientTransactionID: ulid.Make().String(),
		IdempotencyID:       ulid.Make().String(),
		Status:              "received",
	}
	err := paymentsRepo.Add(t.Context(), payment)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}

	// save a request record
	request := requests.Request{
		RequestID:  ulid.Make().String(),
		PaymentID:  payment.PaymentID,
		ExternalID: ulid.Make().String(),
		Partner:    "test",
		Status:     "completed",
	}
	err = requestsRepo.Add(t.Context(), request)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}

	service := webhooks.NewService(webhooksRepo, requestsRepo, paymentsRepo, &MockProvider{}, &MockPublisher{})

	// fake webhook result
	body := `{"ResultCode": "%s","OriginationID": "%s","Amount": "100","ReceiptID": "%s"}`

	// use a fake payment reference, use request external id in the fake webhook result
	paymentReference := ulid.Make().String()
	fakeWebhook := requests.NewWebhookResult("test", "express", strings.NewReader(fmt.Sprintf(body, "0", request.ExternalID, paymentReference)))
	err = service.Process(t.Context(), fakeWebhook)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}

	// fetch payment and assert its values
	var record postgres.MpesaPaymentSchema
	result := inf.Storage.PG.Where(postgres.MpesaPaymentSchema{PaymentID: payment.PaymentID}).First(&record)
	if result.Error != nil {
		t.Errorf("expected nil error, got %v", result.Error)
	}

	assert.Equal(t, paymentReference, record.PaymentReference)
	assert.Equal(t, requests.StatusSucceeded.String(), record.Status)
}
