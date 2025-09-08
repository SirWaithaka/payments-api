package tanda

import (
	"time"
)

type Command string

const (
	// CommandCustomerToMerchantMobileMoneyPayment is a Mobile Money payment prompt to a customer through
	// USSD push or STK push. The customer receives a prompt to enter their mobile money account PIN
	CommandCustomerToMerchantMobileMoneyPayment Command = "CustomerToMerchantMobileMoneyPayment"
	// CommandMerchantToCustomerMobileMoneyPayment is an option for sending funds from a customer via mobile money.
	CommandMerchantToCustomerMobileMoneyPayment Command = "MerchantToCustomerMobileMoneyPayment"
	// CommandMerchantToCustomerBankPayment Instantly send money to any Bank account from a Tanda wallet
	CommandMerchantToCustomerBankPayment Command = "MerchantToCustomerBankPayment"
	// CommandMerchantTo3rdPartyMerchantPayment is for sending money to an M-Pesa till from a Merchant float wallet
	CommandMerchantTo3rdPartyMerchantPayment Command = "MerchantTo3rdPartyMerchantPayment"
	// CommandMerchantToMerchantTandaPayment is an On-Net/Internal transfer.
	// It moves money from a Merchant Collection wallet to a Merchant Disbursement Wallet
	CommandMerchantToMerchantTandaPayment Command = "MerchantToMerchantTandaPayment"
	// CommandMerchantTo3rdPartyBusinessPayment is for sending money to an M-Pesa Paybill from a Merchant float wallet
	CommandMerchantTo3rdPartyBusinessPayment Command = "MerchantTo3rdPartyBusinessPayment"
	// CommandInternationalMoneyTransferBank Send money to your customer's bank account in other countries
	CommandInternationalMoneyTransferBank Command = "InternationalMoneyTransferBank"
	// CommandInternationalMoneyTransferMobile Send money to your customer's mobile wallet
	// in other countries through mobile money.
	CommandInternationalMoneyTransferMobile Command = "InternationalMoneyTransferMobile"
)

type PaymentStatus string

const (
	PaymentStatusS000000 PaymentStatus = "S000000" // Successful and fulfilled
	PaymentStatusP202000 PaymentStatus = "P202000" // Request has been received and is currently being processed.
	PaymentStatusE400000 PaymentStatus = "E400000" // Bad request
	PaymentStatusE401000 PaymentStatus = "E401000" // Unauthorized
	PaymentStatusE403000 PaymentStatus = "E403000" // Access denied
	PaymentStatusE404000 PaymentStatus = "E404000" // Not found
	PaymentStatusE409000 PaymentStatus = "E409000" // Duplicate resource found
	PaymentStatusE422005 PaymentStatus = "E422005" // Request failed. Product not found
	PaymentStatusE422006 PaymentStatus = "E422006" // Request failed. Insufficient Wallet balance
	PaymentStatusE422022 PaymentStatus = "E422022" // Payment Request Validation Failed
	PaymentStatusE500000 PaymentStatus = "E500000" // Internal Server Error
	PaymentStatusE501000 PaymentStatus = "E501000" // Not implemented
	PaymentStatusE503000 PaymentStatus = "E503000" // Service unavailable. Product or service is disabled or unavailable
	PaymentStatusE000002 PaymentStatus = "E000002" // Third party Error
)

type ParameterID string

