package daraja

import (
	"encoding/base64"
	"time"

	jsoniter "github.com/json-iterator/go"
)

// ENUMS

// ResultCode represents the asynchronous result notification from the Daraja API
type ResultCode int32

const (
	ResultCodeSuccess                     ResultCode = 0
	ResultCodeInsufficientBalance         ResultCode = 1
	ResultCodeSubscriberLockFailed        ResultCode = 1001
	ResultCodeTransactionExpired          ResultCode = 1019
	ResultCodePushRequestFailed           ResultCode = 1025
	ResultCodeCancelledRequest            ResultCode = 1032
	ResultCodeUserUnreachable             ResultCode = 1037
	ResultCodeInvalidInitiatorInformation ResultCode = 2001
	ResultCodeLockedSecurityCredential    ResultCode = 8006
	ResultCodeInternalError               ResultCode = 9999
)

// ReplyCode is used to reply to a validation request callback to either reject or accept a transaction
type ReplyCode string

const (
	ReplyCodeC2B00011 ReplyCode = "C2B00011" // C2B00011 - Invalid MSISDN
	ReplyCodeC2B00012 ReplyCode = "C2B00012" // C2B00012 - Invalid Account Number
	ReplyCodeC2B00013 ReplyCode = "C2B00013" // C2B00013 - Invalid Amount
	ReplyCodeC2B00014 ReplyCode = "C2B00014" // C2B00014 - Invalid KYC Details
	ReplyCodeC2B00015 ReplyCode = "C2B00015" // C2B00015 - Invalid Shortcode
	ReplyCodeC2B00016 ReplyCode = "C2B00016" // C2B00016 - Other error
)

type Command string

const (
	CommandAccountBalance      Command = "AccountBalance"
	CommandSalaryPayment       Command = "SalaryPayment"
	CommandBusinessPayment     Command = "BusinessPayment"
	CommandBusinessPayBill     Command = "BusinessPayBill"
	CommandBusinessBuyGoods    Command = "BusinessBuyGoods"
	CommandPromotionPayment    Command = "PromotionPayment"
	CommandTransactionReversal Command = "TransactionReversal"
	CommandTransactionStatus   Command = "TransactionStatusQuery"
)

type IdentifierType string

const (
	IdentifierMSISDN              IdentifierType = "1"  //MSISDN
	IdentifierTillNumber          IdentifierType = "2"  //TillNumber
	IdentifierSPShortCode         IdentifierType = "3"  //SPShortCode
	IdentifierOrgShortCode        IdentifierType = "4"  //OrganizationShortCode
	IdentifierIdentityID          IdentifierType = "5"  //IdentityID
	IdentifierO2CLink             IdentifierType = "6"  //O2CLink
	IdentifierSPOperatorCode      IdentifierType = "9"  //SPOperatorCode
	IdentifierPOSNumber           IdentifierType = "10" //POSNumber
	IdentifierOrgOperatorUsername IdentifierType = "11" //OrganizationOperatorUserName
	IdentifierOrgOperatorCode     IdentifierType = "12" //OrganizationOperatorCode
	IdentifierVoucherCode         IdentifierType = "13" //VoucherCode
)

type TransactionType string

const (
	TypeCustomerPayBillOnline  TransactionType = "CustomerPayBillOnline"
	TypeCustomerBuyGoodsOnline TransactionType = "CustomerBuyGoodsOnline"
)

//go:generate stringer -type=ResponseCode -linecomment -output=models_string.go

// ResponseCode represents a synchronous error notification gotten from the Daraja API
type ResponseCode int

const UnknownResponseCode ResponseCode = -1 // 9999

