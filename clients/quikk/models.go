package quikk

import (
	"fmt"
	"time"
)

// ResultCode represents the code returned from quikk in both an asynchronous
// and synchronous response. The code is added to the meta object in the response.
type ResultCode string

const (
	ResultCodeSuccess                     ResultCode = "0"
	ResultCodeInsufficientBalance         ResultCode = "1"
	ResultCodeRuleLimited                 ResultCode = "17"
	ResultCodeCancelledRequest            ResultCode = "1032"
	ResultCodeUserUnreachable             ResultCode = "1037"
	ResultCodeInvalidInitiatorInformation ResultCode = "2001"
	ResultCodeActivityTimeout             ResultCode = "500.001.1001"
	ResultCodeServiceUnavailable          ResultCode = "500.002.1001"
)

type Data[T any] struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Attributes T      `json:"attributes"`
}

// REQUEST MODELS

type RequestDefault[T any] struct {
	Data Data[T] `json:"data"`
}

// RequestTransactionStatus describes the payload to perform transaction search.
// For transactions that are older than 15 days, you can only find it when you use "txn_id"
// (MPESA transaction / receipt number)
type RequestTransactionStatus struct {
	ShortCode string `json:"short_code"`
	Reference string `json:"q"`
	// can be one of "resource_id", "response_id" or "txn_id"
	// resource_id - is the developer generated id from the original payout request
	// response_id - is the response_id from the original payout request
	// txn_id - is the MPESA receipt number
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
	RecipientType     string  `json:"recipient_type"`     // always "short_code"
	RecipientCategory string  `json:"recipient_category"` // one of "till" or "paybill"
	PostedAt          string  `json:"posted_at"`
}

// RESPONSE MODELS