const (
	ParameterIDAmount                     ParameterID = "amount"                     //  The amount of money in KES
	ParameterIDShortCode                  ParameterID = "shortCode"                  //  Uniquely identifies an entity making a request.
	ParameterIDAccountNumber              ParameterID = "accountNumber"              //  Mobile Phone Number
	ParameterIDNarration                  ParameterID = "narration"                  //  A short description of the transaction
	ParameterIDIpnUrl                     ParameterID = "ipnUrl"                     //  URL to receive Instant payment notifications
	ParameterIDAccountName                ParameterID = "accountName"                //  Beneficiary account name
	ParameterIDBankCode                   ParameterID = "bankCode"                   //  Beneficiary bank code
	ParameterIDPartyA                     ParameterID = "partyA"                     //  Uniquely identifies an entity making a request
	ParameterIDPartyB                     ParameterID = "partyB"                     //  A valid M-Pesa till
	ParameterIDBusinessNumber             ParameterID = "businessNumber"             //  M-Pesa pay bill business number
	ParameterIDAccountReference           ParameterID = "accountReference"           //  M-Pesa pay bill account number
	ParameterIDCurrency                   ParameterID = "currency"                   //  Beneficiary currency
	ParameterIDMobileNumber               ParameterID = "mobileNumber"               //  Beneficiary Mobile number
	ParameterIDSenderType                 ParameterID = "senderType"                 //  The type of the entity sending money. e.g. COMPANY, INDIVIDUAL
	ParameterIDBeneficiaryType            ParameterID = "beneficiaryType"            //  The type of the entity receiving money. e.g. COMPANY, INDIVIDUAL
	ParameterIDBeneficiaryAddress         ParameterID = "beneficiaryAddress"         //  Beneficiary address
	ParameterIDBeneficiaryActivity        ParameterID = "beneficiaryActivity"        //  Beneficiary Activity/Job/Economic Activity. e.g. accountant
	ParameterIDBeneficiaryCountry         ParameterID = "beneficiaryCountry"         //  Beneficiary country
	ParameterIDBeneficiaryEmailAddress    ParameterID = "beneficiaryEmailAddress"    //  Beneficiary email address
	ParameterIDDocumentType               ParameterID = "documentType"               //  Beneficiary document type
	ParameterIDDocumentNumber             ParameterID = "documentNumber"             //  Beneficiary document number
	ParameterIDSenderName                 ParameterID = "senderName"                 //  Sender Name
	ParameterIDSenderAddress              ParameterID = "senderAddress"              //  Sender Address
	ParameterIDSenderPhoneNumber          ParameterID = "senderPhoneNumber"          //  Sender Phone Number
	ParameterIDSenderDocumentType         ParameterID = "senderDocumentType"         //  Sender document type
	ParameterIDSenderDocumentNumber       ParameterID = "senderDocumentNumber"       //  Sender document number
	ParameterIDSenderCountry              ParameterID = "senderCountry"              //  Sender country
	ParameterIDSenderCurrency             ParameterID = "senderCurrency"             //  Sender currency
	ParameterIDSenderSourceOfFunds        ParameterID = "senderSourceOfFunds"        //  Sender source of funds
	ParameterIDSenderPrincipalActivity    ParameterID = "senderPrincipalActivity"    //  Sender principal activity e.g. business
	ParameterIDSenderBankCode             ParameterID = "senderBankCode"             //  Sender bank code (optional)
	ParameterIDSenderEmailAddress         ParameterID = "senderEmailAddress"         //  Sender email address
	ParameterIDSenderPrimaryAccountNumber ParameterID = "senderPrimaryAccountNumber" //  Sender primary account number
	ParameterIDSenderDateOfBirth          ParameterID = "senderDateOfBirth"          //  Sender date of birth
	ParameterIDSenderCompanyName          ParameterID = "senderCompanyName"          // Applicable when senderType is COMPANY
)

// REQUEST MODELS

// RequestAuthentication describes the payload in form data required for authentication
type RequestAuthentication struct {
	ClientID     string
	ClientSecret string
}