const (
	// SuccessSubmission 0 means successful submission
	SuccessSubmission ResponseCode = iota + 1 // 0
	//InvalidAccountReference means an invalid value was passed for AccountReference field in transaction
	InvalidAccountReference // 23
	// EmptyAccountReference means no value was passed for the AccountReference field in transaction
	EmptyAccountReference // 1005
	// CheckSuccess means an org query request was successful
	CheckSuccess // 4000
	// InvalidReceiverIdentifierType means the provided IdentifierType is wrong.
	// Can also mean invalid remarks
	InvalidReceiverIdentifierType // 400.002.02
	// InvalidAccessToken means you might be using a wrong or expired access token
	InvalidAccessToken // 400.003.01
	// BadRequest means the server cannot process the request because something is missing
	BadRequest // 400.003.02
	// InvalidRequestPayload Your request body is not properly drafted
	InvalidRequestPayload // 400.002.05
	// InvalidGrantType means Invalid grant type passed. Select grant_type as client_credentials
	InvalidGrantType // 400.008.02
	// InvalidAuthType means Invalid Authentication passed. Select type as Basic Auth
	InvalidAuthType // 400.008.01
	// InvalidAuthHeader If you’ve possibly misplaced the headers, you will get the error
	InvalidAuthHeader // 404.001.04
	// ResourceNotFound means the requested resource could not be found but may be available in the future. Subsequent requests by the client are permissible
	ResourceNotFound // 404.003.01
	// SubscriberLock means another similar transaction is already in progress for the current subscriber
	SubscriberLock // 500.001.1001
	// ServiceTemporarilyUnavailable means intermittent service. Check transaction status
	ServiceTemporarilyUnavailable // 500.002.1001
	// SpikeArrestViolation error means your endpoints constantly generate a lot of errors that lead to a spike that affects our M-PESA performance
	SpikeArrestViolation // 500.003.02
	// QuotaViolation error means you are sending multiple requests that violate M-PESA transaction per second speed
	QuotaViolation // 500.003.03
	// InternalServerError means Server failure
	InternalServerError // 500.003.1001
)

func ToResponseCode(code string) ResponseCode {
	switch code {
	case SuccessSubmission.String():
		return SuccessSubmission
	case InvalidAccountReference.String():
		return InvalidAccountReference
	case EmptyAccountReference.String():
		return EmptyAccountReference
	case CheckSuccess.String():
		return CheckSuccess
	case InvalidReceiverIdentifierType.String():
		return InvalidReceiverIdentifierType
	case InvalidAccessToken.String():
		return InvalidAccessToken
	case BadRequest.String():
		return BadRequest
	case InvalidRequestPayload.String():
		return InvalidRequestPayload
	case InvalidGrantType.String():
		return InvalidGrantType
	case InvalidAuthType.String():
		return InvalidAuthType
	case InvalidAuthHeader.String():
		return InvalidAuthHeader
	case ResourceNotFound.String():
		return ResourceNotFound
	case SubscriberLock.String():
		return SubscriberLock
	case ServiceTemporarilyUnavailable.String():
		return ServiceTemporarilyUnavailable
	case SpikeArrestViolation.String():
		return SpikeArrestViolation
	case QuotaViolation.String():
		return QuotaViolation
	case InternalServerError.String():
		return InternalServerError
	default:
		return UnknownResponseCode
	}
}

func (code ResponseCode) MarshalJSON() ([]byte, error) {
	return jsoniter.Marshal(code.String())
}

func (code ResponseCode) MarshalText() ([]byte, error) {
	return []byte(code.String()), nil
}

func (code *ResponseCode) UnmarshalJSON(b []byte) error {
	var s string
	if err := jsoniter.Unmarshal(b, &s); err != nil {
		return err
	}
	*code = ToResponseCode(s)
	return nil
}

func (code *ResponseCode) UnmarshalText(text []byte) error {
	*code = ToResponseCode(string(text))
	return nil
}

// TYPES

const (
	timeFormat = "20060102150405"
)

// Timestamp type represents time in the format of YYYYMMDDHHmmss
type Timestamp struct {
	t time.Time
}

func NewTimestamp() Timestamp {
	return Timestamp{t: time.Now()}
}

func (t Timestamp) String() string {
	return t.t.Format(timeFormat)
}

func (t Timestamp) MarshalJSON() ([]byte, error) {
	return jsoniter.Marshal(t.String())
}