// meta response can be embedded in any other type of response
type meta struct {
	Status string     `json:"status,omitempty"`
	Code   ResultCode `json:"code,omitempty"`
	Detail string     `json:"detail,omitempty"`
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

// WEBHOOKS REQUEST MODELS

type WebhookResult[T any] struct {
	Data Data[T] `json:"data"`
	Meta *meta   `json:"meta,omitempty"`
}

// WebhookAttributesPayinValidation describes the fields of the webhook payload when a direct payin
// transaction request is being made. A success response completes a transaction while a failure cancels it.
//
// # Example results
//
// SUCCESSFUL PAYLOAD
// 1. Customer payment to shortcode
//
//	{
//	 "data": {
//	   "type": "payin",
//	   "attributes": {
//	     "txn_id": "NH701HA94Y",
//	     "sender_no": "null or 2547*****567",
//	     "sender_no_sha1": "f001d87b438ff2959d6786f02d7dfa0fa2bbd17b",
//	     "sender_no_sha256": "0b75ca5ae93ccfb939dbeda34f9b5e608d8eca085efaa9ad7f40ff9d81a6acca",
//	     "sender_name": "John J D",
//	     "sender_type": "msisdn",
//	     "category": "PayBill",
//	     "amount": 1000,
//	     "txn_created_at": "2022-07-01T10:51:59+0300",
//	     "short_code": "60134",
//	     "reference": "NTSPF6038"
//	   }
//	 }
//	}
//
// 2. Business payment to shortcode
//
//	{
//	 "data": {
//	   "type": "payin",
//	   "attributes": {
//	     "txn_id": "NH87HBCW05",
//	     "sender_no": "174379",
//	     "sender_no_sha1": "7abe3618da092a099b4073edc699c4582c899303",
//	     "sender_no_sha256": "299d23f4d6223d01e42f568c8da7d85eee88feb3893c0dfbf230f3e7b1b4d2d1",
//	     "sender_name": "Safaricom",
//	     "sender_type": "short_code",
//	     "category": "B2B",
//	     "short_code": "601000",
//	     "amount": 1000,
//	     "txn_created_at": "2022-07-01T10:51:59+0300"
//	   }
//	 }
//	}
type WebhookAttributesPayinValidation struct {
	TxnID              string `json:"txn_id"`
	SenderName         string `json:"sender_name"`
	SenderNumber       string `json:"sender_no"`
	SenderNumberSha1   string `json:"sender_no_sha1"`
	SenderNumberSha256 string `json:"sender_no_sha256"`
	// SenderType will be either "msisdn" or "short_code".
	// msisdn is for customer payments to shortcode
	// short_code is for business payments to shortcode
	SenderType string  `json:"sender_type"`
	Category   string  `json:"category"`
	ShortCode  string  `json:"short_code"`
	Amount     float64 `json:"amount"`
	// Reference applies only to customer payments to shortcode
	Reference    string `json:"reference"`
	TxnCreatedAt string `json:"txn_created_at"`
}

// WebhookAttributesPayinConfirmation describes the fields of the webhook request when a direct payin
// transaction request has been made. The confirmation is an affirmation of a successful transaction
// on the MPESA system.
// Incoming transactions can be from direct customer payments to paybill shortcode, customer payments
// to till number or business payments to shortcode
//
// # Example results
//
// 1. Customer payment to shortcode
//
//	{
//	 "data": {
//	   "type": "payin",
//	   "id": "1",
//	   "attributes": {
//	     "txn_id": "NH761HA954",
//	     "sender_no": "null or 2547*****567",
//	     "sender_no_sha1": "f001d87b438ff2959d6786f02d7dfa0fa2bbd17b",
//	     "sender_no_sha256": "0b75ca5ae93ccfb939dbeda34f9b5e608d8eca085efaa9ad7f40ff9d81a6acca",
//	     "sender_name": "John J D",
//	     "sender_type": "msisdn",
//	     "category": "Paybill",
//	     "amount": 1000,
//	     "txn_created_at": "2022-07-01T10:51:59+0300",
//	     "short_code": "6012345",
//	     "reference": "NTSPF6038",
//	     "balance_utility_ac": 10,
//	     "retry": 1
//	   }
//	 }
//	}
//
// 2. Business payment to shortcode
//
//	{
//	 "data": {
//	   "type": "payin",
//	   "id": "1",
//	   "attributes": {
//	     "txn_id": "NH87HBCW05",
//	     "sender_name": "Safaricom",
//	     "sender_type": "short_code",
//	     "category": "B2B",
//	     "amount": 1000,
//	     "txn_created_at": "2022-07-01T10:51:59+0300",
//	     "short_code": "601000",
//	     "balance_utility_ac": 3060455
//	   }
//	 }
//	}
//
// 3. Customer payment to till
//
//	{
//	 "data": {
//	   "type": "payin",
//	   "attributes": {
//	     "txn_id": "NH75HBCV3B",
//	     "sender_no": "null or 2547*****567",
//	     "sender_no_sha1": "f001d87b438ff2959d6786f02d7dfa0fa2bbd17b",
//	     "sender_no_sha256": "0b75ca5ae93ccfb939dbeda34f9b5e608d8eca085efaa9ad7f40ff9d81a6acca",
//	     "sender_name": "Nicholas S",
//	     "sender_type": "msisdn",
//	     "category": "BuyGoods",
//	     "amount": 1000,
//	     "txn_created_at": "2022-07-01T10:51:59+0300",
//	     "short_code": "601340",
//	     "balance_utility_ac": 20248.75
//	   }
//	 }
//	}
type WebhookAttributesPayinConfirmation struct {
	TxnID                 string  `json:"txn_id"`
	SenderName            string  `json:"sender_name"`
	SenderNumber          string  `json:"sender_no"`
	SenderNumberSha1      string  `json:"sender_no_sha1"`
	SenderNumberSha256    string  `json:"sender_no_sha256"`
	SenderType            string  `json:"sender_type"`
	ShortCode             string  `json:"short_code"`
	Category              string  `json:"category"`
	Amount                float64 `json:"amount"`
	UtilityBalanceAccount float64 `json:"balance_utility_ac"`
	Reference             string  `json:"reference"`
	TxnCreatedAt          string  `json:"txn_created_at"`
	Retry                 int     `json:"retry"`
}

// WebhookAttributesCharge describes the fields returned from quikk from a successful charge request
// In the case of a failed charge request, all fields will be empty except for TxnChargeID
//
// # Example results
//
// SUCCESSFUL PAYLOAD
//
//	{
//	 "data": {
//	   "type": "payin",
//	   "id": "gid",
//	   "attributes": {
//	     "txn_id": "FDN34Y9809",
//	     "sender_type": "msisdn",
//	     "sender_no": "254701234567",
//	     "sender_no_sha1": "f001d87b438ff2959d6786f02d7dfa0fa2bbd17b",
//	     "sender_no_sha256": "0b75ca5ae93ccfb939dbeda34f9b5e608d8eca085efaa9ad7f40ff9d81a6acca",
//	     "amount": 500,
//	     "txn_charge_created_at": "2022-07-01T10:51:59+0300",
//	     "txn_charge_id": "ws_CO_27072017151044001"
//	   }
//	 }
//	}
//
// FAILED PAYLOAD
//
//	{
//	 "data": {
//	   "type": "payin",
//	   "id": "8555-67195-1",
//	   "attributes": {
//	     "txn_charge_id": "AG_20190809_000040b4caf4c7a029c0"
//	   }
//	 },
//	 "meta": {
//	   "status": "FAIL",
//	   "code": "1029",
//	   "detail": "[STK_CB - ]Request cancelled by user"
//	 }
//	}
type WebhookAttributesCharge struct {
	TxnChargeID        string  `json:"txn_charge_id"` // available in both success and failure cases
	Amount             float64 `json:"amount"`
	TxnID              string  `json:"txn_id"`
	TxnChargeCreatedAt string  `json:"txn_charge_created_at"`
	SenderType         string  `json:"sender_type"`
	SenderNumber       string  `json:"sender_no"`
	SenderNumberSha1   string  `json:"sender_no_sha1"`
	SenderNumberSha256 string  `json:"sender_no_sha256"`
}

// WebhookAttributesPayout describes the fields returned from quikk from a successful payout request
// In the case of a failed payout request, all fields will be empty except for TxnID and ResponseID
//
// # Example results
//
// SUCCESSFUL PAYOUT
//
//	{
//	 "data": {
//	   "type": "payout",
//	   "id": "1",
//	   "attributes": {
//	     "txn_id": "NI51HBHO4D",
//	     "response_id": "AG_20190905_00005f0dcb86732c611c",
//	     "recipient_type": "msisdn",
//	     "recipient_no": "2547*****024",
//	     "recipient_no_sha1": "a8150b7b38fa70ae1c0e4b5312f3a79827ceb657",
//	     "recipient_no_sha256": "bc89782d032e6f28356888aba50dc35a1c180078eae95465c13ab1b82b8492b9",
//	     "recipient_registered": "Y",
//	     "amount": 10,
//	     "txn_created_at": "2022-07-01T10:51:59+0300",
//	     "balance_utility_ac": 4517632.27,
//	     "balance_working_ac": 299825045,
//	     "balance_charges_paid_ac": 0
//	   }
//	 }
//	}
//
// FAILED PAYOUT
//
//	{
//	 "data": {
//	   "type": "payout",
//	   "id": "1",
//	   "attributes": {
//	     "txn_id": "NH90HBCXPM",
//	     "response_id": "AG_20190809_000040b4caf4c7a029c0"
//	   }
//	 },
//	 "meta": {
//	   "status": "FAIL",
//	   "code": "17",
//	   "detail": "The initiator is not allowed to initiate this request"
//	 }
//	}
type WebhookAttributesPayout struct {
	TxnID                     string  `json:"txn_id"`      // available in both success and failure cases
	ResponseID                string  `json:"response_id"` // available in both success and failure cases
	RecipientType             string  `json:"recipient_type"`
	RecipientNumber           string  `json:"recipient_no"`
	RecipientNumberSha1       string  `json:"recipient_no_sha1"`
	RecipientNumberSha256     string  `json:"recipient_no_sha256"`
	RecipientRegistered       string  `json:"recipient_registered"`
	Amount                    float64 `json:"amount"`
	UtilityAccountBalance     float64 `json:"balance_utility_ac"`
	WorkingAccountBalance     float64 `json:"balance_working_ac"`
	ChargesPaidAccountBalance float64 `json:"balance_charges_paid_ac"`
	TxnCreatedAt              string  `json:"txn_created_at"`
}

// WebhookAttributesTransfer describes the fields returned from quikk from successful paybill
// and till transfer requests. In the case of till transfer request, only the fields; TxnID,
// ResponseID, RecipientType and WorkingAccountBalance will be populated.
// In the case of a failed transfer request, all fields will be empty except for TxnID and ResponseID
//
// Example results
//
//	SUCCESSFUL PAYBILL TRANSFER
//
//	{
//	 "data": {
//	   "type": "transfer",
//	   "id": "1",
//	   "attributes": {
//	     "txn_id": "NH90HBCXPM",
//	     "response_id": "AG_20190809_000040b4caf4c7a029c0",
//	     "recipient_type": "short_code",
//	     "recipient_no": "12348",
//	     "recipient_no_sha1": "ad8f96d2cbd19db18de8e2cc70703ba9c3c6840b",
//	     "recipient_no_sha256": "015e81eddfab44be16ac53a8653feab50859b4c5508a915679e33c271d2b54df",
//	     "recipient_name": "some company",
//	     "short_code": "123456",
//	     "short_code_name": "thingamagik",
//	     "amount": 10,
//	     "txn_created_at": "2022-07-01T10:51:59+0300",
//	     "balance_working_ac": 111745,
//	     "balance_utility_ac": 111745,
//	     "balance_charges_paid_ac": 111745,
//	     "reference": "7000"
//	   }
//	 }
//	}
//
// SUCCESSFUL TILL TRANSFER
//
//	{
//	 "data": {
//	   "type": "transfer",
//	   "id": "1",
//	   "attributes": {
//	     "txn_id": "RCU5XJOXPM",
//	     "response_id": "AG_20230330_20205219caf4c7a029c0",
//	     "recipient_type": "short_code",
//	     "balance_working_ac": 299825
//	   }
//	 }
//	}
//
// FAILED TRANSFER
//
//	{
//	 "data": {
//	   "type": "transfer",
//	   "id": "1",
//	   "attributes": {
//	     "txn_id": "NH90HBCXPM",
//	     "response_id": "AG_20190809_000040b4caf4c7a029c0"
//	   }
//	 },
//	 "meta": {
//	   "status": "FAIL",
//	   "code": "20",
//	   "detail": "Insufficient balance"
//	 }
//	}
type WebhookAttributesTransfer struct {
	TxnID                     string  `json:"txn_id"`
	ResponseID                string  `json:"response_id"`
	Reference                 string  `json:"reference"`
	RecipientType             string  `json:"recipient_type"`
	RecipientName             string  `json:"recipient_name"`
	RecipientNumber           string  `json:"recipient_no"`
	RecipientNumberSha1       string  `json:"recipient_no_sha1"`
	RecipientNumberSha256     string  `json:"recipient_no_sha256"`
	ShortCode                 string  `json:"short_code"`
	ShortCodeName             string  `json:"short_code_name"`
	Amount                    float64 `json:"amount"`
	WorkingAccountBalance     float64 `json:"balance_working_ac"`
	UtilityAccountBalance     float64 `json:"balance_utility_ac"`
	ChargesPaidAccountBalance float64 `json:"balance_charges_paid_ac"`
	TxnCreatedAt              string  `json:"txn_created_at"`
}

// WebhookAttributesRefund describes the fields returned from quikk from a successful refund request.
// In the case of a failed refund request, all fields will be empty except for TxnID and ResponseID.
//
// Example results
//
//	SUCCESSFUL PAYLOAD
//	{
//	 "data": {
//	   "type": "refund",
//	   "id": "1",
//	   "attributes": {
//	     "txn_id": "NH94HBCXII",
//	     "sender_no": "2547*****024",
//	     "sender_no_sha1": "a8150b7b38fa70ae1c0e4b5312f3a79827ceb657",
//	     "sender_no_sha256": "bc89782d032e6f28356888aba50dc35a1c180078eae95465c13ab1b82b8492b9",
//	     "sender_name": "Jane J D",
//	     "amount": 450,
//	     "txn_created_at": "2022-07-01T10:51:59+0300",
//	     "short_code": "511382",
//	     "shortcode_name": "Safaricom Disbursement Account",
//	     "origin_txn_id": "NH90HBCXI4",
//	     "response_id": "AG_20190808_000051f18a81f3aee279",
//	     "balance_utility_ac": 4761531.1,
//	     "txn_charge": 0
//	   }
//	 }
//	}
//
// FAILED
//
//	{
//	 "data": {
//	   "type": "refund",
//	   "id": "be140604-b53d-4725-8cdb-60b946286536",
//	   "attributes": {
//	     "txn_id": "NH90000000",
//	     "response_id": "AG_20190809_00005572cb7edaef9ce6"
//	   }
//	 },
//	 "meta": {
//	   "status": "FAIL",
//	   "code": "R000001",
//	   "detail": "The transaction has already been reversed."
//	 }
//	}
type WebhookAttributesRefund struct {
	TxnID                 string  `json:"txn_id"`
	OriginTxnID           string  `json:"origin_txn_id"`
	ResponseID            string  `json:"response_id"`
	SenderName            string  `json:"sender_name"`
	SenderNumber          string  `json:"sender_no"`
	SenderNumberSha1      string  `json:"sender_no_sha1"`
	SenderNumberSha256    string  `json:"sender_no_sha256"`
	ShortCode             string  `json:"short_code"`
	ShortcodeName         string  `json:"shortcode_name"`
	Amount                float64 `json:"amount"`
	TxnCharge             float64 `json:"txn_charge"`
	UtilityAccountBalance float64 `json:"balance_utility_ac"`
	TxnCreatedAt          string  `json:"txn_created_at"`
}

// WebhookAttributesTransactionSearch describes the fields returned from quikk from a successful
// transaction search request.
//
// # Example results
//
//	{
//	 "data": {
//	   "type": "search",
//	   "id": "6e9a2aad-621e-444c-9fc8-1e08fdaaaa6c",
//	   "attributes": {
//	     "resource_id": "1",
//	     "response_id": "AG_20190417_000049de14ae0c48",
//	     "txn_id": "JK34DSL0UW",
//	     "amount": 1222.22,
//	     "recipient_fee": 33,
//	     "recipient_name": "Safaricom Disbursement Account",
//	     "recipient_no": "511382",
//	     "recipient_type": "short_code",
//	     "sender_name": "Jane J D",
//	     "sender_no": "2547*****678",
//	     "sender_no_sha1": "90fbb07a296a87f476af720ccb89f35822e50182",
//	     "sender_no_sha256": "7132104d6aae9c3fac82095a42c2817952bca48e09d98d5bf4ac08218982fb90",
//	     "sender_fee": "33.0,",
//	     "txn_type": "payin",
//	     "category": "Paybill",
//	     "txn_status": "Authorized",
//	     "txn_created_at": "2022-07-01T10:51:59+0300"
//	   }
//	 }
//	}
type WebhookAttributesTransactionSearch struct {
	TxnID              string  `json:"txn_id"`
	ResourceID         string  `json:"resource_id"`
	ResponseID         string  `json:"response_id"`
	Amount             float64 `json:"amount"`
	RecipientFee       float64 `json:"recipient_fee"`
	RecipientName      string  `json:"recipient_name"`
	RecipientNumber    string  `json:"recipient_no"`
	RecipientType      string  `json:"recipient_type"`
	SenderName         string  `json:"sender_name"`
	SenderNumber       string  `json:"sender_no"`
	SenderNumberSha1   string  `json:"sender_no_sha1"`
	SenderNumberSha256 string  `json:"sender_no_sha256"`
	SenderFee          string  `json:"sender_fee"`
	Category           string  `json:"category"`
	TxnType            string  `json:"txn_type"`
	TxnStatus          string  `json:"txn_status"`
	TxnCreatedAt       string  `json:"txn_created_at"`
}

// WebhookAttributesBalanceSearch describes the fields returned from quikk from a successful
// balance search request.
//
// # Example results
//
//	{
//	 "data": {
//	   "type": "search",
//	   "id": "1",
//	   "attributes": {
//	     "response_id": "AG_20190808_000051f18a81f3aee279",
//	     "txn_id": "NH94HBCXII",
//	     "balance_working_ac": 4761531.1,
//	     "balance_utility_ac": 4761531,
//	     "balance_charges_paid_ac": 4761531,
//	     "balance_merchant_ac": 4761531,
//	     "balance_organization_settlement_ac": 4761531,
//	     "checked_at": "2019-03-18T17:22:09.651011Z"
//	   }
//	 }
//	}
type WebhookAttributesBalanceSearch struct {
	TxnID                       string    `json:"txn_id"`
	ResponseID                  string    `json:"response_id"`
	WorkingAccountBalance       float64   `json:"balance_working_ac"`
	UtilityAccountBalance       float64   `json:"balance_utility_ac"`
	MerchantAccountBalance      float64   `json:"balance_merchant_ac"`
	ChargesPaidAccountBalance   float64   `json:"balance_charges_paid_ac"`
	OrgSettlementAccountBalance float64   `json:"balance_organization_settlement_ac"`
	CheckedAt                   time.Time `json:"checked_at"`
}
