package quikk

import "fmt"

type Data struct {
	ID         string      `json:"id"`
	Type       string      `json:"type"`
	Attributes interface{} `json:"attributes"`
}

type RequestDefault struct {
	Data Data `json:"data"`
}

type RequestTransactionStatus struct {
	ShortCode     string `json:"short_code"`
	Reference     string `json:"q"`
	ReferenceType string `json:"on"`
}

type RequestAccountBalance struct {
	ShortCode string `json:"short_code"`
}

type RequestCharge struct {
	Amount       float64 `json:"amount"`
	CustomerNo   string  `json:"customer_no"`
	Reference    string  `json:"reference"`
	CustomerType string  `json:"customer_type"`
	ShortCode    string  `json:"short_code"`
	PostedAt     string  `json:"posted_at"` // ISO string
}

type RequestPayout struct {
	Amount        float64 `json:"amount"`
	RecipientNo   string  `json:"recipient_no"`
	RecipientType string  `json:"recipient_type"`
	ShortCode     string  `json:"short_code"`
	PostedAt      string  `json:"posted_at"`
}

type RequestTransfer struct {
	Amount            float64 `json:"amount"`
	RecipientNo       string  `json:"recipient_no"`
	AccountNo         string  `json:"reference"`
	ShortCode         string  `json:"short_code"`
	RecipientType     string  `json:"recipient_type"`
	RecipientCategory string  `json:"recipient_category"`
	PostedAt          string  `json:"posted_at"`
}

// Response models

// meta response can be embedded in any other type of response
type meta struct {
	Status string `json:"status,omitempty"`
	Code   string `json:"code,omitempty"`
	Detail string `json:"detail,omitempty"`
}

func (meta meta) Error() string {
	if meta.Status != "FAIL" {
		return ""
	}
	return fmt.Sprintf("<%v: %v> - %v", meta.Status, meta.Code, meta.Detail)
}

// ResponseDefault Response common to all/some api calls
type ResponseDefault struct {
	Data *struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			ResourceID string `json:"resource_id"`
		} `json:"attributes"`
	} `json:"data,omitempty"`
	Meta *meta `json:"meta,omitempty"`
}

type ErrorResponse struct {
	Errors []struct {
		Status string `json:"status"`
		Title  string `json:"title"`
		Detail string `json:"detail"`
	} `json:"errors"`
}