func (t *Timestamp) UnmarshalJSON(b []byte) error {
	var (
		s   string
		err error
	)

	if err = jsoniter.Unmarshal(b, &s); err != nil {
		return err
	}
	t.t, err = time.Parse(timeFormat, s)
	return err
}

// Password represents the daraja platform password format, which is a
// concatenation of shortcode, passphrase and the current timestamp
type Password struct {
	shortCode  string
	passphrase string
	timestamp  Timestamp
}

func NewPassword(shortcode, passphrase string, timestamp Timestamp) Password {
	return Password{shortCode: shortcode, passphrase: passphrase, timestamp: timestamp}
}

func (p Password) Encode() string {
	return base64.StdEncoding.EncodeToString([]byte(p.shortCode + p.passphrase + p.timestamp.String()))
}

// ERROR RESPONSE MODEL

// ErrorResponse represents the global structure of daraja errors
type ErrorResponse struct {
	//This is a unique requestID for the payment request
	RequestID string `json:"requestId"`

	//This is a predefined code that indicates the reason for a request failure.
	//This is defined in the Response Error Details below.
	//The error codes map to specific error messages
	ErrorCode ResponseCode `json:"errorCode"`

	//This is a short descriptive message of the failure reason
	ErrorMessage string `json:"errorMessage"`
}

// REQUEST MODELS

type RequestC2BExpress struct {
	//This is the organization's shortcode (Paybill or Buygoods)
	//used to identify an organization and receive the transaction
	BusinessShortCode string `json:"BusinessShortCode"`

	//This is the password used for encrypting the request sent: A base64 encoded string.
	//The base64 string is a combination of Shortcode+Passkey+Timestamp
	Password string `json:"Password"`

	//Timestamp of the transaction, normally in the format of YYYYMMDDHHMMSS
	Timestamp Timestamp `json:"Timestamp"`

	//This is the transaction type that is used to identify the transaction when
	//sending the request to M-PESA. The transaction type for M-PESA Express is
	//"CustomerPayBillOnline" for PayBill Numbers and "CustomerBuyGoodsOnline" for Till Numbers
	TransactionType TransactionType `json:"TransactionType"`

	//This is the amount transacted normally a numeric value. Money that the
	//customer pays to the Shortcode. Only whole numbers are supported
	Amount string `json:"Amount"`

	//The phone number sending money. The parameter expected is a Valid
	//Safaricom Mobile Number that is M-PESA registered in the format 2547XXXXXXXX
	PartyA string `json:"PartyA"`

	//The organization that receives the funds. The parameter expected is a 5 to 6-digit
	//as defined in the Shortcode description above.
	//This can be the same as the BusinessShortCode value above
	PartyB string `json:"PartyB"`

	//The Mobile Number to receive the STK Pin Prompt.
	//This number can be the same as PartyA value above
	PhoneNumber string `json:"PhoneNumber"`

	//A CallBack URL is a valid secure URL that is used to receive notifications from M-Pesa API.
	//It is the endpoint to which the results will be sent by M-Pesa API
	CallBackURL string `json:"CallBackURL"`

	//Account Reference: This is an Alpha-Numeric parameter that is defined by your system
	//as an Identifier of the transaction for the CustomerPayBillOnline transaction type.
	//Along with the business name, this value is also displayed to the customer in the
	//STK Pin Prompt message. Maximum of 12 characters
	AccountReference string `json:"AccountReference"`

	//This is any additional information/comment that can be sent along with the request
	//from your system. Maximum of 13 Characters
	TransactionDesc string `json:"TransactionDesc"`
}

type RequestC2BExpressQuery struct {
	//This is the organization's shortcode (Paybill or Buygoods)
	//used to identify an organization and receive the transaction
	BusinessShortCode string `json:"BusinessShortCode"`

	//This is the password used for encrypting the request sent: A base64 encoded string.
	//The base64 string is a combination of Shortcode+Passkey+Timestamp
	Password string `json:"Password"`

	//Timestamp of the transaction, normally in the format of YYYYMMDDHHMMSS
	Timestamp string `json:"Timestamp"`

	//This is a global unique identifier of the processed checkout transaction request
	CheckoutRequestID string `json:"CheckoutRequestID"`
}

