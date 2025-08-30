package tanda

type Command string

const (
	// CommandCustomerToMerchantMobileMoneyPayment is a Mobile Money payment prompt to a customer through
	// USSD push or STK push. The customer receives a prompt to enter their mobile money account PIN
	CommandCustomerToMerchantMobileMoneyPayment = "CustomerToMerchantMobileMoneyPayment"
	// CommandMerchantToCustomerMobileMoneyPayment is an option for sending funds from a customer via mobile money.
	CommandMerchantToCustomerMobileMoneyPayment = "MerchantToCustomerMobileMoneyPayment"
	// CommandMerchantToCustomerBankPayment Instantly send money to any Bank account from a Tanda wallet
	CommandMerchantToCustomerBankPayment = "MerchantToCustomerBankPayment"
	// CommandMerchantTo3rdPartyMerchantPayment is for sending money to an M-Pesa till from a Merchant float wallet
	CommandMerchantTo3rdPartyMerchantPayment = "MerchantTo3rdPartyMerchantPayment"
	// CommandMerchantToMerchantTandaPayment is an On-Net/Internal transfer.
	// It moves money from a Merchant Collection wallet to a Merchant Disbursement Wallet
	CommandMerchantToMerchantTandaPayment = "MerchantToMerchantTandaPayment"
	// CommandMerchantTo3rdPartyBusinessPayment is for sending money to an M-Pesa Paybill from a Merchant float wallet
	CommandMerchantTo3rdPartyBusinessPayment = "MerchantTo3rdPartyBusinessPayment"
	// CommandInternationalMoneyTransferBank Send money to your customer's bank account in other countries
	CommandInternationalMoneyTransferBank = "InternationalMoneyTransferBank"
	// CommandInternationalMoneyTransferMobile Send money to your customer's mobile wallet
	// in other countries through mobile money.
	CommandInternationalMoneyTransferMobile = "InternationalMoneyTransferMobile"
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
	// Valid values for PaymentRequestParameter.Id are:
	// When CommandID is "CustomerToMerchantMobileMoneyPayment" or
	// "MerchantToCustomerMobileMoneyPayment"
	// -> amount - The amount of money in KES
	// -> shortCode - Uniquely identifies an entity making a request.
	// -> accountNumber - Mobile Phone Number
	// -> narration - A short description of the transaction
	// -> ipnUrl - URL to receive Instant payment notifications
	//
	// When CommandID is "MerchantToCustomerBankPayment"
	// -> amount - The amount of money in KES
	// -> shortCode - Uniquely identifies an entity making a request.
	// -> accountNumber - Beneficiary account number
	// -> accountName - Beneficiary account name
	// -> bankCode - Beneficiary bank code
	// -> narration - A short description of the transaction
	// -> ipnUrl - URL to receive Instant payment notifications
	//
	// When CommandID is "MerchantTo3rdPartyMerchantPayment" or "MerchantToMerchantTandaPayment"
	// -> amount - The amount of money in KES
	// -> partyA - Uniquely identifies an entity making a request
	// -> partyB - A valid M-Pesa till
	// -> narration - Description of the payment.
	// -> ipnUrl - URL to receive Instant payment notifications on.
	//
	// When CommandID is "MerchantTo3rdPartyBusinessPayment"
	// -> amount - The amount of money in KES
	// -> shortCode - Uniquely identifies an entity making a request
	// -> businessNumber - M-Pesa pay bill business number
	// -> accountReference - M-Pesa pay bill account number
	// -> narration - Description of the payment.
	// -> ipnUrl - URL to receive Instant payment notifications on
	//
	// When CommandID is "InternationalMoneyTransferBank"
	// -> amount - The amount of money in KES
	// -> currency - Beneficiary currency
	// -> mobileNumber - Beneficiary Mobile number
	// -> accountName - Beneficiary account name
	// -> accountNumber - Beneficiary account number
	// -> bankCode - Beneficiary bank code
	// -> senderType - The type of the entity sending money. e.g. COMPANY, INDIVIDUAL
	// -> beneficiaryType - The type of the entity receiving money. e.g. COMPANY, INDIVIDUAL
	// -> beneficiaryAddress - Beneficiary address
	// -> beneficiaryActivity - Beneficiary Activity/Job/Economic Activity. e.g. accountant
	// -> beneficiaryCountry - Beneficiary country
	// -> beneficiaryEmailAddress - Beneficiary email address
	// -> documentType - Beneficiary document type
	// -> documentNumber - Beneficiary document number
	// -> narration - narration
	// -> senderName - Sender Name
	// -> senderAddress - Sender Address
	// -> senderPhoneNumber - Sender Phone Number
	// -> senderDocumentType - Sender document type
	// -> senderDocumentNumber - Sender document number
	// -> senderCountry - Sender country
	// -> senderCurrency - Sender currency
	// -> senderSourceOfFunds - Sender source of funds
	// -> senderPrincipalActivity - Sender principal activity e.g. business
	// -> senderBankCode - Sender bank code (optional)
	// -> senderEmailAddress - Sender email address
	// -> senderPrimaryAccountNumber - Sender primary account number
	// -> senderDateOfBirth - Sender date of birth
	// -> ipnUrl - Notification URL
	// -> shortCode - Short code issued by Tanda
	//
	// When CommandID is "InternationalMoneyTransferMobile"
	// -> amount - Amount
	// -> currency - Beneficiary currency
	// -> mobileNumber - Beneficiary Mobile number
	// -> accountName - Beneficiary Account Name
	// -> accountNumber - Beneficiary Mobile Wallet Number
	// -> senderType - The type of the entity sending money. e.g. COMPANY, INDIVIDUAL
	// -> senderCompanyName - Applicable when senderType is COMPANY
	// -> beneficiaryType - The type of the entity sending money. e.g. COMPANY, INDIVIDUAL
	// -> beneficiaryActivity - Beneficiary Activity/Job/Economic Activity. e.g. accountant
	// -> beneficiaryCountry - Beneficiary country
	// -> documentType - Beneficiary legal document type e.g. NationalID, Passport
	// -> documentNumber - Beneficiary document number e.g. NationalID, Passport
	// -> narration - Narration
	// -> senderName - Sender Name
	// -> senderPhoneNumber - Sender Phone Number
	// -> senderDocumentType - Sender legal document type. e.g. NationalID, Passport
	// -> senderDocumentNumber - Sender document number
	// -> senderCountry - Sender country
	// -> senderCurrency - Sender currency
	// -> senderSourceOfFunds - Sender source of funds
	// -> senderPrincipalActivity - Sender principal activity E.g business
	// -> ipnUrl - Notification URL
	// -> shortCode - Short code issued by Tanda
	Request []PaymentRequestParameter `json:"request"`
}

type PaymentRequestParameter struct {
	Id    string `json:"id"`
	Value string `json:"value"`
	Label string `json:"label"`
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