// RequestPayment describes the JSON payload format required to make payment requests
type RequestPayment struct {
	// Specifies the command to be executed
	CommandID Command `json:"commandId"`
	// Specifies the service provider for a mobile money payment or bank transfer
	ServiceProviderID string `json:"serviceProviderId"`
	// A unique reference for the transaction.
	// It should be between 8-16 alphanumeric characters
	Reference string `json:"reference"`
	// Object array containing additional request details
	//
	// Valid values for PaymentRequestParameter.ID are:
	// When CommandID is "CustomerToMerchantMobileMoneyPayment" or
	// "MerchantToCustomerMobileMoneyPayment"
	// -> amount, shortCode, accountNumber, narration, ipnUrl
	//
	// When CommandID is "MerchantToCustomerBankPayment"
	// -> amount, shortCode, accountNumber, accountName, bankCode, narration, ipnUrl
	//
	// When CommandID is "MerchantTo3rdPartyMerchantPayment" or "MerchantToMerchantTandaPayment"
	// -> amount, partyA, partyB, narration, ipnUrl
	//
	// When CommandID is "MerchantTo3rdPartyBusinessPayment"
	// -> amount, shortCode, businessNumber, accountReference, narration, ipnUrl
	//
	// When CommandID is "InternationalMoneyTransferBank"
	// -> amount, currency, mobileNumber, accountName, accountNumber, bankCode, senderType,
	// beneficiaryType, beneficiaryAddress, beneficiaryActivity, beneficiaryCountry,
	// beneficiaryEmailAddress, documentType, documentNumber, narration, senderName, senderAddress,
	// senderPhoneNumber, senderDocumentType, senderDocumentNumber, senderCountry, senderCurrency,
	// senderSourceOfFunds, senderPrincipalActivity, senderBankCode, senderEmailAddress,
	// senderPrimaryAccountNumber, senderDateOfBirth, ipnUrl, shortCode
	//
	// When CommandID is "InternationalMoneyTransferMobile"
	// -> amount, currency, mobileNumber, accountName, accountNumber, senderType, senderCompanyName,
	// beneficiaryType, beneficiaryActivity, beneficiaryCountry, documentType, documentNumber, narration,
	// senderName, senderPhoneNumber, senderDocumentType, senderDocumentNumber, senderCountry,
	// senderCurrency, senderSourceOfFunds, senderPrincipalActivity, ipnUrl, shortCode
	Request []PaymentRequestParameter `json:"request"`
}

// AddParameter appends a new parameter to RequestPayment.Request
func (r *RequestPayment) AddParameter(id ParameterID, value string) {
	r.Request = append(r.Request, PaymentRequestParameter{
		ID:    id,
		Value: value,
	})
}

type PaymentRequestParameter struct {
	ID    ParameterID `json:"id"`
	Value string      `json:"value"`
	Label string      `json:"label"`
}

// RESPONSE MODELS

// ErrorResponse describes the JSON response from a request that fails
//
// Example
//
//	{
//	 "status": "E401000",
//	 "category": "Business",
//	 "severity": "Low",
//	 "error": "Unauthorized",
//	 "description": "invalid token."
//	}
type ErrorResponse struct {
	Status      string `json:"status"`
	Category    string `json:"category"`
	Severity    string `json:"severity"`
	Error       string `json:"error"`
	Description string `json:"description"`
}

// ResponseAuthentication describes JSON response from authentication endpoint
//
// Example
//
//	{
//	 "access_token": "eyJraWQiOiIxZjhjYTdhMC0yZDAwLTQxNDItODRmMi1iMzkzMWJiYzNjYjYiLCJhbGciOiJSUzI1NiJ9...",
//	 "token_type": "Bearer",
//	 "expires_in": 3599
//	}
type ResponseAuthentication struct {
	TokenType   string  `json:"token_type"`
	IssuedAt    *string `json:"issued_at,omitempty"`
	ExpiresIn   uint    `json:"expires_in"`
	AccessToken string  `json:"access_token"`
}

type ResponsePayment struct {
	TrackingID string        `json:"trackingId"`
	Reference  string        `json:"reference"`
	Status     PaymentStatus `json:"status"`
	Message    string        `json:"message"`
}

type ResponseTransactionStatus struct {
	Status  PaymentStatus `json:"status"`
	Message string        `json:"message"`
}

// WEBHOOK REQUEST MODELS

type WebhookRequestPaymentStatus struct {
	// A unique identifier for tracking payments, generated on every successful request.
	TrackingID string `json:"trackingId"`
	// A unique identifier for the transaction.
	TransactionID string `json:"transactionId"`
	// A unique reference for the transaction that was generated by the API consumer.
	Reference string `json:"reference"`
	// The status code of the payment.
	Status PaymentStatus `json:"status"`
	// A description of the status.
	Message string `json:"message"`
	// The timestamp of when the IPN was sent.
	Timestamp time.Time `json:"timestamp"`
	// Additional parameters from the service provider like transaction ref e.g. `ref`
	Result struct {
		Ref string `json:"ref"`
	} `json:"result"`
}