type RequestReversal struct {
	//This is the credential/username used to authenticate the transaction request
	Initiator string `json:"Initiator"`

	//This is the value obtained after encrypting the API initiator password
	SecurityCredential string `json:"SecurityCredential"`

	//Takes only the 'TransactionReversal' Command id
	CommandID Command `json:"CommandID"`

	//M-Pesa transaction ID of the transaction that is being reversed
	TransactionID string `json:"TransactionID"`

	//Amount of the transaction
	Amount string `json:"Amount,omitempty"`

	//The organization that receives the transaction
	ReceiverParty string `json:"ReceiverParty"`

	//Type of organization that receives the transaction
	ReceiverIdentifierType IdentifierType `json:"RecieverIdentifierType"`

	//The path that stores information about the transaction
	ResultURL string `json:"ResultURL"`

	//The path that stores information about the time-out transaction
	QueueTimeOutURL string `json:"QueueTimeOutURL"`

	//Comments that are sent along with the transaction
	Remarks string `json:"Remarks"`

	//Any additional information to be associated with the transaction
	Occasion string `json:"Occasion,omitempty"`
}

type RequestB2C struct {
	//This is a unique string you specify for every API request you simulate
	OriginatorConversationID string `json:"OriginatorConversationID"`

	//This is an API user created by the Business Administrator of the M-PESA Bulk
	//disbursement account that is active and authorized to initiate B2C transactions via API
	InitiatorName string `json:"InitiatorName"`

	//This is the value obtained after encrypting the API initiator password
	SecurityCredential string `json:"SecurityCredential"`

	//This is a unique command that specifies the B2C transaction type
	CommandID Command `json:"CommandID"`

	//The amount of money being sent to the customer
	Amount string `json:"Amount"`

	//This is the B2C organization shortcode from which the money is sent from
	PartyA string `json:"PartyA"`

	//This is the customer mobile number to receive the amount.
	//The number should have the country code (254) without the plus sign
	PartyB string `json:"PartyB"`

	//Any additional information to be associated with the transaction.
	//String up to 100 characters
	Remarks string `json:"Remarks"`

	//This is the URL to be specified in your request that will be used by API Proxy
	//to send notification incase the payment request is timed out while awaiting processing in the queue
	QueueTimeOutURL string `json:"QueueTimeOutURL"`

	//This is the URL to be specified in your request that will be used by M-PESA to
	//send notification upon processing of the payment request
	ResultURL string `json:"ResultURL"`

	//Any additional information to be associated with the transaction
	Occasion string `json:"Occasion"`
}

type RequestB2B struct {
	//This is the credential/username used to authenticate the request
	Initiator string `json:"Initiator"`

	//The encrypted password of the M-Pesa API operator
	SecurityCredential string `json:"SecurityCredential"`

	//Use 'BusinessPayBill' Command for paybill. Use 'BusinessBuyGoods' for till number.
	CommandID Command `json:"CommandID"`

	//The type of shortcode from which money is deducted. For this request, only "4" is allowed
	SenderIdentifierType IdentifierType `json:"SenderIdentifierType"`

	//The type of shortcode to which money is credited. This API supports type 4 only
	RecieverIdentifierType IdentifierType `json:"RecieverIdentifierType"`

	//The transaction amount
	Amount string `json:"Amount"`

	//Your shortcode. The shortcode from which money will be deducted
	PartyA string `json:"PartyA"`

	//The shortcode to which money will be moved
	PartyB string `json:"PartyB"`

	//The account number to be associated with the payment
	AccountReference string `json:"AccountReference"`

	//The consumer’s mobile number on behalf of whom you are paying (optional)
	Requester *string `json:"Requester,omitempty"`

	//Any additional information to be associated with the transaction
	Remarks string `json:"Remarks"`

	//A URL that will be used to notify your system in case the request times out before processing
	QueueTimeOutURL string `json:"QueueTimeOutURL"`

	//A URL that will be used to send transaction results after processing
	ResultURL string `json:"ResultURL"`
}

