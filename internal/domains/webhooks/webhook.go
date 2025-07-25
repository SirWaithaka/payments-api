package webhooks

import (
	"bytes"
	"context"
	"io"
	"time"
)

//func NewReader(reader io.Reader) *Reader {
//
//}
//
//type Reader struct {
//	buf *bytes.Buffer
//}

type WebhookRequest struct {
	ID        string
	Action    string
	Partner   string
	Payload   io.Reader
	CreatedAt time.Time
}

func NewWebhookResult(partner string, action string, body io.Reader) *WebhookResult {
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(body); err != nil {
		buf.Write([]byte{})
	}

	return &WebhookResult{Service: partner, Action: action, body: buf.Bytes()}
}

type WebhookResult struct {
	Service string
	Action  string
	body    []byte
	Data    any
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

type WebhookRepository interface {
	Add(ctx context.Context, partner, action string, payload []byte) error
	Find(ctx context.Context, id string) (WebhookRequest, error)
}

type WebhookProcessor interface {
	Process(ctx context.Context, result *WebhookResult) error
}

type Provider interface {
	GetWebhookClient(service string) WebhookProcessor
}
