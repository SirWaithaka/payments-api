package requests

import (
	"bytes"
	"context"
	"io"
	"time"

	jsoniter "github.com/json-iterator/go"
)

type Status string

const (
	StatusReceived  Status = "received"
	StatusSucceeded Status = "succeeded"
	StatusFailed    Status = "failed"
	StatusDeclined  Status = "declined"
)

func (s Status) String() string {
	return string(s)
}

type Payment struct {
	BankCode string
	// automatically generated unique id
	PaymentID string
	// client's transaction reference
	ClientTransactionID string
	// client generated unique id identifying the payment request
	IdempotencyID string
	// payment reference from the payment processor
	PaymentReference string
	// source account
	SourceAccountNumber string
	// destination account
	DestinationAccountNumber string
	// (optional) - this can be an account number, account name
	// of the beneficiary for the payment
	Beneficiary string
	// amount for payment
	Amount string
	// short description for payment
	Description string
	// status of the payment, "received", "pending", "completed", "failed", "refunded"
	Status string // TODO: Use enum type
}

type OptionsFindOnePayment struct {
	PaymentID           *string
	ClientTransactionID *string
	PaymentReference    *string
	IdempotencyID       *string
}

type OptionsUpdatePayment struct {
	Status           *Status
	PaymentReference *string
}

type Request struct {
	RequestID  string // unique request id
	PaymentID  string // foreign id tied to the original payment request
	ExternalID string // request id we get back from partner from response
	Partner    string
	Status     string
	Latency    time.Duration
	Response   map[string]any
	CreatedAt  time.Time

	Payment *Payment
}

type OptionsFindOneRequest struct {
	RequestID  *string
	ExternalID *string
}

type OptionsUpdateRequest struct {
	ExternalID *string
	Status     *string
	Response   map[string]any
}

type WebhookResult struct {
	Service string
	Action  string
	Body    io.Reader
	Data    any

	body []byte
}

// Bytes converts the Body io.Reader into a byte slice.
// Should probably avoid calling this method.
func (result WebhookResult) Bytes() []byte {
	buf := make([]byte, len(result.body))
	copy(buf, result.body)
	return buf
}

func (result WebhookResult) Reader() io.Reader {
	return bytes.NewReader(result.Bytes())
}

func (result WebhookResult) MarshalJSON() ([]byte, error) {
	r := struct {
		Action  string
		Body    io.Reader
		Service string
	}{
		Action:  result.Action,
		Body:    result.Reader(),
		Service: result.Service,
	}

	return jsoniter.Marshal(r)
}

//func (result *WebhookResult) UnmarshalJSON(data []byte) error {
//	r := struct {
//		Action  string
//		Body    any
//		Service string
//	}{}
//	if err := jsoniter.Unmarshal(data, &r); err != nil {
//		return err
//	}
//	result.Action = r.Action
//	result.Service = r.Service
//	result.Body = r.Body
//	return nil
//}

func NewWebhookResult(partner string, action string, body io.Reader) *WebhookResult {
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(body); err != nil {
		buf.Write([]byte{})
	}

	return &WebhookResult{Service: partner, Action: action, body: buf.Bytes()}
}

// Repository defines methods for managing and interacting with Request entities.
type Repository interface {
	Add(ctx context.Context, req Request) error
	FindOneRequest(ctx context.Context, opts OptionsFindOneRequest) (Request, error)
	UpdateRequest(ctx context.Context, id string, opts OptionsUpdateRequest) error
}

type WebhookProcessor interface {
	Process(ctx context.Context, result *WebhookResult) (OptionsUpdatePayment, error)
}

type Provider interface {
	GetWebhookClient(service string) WebhookProcessor
}