type RequestTransactionStatus struct {
	//This is the credential/username used to authenticate the request
	Initiator string `json:"Initiator"`

	//Encrypted credential of the user getting transaction status
	SecurityCredential string `json:"SecurityCredential"`

	//Takes only the 'TransactionStatusQuery' Command ID
	CommandID Command `json:"CommandID"`

	//Unique identifier to identify a transaction on Mpesa
	TransactionID *string `json:"TransactionID,omitempty"`

	//This is a global unique identifier for the transaction request returned by the API proxy upon
	//successful request submission. If you don’t have the M-PESA transaction ID, you can use this to query
	OriginatorConversationID *string `json:"OriginalConversationID,omitempty"`

	//Organization/MSISDN receiving the transaction
	PartyA string `json:"PartyA"`

	//Type of organization receiving the transaction
	IdentifierType IdentifierType `json:"IdentifierType"`

	//The path that stores information of a transaction
	ResultURL string `json:"ResultURL"`

	//The path that stores information of timeout transaction
	QueueTimeOutURL string `json:"QueueTimeOutURL"`

	//Comments that are sent along with the transaction
	Remarks string `json:"Remarks"`

	//Any additional information to be associated with the transaction
	Occasion string `json:"Occasion"`
}

type RequestBalance struct {
	//This is the credential/username used to authenticate the request
	Initiator string `json:"Initiator"`

	//Base64 encoded string of the M-PESA short code and password, which is encrypted
	//using the M-PESA public key and validates the transaction on the M-PESA Core system.
	//It indicates the Encrypted credential of the initiator getting the account balance.
	//Its value must match the inputted value of the parameter IdentifierType
	SecurityCredential string `json:"SecurityCredential"`

	//A unique command is passed to the M-PESA system. Max length is 64
	CommandID Command `json:"CommandID"`

	//The shortcode of the organization querying for the account balance
	PartyA string `json:"PartyA"`

	//Type of organization querying for the account balance
	IdentifierType IdentifierType `json:"IdentifierType"`

	//Comments that are sent along with the transaction
	//String sequence of characters up to 100
	Remarks string `json:"Remarks"`

	//The end-point that receives a timeout message
	QueueTimeOutURL string `json:"QueueTimeOutURL"`

	//It indicates the destination URL which Daraja should send the result message to
	ResultURL string `json:"ResultURL"`
}

type RequestOrgInfoQuery struct {
	//Type of organization querying for the account balance
	IdentifierType IdentifierType `json:"IdentifierType"`

	//Can be one of `paybill` or `'buygoods`
	Identifier string `json:"Identifier"`
}

// RESPONSE MODELS

type ResponseAuthorization struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   string `json:"expires_in"`
}

type ResponseDefault struct {
	//This is a global unique identifier for the request
	//returned by the M-PESA upon successful request submission
	ConversationID string `json:"ConversationID"`

	//This is a global unique identifier for the request
	//returned by the API proxy upon successful request submission
	OriginatorConversationID string `json:"OriginatorConversationID"`

	//It indicates whether Mobile Money accepts the request or not
	ResponseCode ResponseCode `json:"ResponseCode"`

	//This is the description of the request submission status
	ResponseDescription string `json:"ResponseDescription"`
}

type ResponseB2C ResponseDefault
type ResponseB2B ResponseDefault
type ResponseBalance = ResponseDefault
type ResponseTransactionStatus = ResponseDefault
type ResponseReversal ResponseDefault

type ResponseC2BExpress struct {
	//This is a global unique Identifier for any submitted payment request
	MerchantRequestID string `json:"MerchantRequestID"`

	//This is a global unique identifier of the processed checkout transaction request
	CheckoutRequestID string `json:"CheckoutRequestID"`

	//This is a Numeric status code that indicates the status of the transaction submission.
	//0 means successful submission, and any other code means an error occurred
	ResponseCode ResponseCode `json:"ResponseCode"`

	//Response description is an acknowledgment message from the API
	//that gives the status of the request submission
	ResponseDescription string `json:"ResponseDescription"`

	//This is a message that your system can display to the customer
	//as an acknowledgment of the payment request submission
	CustomerMessage string `json:"CustomerMessage"`
}

