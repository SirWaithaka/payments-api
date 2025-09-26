package requests

import (
	"bytes"
	"context"
	"io"
	"time"
)

// Status type describes the lifecycle of a payment request or api request
type Status string

const (
	StatusReceived  Status = "received"
	StatusSent      Status = "sent"
	StatusSucceeded Status = "succeeded"
	StatusFailed    Status = "failed"
	StatusError     Status = "error"
	StatusTimeout   Status = "timeout"
	StatusDeclined  Status = "declined"
)

func (s Status) String() string {
	return string(s)
}

// Final returns true if status equals to a final state of request
func (s Status) Final() bool {
	return s == StatusSucceeded || s == StatusFailed || s == StatusDeclined || s == StatusError
}

func ToStatus(s string) Status {
	switch s {
	case string(StatusReceived):
		return StatusReceived
	case string(StatusSent):
		return StatusSent
	case string(StatusSucceeded):
		return StatusSucceeded
	case string(StatusFailed):
		return StatusFailed
	case string(StatusError):
		return StatusError
	case string(StatusTimeout):
		return StatusTimeout
	case string(StatusDeclined):
		return StatusDeclined
	default:
		return "unknown"
	}
}

//go:generate stringer -type=Partner -linecomment -output=request_string.go

// Partner is an enum for all external partner apis we support
type Partner int

const (
	PartnerUnknown Partner = iota // unknown
	PartnerDaraja                 // daraja
	PartnerQuikk                  // quikk
	PartnerTanda                  // tanda
)

func (partner Partner) MarshalText() ([]byte, error) {
	return []byte(partner.String()), nil
}

func ToPartner(partner string) Partner {
	switch partner {
	case PartnerDaraja.String():
		return PartnerDaraja
	case PartnerTanda.String():
		return PartnerTanda
	case PartnerQuikk.String():
		return PartnerQuikk
	default:
		return PartnerUnknown
	}
}

type Request struct {
	RequestID  string // unique request id
	PaymentID  string // foreign id tied to the original payment request
	ExternalID string // request id we get back from partner from response
	Partner    string
	Status     Status
	Latency    time.Duration
	Response   map[string]any
	CreatedAt  time.Time
}

// OptionsFindRequest defines all options that can be used to find
// one or many requests.
type OptionsFindRequest struct {
	RequestID  *string // use to find a single request
	ExternalID *string // use to find a single request
	PaymentID  *string // use to find multiple requests
}

type OptionsUpdateRequest struct {
	ExternalID *string
	Status     *Status
	Latency    *time.Duration
	Response   map[string]any
}

type WebhookResult struct {
	Service Partner
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

func NewWebhookResult(partner string, action string, body io.Reader) *WebhookResult {
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(body); err != nil {
		buf.Write([]byte{})
	}

	return &WebhookResult{Service: ToPartner(partner), Action: action, body: buf.Bytes()}
}

// Repository defines methods for managing and interacting with Request entities.
type Repository interface {
	Add(ctx context.Context, req Request) error
	FindOne(ctx context.Context, opts OptionsFindRequest) (Request, error)
	UpdateRequest(ctx context.Context, id string, opts OptionsUpdateRequest) error
}

type WebhookProcessor interface {
	Process(ctx context.Context, in *WebhookResult, out any) error
}