type ResponseC2BExpressQuery struct {
	//This is a Numeric status code that indicates the status of the transaction submission.
	//0 means successful submission, and any other code means an error occurred
	ResponseCode ResponseCode `json:"ResponseCode"`

	//Response description is an acknowledgment message from the API
	//that gives the status of the request submission
	ResponseDescription string `json:"ResponseDescription"`

	//This is a global unique Identifier for any submitted payment request
	MerchantRequestID string `json:"MerchantRequestID"`

	//This is a global unique identifier of the processed checkout transaction request
	CheckoutRequestID string `json:"CheckoutRequestID"`

	//This is a numeric status code that indicates the status of the transaction processing.
	//0 means successful processing, and any other code means an error occurred or the transaction failed
	ResultCode string `json:"ResultCode"`

	//The result description is a message from the API that gives the status of the request processing.
	//It can be a success processing message or an error description message
	ResultDesc string `json:"ResultDesc"`
}

type ResponseOrgInfoQuery struct {
	ConversationID string `json:"ConversationID"`
	//0 or 4000 means a successful request
	ResponseCode ResponseCode `json:"ResponseCode"`
	//Text description of the response code
	ResponseMessage string `json:"ResponseMessage"`
	DetailedMessage string `json:"DetailedMessage"`
	//Name of the organization
	OrganizationName string `json:"OrganizationName"`
	StoreName        string `json:"StoreName"`
	//Organization shortcode
	OrganizationShortCode string `json:"OrganizationShortCode"`
	ChargeProfileID       string `json:"ChargeProfileID"`
}

// WEBHOOK REQUEST MODELS

type WebhookRequestDirectC2B struct {
	//The transaction type specified during the payment request enum("Pay Bill", "Buy Goods")
	TransactionType string `json:"TransactionType"`

	//This is the unique M-Pesa transaction ID for every payment request.
	//This is sent to both the call-back messages and a confirmation SMS sent to the customer
	TransID string `json:"TransID"`

	//This is the Timestamp of the transaction, normally in the format of
	//YEAR+MONTH+DATE+HOUR+MINUTE+SECOND (YYYYMMDDHHMMSS). Each part should be
	//at least two digits apart from the year which takes four digits
	TransTime string `json:"TransTime"`

	//This is the amount transacted (normally a numeric value), money that
	//the customer pays to the Shortcode. Only whole numbers are supported
	TransAmount string `json:"TransAmount"`

	//This is the organization's shortcode (Paybill or Buygoods - a 5 to 6-digit account number)
	//used to identify an organization and receive the transaction
	BusinessShortCode string `json:"BusinessShortCode"`

	//This is the account number for which the customer is making the payment.
	//This is only applicable to Customer PayBill Transactions
	BillRefNumber string `json:"BillRefNumber"`

	//InvoiceNumber     string `json:"InvoiceNumber"`

	//The current utility account balance of the payment-receiving organization shortcode.
	//For validation requests, this field is usually blank, whereas, for the confirmation message,
	//the value represents the new balance after the payment has been received
	OrgAccountBalance string `json:"OrgAccountBalance"`

	//This is a transaction ID that the partner can use to identify the transaction.
	//When a validation request is sent, the partner can respond with ThirdPartyTransID
	//and this will be sent back with the Confirmation notification
	ThirdPartyTransID string `json:"ThirdPartyTransID"`

	//This is a masked mobile number of the customer making the payment
	MSISDN string `json:"MSISDN"`

	//The customer's first name is as per the M-Pesa register. This parameter can be empty
	FirstName string `json:"FirstName"`
}

type WebhookRequestC2BValidation = WebhookRequestDirectC2B
type WebhookRequestC2BConfirmation = WebhookRequestDirectC2B

type WebhookRequestC2BExpress struct {
	Body struct {
		StkCallback struct {
			//This is a globally unique identifier of the processed checkout transaction request.
			//This is the same value returned in the acknowledgment message of the initial request
			CheckoutRequestID string `json:"CheckoutRequestID"`

			//This is a global unique Identifier for any submitted payment request.
			//This is the same value returned in the acknowledgment message of the initial request
			MerchantRequestID string `json:"MerchantRequestID"`

			//This is a numeric status code that indicates the status of the transaction processing.
			//0 means successful processing, and any other code means an error occurred or the transaction failed
			ResultCode ResultCode `json:"ResultCode"`

			//The result description is a message from the API that gives the status of the request processing
			ResultDesc string `json:"ResultDesc"`

			//This is the JSON object that holds more details for the transaction.
			//It is only returned for successful transactions
			CallbackMetadata *struct {

				//Item Array holds additional transaction details in JSON objects. Since this array is
				//returned under the CallbackMetadata, it is only returned for successful transactions
				Item []struct {
					//Possible values for Name and corresponding type for Value are
					// - "Amount" (float) This is the Amount that was transacted
					// - "MpesaReceiptNumber" (string) This is the unique M-PESA transaction ID for the payment request.
					// - "Balance" (float) This is the Balance of the account for the shortcode used as partyB
					// - "PhoneNumber" (int64) This is the number of the customer who made the payment
					// - "TransactionDate" (int64) The date and time that the transaction was completed in the format YYYYMMDDHHmmss
					Name  string      `json:"Name"`
					Value interface{} `json:"Value,omitempty"`
				} `json:"Item"`
			} `json:"CallbackMetadata,omitempty"`
		} `json:"stkCallback"`
	} `json:"Body"`
}

// WebhookRequestDefault model represents the result structure returned by some common
// apis e.g. b2c, reversal and balance.
//
// Example results sent in Result.ResultParameters.ResultParameter[n].Key
// For B2C result
// - "TransactionAmount" (number)
// - "TransactionReceipt" (string)
// - "B2CRecipientIsRegisteredCustomer" (string)
// - "B2CChargesPaidAccountAvailableFunds" (number)
// - "ReceiverPartyPublicName" (string)
// - "TransactionCompletedDateTime" (string) e.g. "12.03.2024 01:26:31"
// - "B2CUtilityAccountAvailableFunds" (number)
// - "B2CWorkingAccountAvailableFunds" (number)
//
// For B2B result
// - "DebitAccountBalance"
// - "Amount" (number)
// - "DebitPartyAffectedAccountBalance"
// - "TransCompletedTime" (number)
// - "DebitPartyCharges" Transaction fee deducted on the debit party if applicable
// - "ReceiverPartyPublicName" (string) The public name of the credit party/organization
// - "Currency" (string) A currency code of the transaction amount
// - "InitiatorAccountCurrentBalance" (string) The balance in the organization accounts from which funds were deducted under the shortcode
//
// For C2B Reversal
// - "DebitAccountBalance"
// - "Amount"
// - "TransCompletedTime"
// - "OriginalTransactionID"
// - "Charge"
// - "CreditPartyPublicName"
// - "DebitPartyPublicName"
//
// For Balance
//   - "AccountBalance"  `Working Account|KES|700000.00|700000.00|0.00|0.00&
//     Float Account|KES|0.00|0.00|0.00|0.00&
//     Utility Account|KES|228037.00|228037.00|0.00|0.00&
//     Charges Paid Account|KES|-1540.00|-1540.00|0.00|0.00&
//     Organization Settlement Account|KES|0.00|0.00|0.00|0.00`
//
// For Transaction Status
// - "DebitPartyName" (string)
// - "CreditPartyName" (string)
// - "OriginatorConversationID" (string)
// - "InitiatedTime" (number)
// - "DebitAccountType" (string)
// - "DebitPartyCharges" (string - optional)
// - "TransactionReason" (string - optional)
// - "ReasonType" (string)
// - "TransactionStatus" (string) enum{"Completed",}
// - "FinalisedTime" (number)
// - "Amount" (number)
// - "ConversationID" (string)
// - "ReceiptNo" (string)
type WebhookRequestDefault struct {
	Result struct {
		//This is a status code that indicates whether the transaction was already sent to your listener.
		//Usual value is 0
		ResultType int `json:"ResultType"`

		//This is a numeric status code that indicates the status of the transaction processing.
		//0 means success, and any other code means an error occurred or the transaction failed
		ResultCode ResultCode `json:"ResultCode"`

		//This is a message from the API that gives the status of the request processing
		//and usually maps to a specific result code value
		ResultDesc string `json:"ResultDesc"`

		//This is a global unique identifier for the transaction request returned by
		//the API proxy upon successful request submission
		OriginatorConversationID string `json:"OriginatorConversationID"`

		//This is a global unique identifier for the transaction request returned
		//by the M-PESA upon successful request submission
		ConversationID string `json:"ConversationID"`

		//This is a unique M-PESA transaction ID for every payment request.
		//The same value is sent to the customer over SMS upon successful processing
		TransactionID string `json:"TransactionID"`

		//This is a JSON object that holds more details for the transaction/result
		ResultParameters *struct {
			ResultParameter []struct {
				Key   string      `json:"Key"`
				Value interface{} `json:"Value,omitempty"`
			} `json:"ResultParameter"`
		} `json:"ResultParameters,omitempty"`
		ReferenceData *struct {
			ReferenceItem struct {
				Key   string `json:"Key"`
				Value string `json:"Value"`
			} `json:"ReferenceItem"`
		} `json:"ReferenceData,omitempty"`
	} `json:"Result"`
}

type WebhookRequestC2BReversal WebhookRequestDefault
type WebhookRequestB2C WebhookRequestDefault
type WebhookRequestTransactionStatus WebhookRequestDefault

type WebhookRequestB2B struct {
	Result struct {
		//This is a status code that indicates whether the transaction was already sent to your listener.
		//Usual value is 0
		ResultType int `json:"ResultType"`

		//This is a numeric status code that indicates the status of the transaction processing.
		//0 means success, and any other code means an error occurred or the transaction failed
		ResultCode ResultCode `json:"ResultCode"`

		//This is a message from the API that gives the status of the request processing
		//and usually maps to a specific result code value
		ResultDesc string `json:"ResultDesc"`

		//This is a global unique identifier for the transaction request returned by
		//the API proxy upon successful request submission
		OriginatorConversationID string `json:"OriginatorConversationID"`

		//This is a global unique identifier for the transaction request returned
		//by the M-PESA upon successful request submission
		ConversationID string `json:"ConversationID"`

		//This is a unique M-PESA transaction ID for every payment request.
		//The same value is sent to the customer over SMS upon successful processing
		TransactionID string `json:"TransactionID"`

		//This is a JSON object that holds more details for the transaction/result
		ResultParameters *struct {
			ResultParameter []struct {
				Key   string      `json:"Key"`
				Value interface{} `json:"Value,omitempty"`
			} `json:"ResultParameter"`
		} `json:"ResultParameters,omitempty"`
		ReferenceData *struct {
			ReferenceItem []struct {
				Key   string      `json:"Key"`
				Value interface{} `json:"Value"`
			} `json:"ReferenceItem"`
		} `json:"ReferenceData,omitempty"`
	} `json:"Result"`
}

// WebhookRequestBalance model represent result after fetching shortcode balance
//
// For B2C shortcodes
// Utility Account - A utility account in B2C is for the disbursement of funds
// Working (MMF) Account - It’s a transition account that holds money awaiting
// settlement to the bank or your individual M-PESA account
// Charges Paid - This account deducts charges incurred depending on the business tariff you are in
type WebhookRequestBalance WebhookRequestDefault

// RESPONSE MODELS TO CALLBACK REQUESTS

type WebhookResponseValidation struct {
	//A code indicating whether to complete the transaction. 0(Zero) always means complete.
	//Other values mean canceling the transaction, which also determines the customer notification SMS type
	ResultCode ReplyCode `json:"ResultCode"`

	//Short validation result description
	ResultDesc string `json:"ResultDesc"`

	//An optional value that can be used to identify the payment during a confirmation callback.
	//If a value is set, it would be passed back in a confirmation callback
	ThirdPartyTransID *string `json:"ThirdPartyTransID,omitempty"`
}
