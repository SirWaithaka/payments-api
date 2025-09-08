package tanda

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/SirWaithaka/payments-api/request"
)

func TestPaymentParametersValidator(t *testing.T) {
	testcases := []struct {
		name        string
		params      interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Invalid payload type",
			params:      "invalid",
			expectError: true,
			errorMsg:    "invalid payload",
		},
		{
			name: "Empty parameters",
			params: RequestPayment{
				CommandID: CommandCustomerToMerchantMobileMoneyPayment,
				Request:   nil,
			},
			expectError: true,
			errorMsg:    "invalid payload",
		},
		{
			name: "Empty command ID",
			params: RequestPayment{
				CommandID: "",
				Request: []PaymentRequestParameter{
					{ID: ParameterIDAmount, Value: "100", Label: "Amount"},
				},
			},
			expectError: true,
			errorMsg:    "invalid payload",
		},
		{
			name: "Invalid command ID",
			params: RequestPayment{
				CommandID: Command("InvalidCommand"),
				Request: []PaymentRequestParameter{
					{ID: ParameterIDAmount, Value: "100", Label: "Amount"},
				},
			},
			expectError: true,
			errorMsg:    "invalid command id: InvalidCommand",
		},
		{
			name: "Valid CustomerToMerchantMobileMoneyPayment - all required parameters",
			params: RequestPayment{
				CommandID: CommandCustomerToMerchantMobileMoneyPayment,
				Request: []PaymentRequestParameter{
					{ID: ParameterIDAmount, Value: "100", Label: "Amount"},
					{ID: ParameterIDShortCode, Value: "174379", Label: "Short Code"},
					{ID: ParameterIDAccountNumber, Value: "254712345678", Label: "Phone Number"},
					{ID: ParameterIDNarration, Value: "Payment", Label: "Description"},
					{ID: ParameterIDIpnUrl, Value: "https://example.com/callback", Label: "Callback URL"},
				},
			},
			expectError: false,
		},
		{
			name: "Valid MerchantToCustomerMobileMoneyPayment - all required parameters",
			params: RequestPayment{
				CommandID: CommandMerchantToCustomerMobileMoneyPayment,
				Request: []PaymentRequestParameter{
					{ID: ParameterIDAmount, Value: "100", Label: "Amount"},
					{ID: ParameterIDShortCode, Value: "174379", Label: "Short Code"},
					{ID: ParameterIDAccountNumber, Value: "254712345678", Label: "Phone Number"},
					{ID: ParameterIDNarration, Value: "Payment", Label: "Description"},
					{ID: ParameterIDIpnUrl, Value: "https://example.com/callback", Label: "Callback URL"},
				},
			},
			expectError: false,
		},
		{
			name: "Missing required parameter for CustomerToMerchantMobileMoneyPayment",
			params: RequestPayment{
				CommandID: CommandCustomerToMerchantMobileMoneyPayment,
				Request: []PaymentRequestParameter{
					{ID: ParameterIDAmount, Value: "100", Label: "Amount"},
					{ID: ParameterIDShortCode, Value: "174379", Label: "Short Code"},
					// Missing ParameterIDAccountNumber, ParameterIDNarration, ParameterIDIpnUrl
				},
			},
			expectError: true,
			errorMsg:    "missing required parameter: accountNumber",
		},
		{
			name: "Valid MerchantToCustomerBankPayment - all required parameters",
			params: RequestPayment{
				CommandID: CommandMerchantToCustomerBankPayment,
				Request: []PaymentRequestParameter{
					{ID: ParameterIDAmount, Value: "500", Label: "Amount"},
					{ID: ParameterIDShortCode, Value: "174379", Label: "Short Code"},
					{ID: ParameterIDAccountNumber, Value: "1234567890", Label: "Account Number"},
					{ID: ParameterIDAccountName, Value: "John Doe", Label: "Account Name"},
					{ID: ParameterIDBankCode, Value: "01", Label: "Bank Code"},
					{ID: ParameterIDNarration, Value: "Bank transfer", Label: "Description"},
					{ID: ParameterIDIpnUrl, Value: "https://example.com/callback", Label: "Callback URL"},
				},
			},
			expectError: false,
		},
		{
			name: "Missing bank-specific parameter for MerchantToCustomerBankPayment",
			params: RequestPayment{
				CommandID: CommandMerchantToCustomerBankPayment,
				Request: []PaymentRequestParameter{
					{ID: ParameterIDAmount, Value: "500", Label: "Amount"},
					{ID: ParameterIDShortCode, Value: "174379", Label: "Short Code"},
					{ID: ParameterIDAccountNumber, Value: "1234567890", Label: "Account Number"},
					{ID: ParameterIDNarration, Value: "Bank transfer", Label: "Description"},
					{ID: ParameterIDIpnUrl, Value: "https://example.com/callback", Label: "Callback URL"},
					// Missing ParameterIDAccountName and ParameterIDBankCode
				},
			},
			expectError: true,
			errorMsg:    "missing required parameter: accountName",
		},
		{
			name: "Valid MerchantTo3rdPartyMerchantPayment - all required parameters",
			params: RequestPayment{
				CommandID: CommandMerchantTo3rdPartyMerchantPayment,
				Request: []PaymentRequestParameter{
					{ID: ParameterIDAmount, Value: "200", Label: "Amount"},
					{ID: ParameterIDPartyA, Value: "600000", Label: "Party A"},
					{ID: ParameterIDPartyB, Value: "174379", Label: "Party B"},
					{ID: ParameterIDNarration, Value: "Transfer", Label: "Description"},
					{ID: ParameterIDIpnUrl, Value: "https://example.com/callback", Label: "Callback URL"},
				},
			},
			expectError: false,
		},
		{
			name: "Valid MerchantToMerchantTandaPayment - all required parameters",
			params: RequestPayment{
				CommandID: CommandMerchantToMerchantTandaPayment,
				Request: []PaymentRequestParameter{
					{ID: ParameterIDAmount, Value: "300", Label: "Amount"},
					{ID: ParameterIDPartyA, Value: "600000", Label: "Party A"},
					{ID: ParameterIDPartyB, Value: "600001", Label: "Party B"},
					{ID: ParameterIDNarration, Value: "Internal transfer", Label: "Description"},
					{ID: ParameterIDIpnUrl, Value: "https://example.com/callback", Label: "Callback URL"},
				},
			},
			expectError: false,
		},
		{
			name: "Valid MerchantTo3rdPartyBusinessPayment - all required parameters",
			params: RequestPayment{
				CommandID: CommandMerchantTo3rdPartyBusinessPayment,
				Request: []PaymentRequestParameter{
					{ID: ParameterIDAmount, Value: "150", Label: "Amount"},
					{ID: ParameterIDShortCode, Value: "174379", Label: "Short Code"},
					{ID: ParameterIDBusinessNumber, Value: "400200", Label: "Business Number"},
					{ID: ParameterIDAccountReference, Value: "INV001", Label: "Account Reference"},
					{ID: ParameterIDNarration, Value: "Paybill payment", Label: "Description"},
					{ID: ParameterIDIpnUrl, Value: "https://example.com/callback", Label: "Callback URL"},
				},
			},
			expectError: false,
		},
		{
			name: "Missing parameter for MerchantTo3rdPartyBusinessPayment",
			params: RequestPayment{
				CommandID: CommandMerchantTo3rdPartyBusinessPayment,
				Request: []PaymentRequestParameter{
					{ID: ParameterIDAmount, Value: "150", Label: "Amount"},
					{ID: ParameterIDShortCode, Value: "174379", Label: "Short Code"},
					{ID: ParameterIDNarration, Value: "Paybill payment", Label: "Description"},
					{ID: ParameterIDIpnUrl, Value: "https://example.com/callback", Label: "Callback URL"},
					// Missing ParameterIDBusinessNumber and ParameterIDAccountReference
				},
			},
			expectError: true,
			errorMsg:    "missing required parameter: businessNumber",
		},
		{
			name: "Valid InternationalMoneyTransferBank - all required parameters",
			params: RequestPayment{
				CommandID: CommandInternationalMoneyTransferBank,
				Request: []PaymentRequestParameter{
					{ID: ParameterIDAmount, Value: "1000", Label: "Amount"},
					{ID: ParameterIDCurrency, Value: "USD", Label: "Currency"},
					{ID: ParameterIDMobileNumber, Value: "254712345678", Label: "Mobile Number"},
					{ID: ParameterIDAccountName, Value: "Jane Smith", Label: "Account Name"},
					{ID: ParameterIDAccountNumber, Value: "9876543210", Label: "Account Number"},
					{ID: ParameterIDBankCode, Value: "SWIFT123", Label: "Bank Code"},
					{ID: ParameterIDSenderType, Value: "INDIVIDUAL", Label: "Sender Type"},
					{ID: ParameterIDBeneficiaryType, Value: "INDIVIDUAL", Label: "Beneficiary Type"},
					{ID: ParameterIDBeneficiaryAddress, Value: "123 Main St", Label: "Beneficiary Address"},
					{ID: ParameterIDBeneficiaryActivity, Value: "accountant", Label: "Beneficiary Activity"},
					{ID: ParameterIDBeneficiaryCountry, Value: "US", Label: "Beneficiary Country"},
					{ID: ParameterIDBeneficiaryEmailAddress, Value: "jane@example.com", Label: "Beneficiary Email"},
					{ID: ParameterIDDocumentType, Value: "passport", Label: "Document Type"},
					{ID: ParameterIDDocumentNumber, Value: "P123456", Label: "Document Number"},
					{ID: ParameterIDNarration, Value: "International transfer", Label: "Narration"},
					{ID: ParameterIDSenderName, Value: "John Doe", Label: "Sender Name"},
					{ID: ParameterIDSenderAddress, Value: "456 Oak Ave", Label: "Sender Address"},
					{ID: ParameterIDSenderPhoneNumber, Value: "254700123456", Label: "Sender Phone"},
					{ID: ParameterIDSenderDocumentType, Value: "id", Label: "Sender Document Type"},
					{ID: ParameterIDSenderDocumentNumber, Value: "ID123456", Label: "Sender Document Number"},
					{ID: ParameterIDSenderCountry, Value: "KE", Label: "Sender Country"},
					{ID: ParameterIDSenderCurrency, Value: "KES", Label: "Sender Currency"},
					{ID: ParameterIDSenderSourceOfFunds, Value: "salary", Label: "Sender Source of Funds"},
					{ID: ParameterIDSenderPrincipalActivity, Value: "business", Label: "Sender Principal Activity"},
					{ID: ParameterIDSenderBankCode, Value: "BANK001", Label: "Sender Bank Code"},
					{ID: ParameterIDSenderEmailAddress, Value: "john@example.com", Label: "Sender Email"},
					{ID: ParameterIDSenderPrimaryAccountNumber, Value: "ACC123", Label: "Sender Primary Account"},
					{ID: ParameterIDSenderDateOfBirth, Value: "1990-01-01", Label: "Sender Date of Birth"},
					{ID: ParameterIDIpnUrl, Value: "https://example.com/callback", Label: "IPN URL"},
					{ID: ParameterIDShortCode, Value: "174379", Label: "Short Code"},
				},
			},
			expectError: false,
		},
		{
			name: "Valid InternationalMoneyTransferMobile - all required parameters",
			params: RequestPayment{
				CommandID: CommandInternationalMoneyTransferMobile,
				Request: []PaymentRequestParameter{
					{ID: ParameterIDAmount, Value: "500", Label: "Amount"},
					{ID: ParameterIDCurrency, Value: "USD", Label: "Currency"},
					{ID: ParameterIDMobileNumber, Value: "1234567890", Label: "Mobile Number"},
					{ID: ParameterIDAccountName, Value: "Alice Johnson", Label: "Account Name"},
					{ID: ParameterIDAccountNumber, Value: "MOB123456", Label: "Account Number"},
					{ID: ParameterIDSenderType, Value: "COMPANY", Label: "Sender Type"},
					{ID: ParameterIDSenderCompanyName, Value: "Tech Corp", Label: "Sender Company Name"},
					{ID: ParameterIDBeneficiaryType, Value: "INDIVIDUAL", Label: "Beneficiary Type"},
					{ID: ParameterIDBeneficiaryActivity, Value: "teacher", Label: "Beneficiary Activity"},
					{ID: ParameterIDBeneficiaryCountry, Value: "UG", Label: "Beneficiary Country"},
					{ID: ParameterIDDocumentType, Value: "passport", Label: "Document Type"},
					{ID: ParameterIDDocumentNumber, Value: "P654321", Label: "Document Number"},
					{ID: ParameterIDNarration, Value: "Mobile money transfer", Label: "Narration"},
					{ID: ParameterIDSenderName, Value: "Tech Corp Ltd", Label: "Sender Name"},
					{ID: ParameterIDSenderPhoneNumber, Value: "254700987654", Label: "Sender Phone"},
					{ID: ParameterIDSenderDocumentType, Value: "registration", Label: "Sender Document Type"},
					{ID: ParameterIDSenderDocumentNumber, Value: "REG456789", Label: "Sender Document Number"},
					{ID: ParameterIDSenderCountry, Value: "KE", Label: "Sender Country"},
					{ID: ParameterIDSenderCurrency, Value: "KES", Label: "Sender Currency"},
					{ID: ParameterIDSenderSourceOfFunds, Value: "business_revenue", Label: "Sender Source of Funds"},
					{ID: ParameterIDSenderPrincipalActivity, Value: "technology", Label: "Sender Principal Activity"},
					{ID: ParameterIDIpnUrl, Value: "https://example.com/callback", Label: "IPN URL"},
					{ID: ParameterIDShortCode, Value: "174379", Label: "Short Code"},
				},
			},
			expectError: false,
		},
		{
			name: "Missing parameter for InternationalMoneyTransferMobile",
			params: RequestPayment{
				CommandID: CommandInternationalMoneyTransferMobile,
				Request: []PaymentRequestParameter{
					{ID: ParameterIDAmount, Value: "500", Label: "Amount"},
					{ID: ParameterIDCurrency, Value: "USD", Label: "Currency"},
					{ID: ParameterIDMobileNumber, Value: "1234567890", Label: "Mobile Number"},
					// Missing many required parameters
				},
			},
			expectError: true,
			errorMsg:    "missing required parameter: accountName",
		},
		{
			name: "Extra parameters with valid required ones - should pass",
			params: RequestPayment{
				CommandID: CommandCustomerToMerchantMobileMoneyPayment,
				Request: []PaymentRequestParameter{
					{ID: ParameterIDAmount, Value: "100", Label: "Amount"},
					{ID: ParameterIDShortCode, Value: "174379", Label: "Short Code"},
					{ID: ParameterIDAccountNumber, Value: "254712345678", Label: "Phone Number"},
					{ID: ParameterIDNarration, Value: "Payment", Label: "Description"},
					{ID: ParameterIDIpnUrl, Value: "https://example.com/callback", Label: "Callback URL"},
					// Extra parameter - validator should allow this
					{ID: ParameterIDAccountName, Value: "Extra Field", Label: "Extra"},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock request
			req := &request.Request{
				Params: tt.params,
			}

			// Run the validator
			PaymentParametersValidator.Fn(req)

			// Check the result
			if tt.expectError {
				if req.Error == nil {
					t.Errorf("expected error, got nil")
				} else if req.Error.Error() != tt.errorMsg {
					t.Errorf("expected error message %q, got %q", tt.errorMsg, req.Error.Error())
				}
			} else {
				assert.Nil(t, req.Error)
			}
		})
	}
}

func TestGetRequiredParametersForCommand(t *testing.T) {
	tests := []struct {
		name     string
		command  Command
		expected []ParameterID
	}{
		{
			name:    "CustomerToMerchantMobileMoneyPayment",
			command: CommandCustomerToMerchantMobileMoneyPayment,
			expected: []ParameterID{
				ParameterIDAmount,
				ParameterIDShortCode,
				ParameterIDAccountNumber,
				ParameterIDNarration,
				ParameterIDIpnUrl,
			},
		},
		{
			name:    "MerchantToCustomerBankPayment",
			command: CommandMerchantToCustomerBankPayment,
			expected: []ParameterID{
				ParameterIDAmount,
				ParameterIDShortCode,
				ParameterIDAccountNumber,
				ParameterIDAccountName,
				ParameterIDBankCode,
				ParameterIDNarration,
				ParameterIDIpnUrl,
			},
		},
		{
			name:     "Invalid command",
			command:  Command("InvalidCommand"),
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getRequiredParametersForCommand(tt.command)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d parameters, got %d", len(tt.expected), len(result))
				return
			}

			for i, param := range tt.expected {
				if result[i] != param {
					t.Errorf("Expected parameter %s at index %d, got %s", param, i, result[i])
				}
			}
		})
	}
}

func TestParameterIDs(t *testing.T) {
	params := []PaymentRequestParameter{
		{ID: ParameterIDAmount, Value: "100", Label: "Amount"},
		{ID: ParameterIDShortCode, Value: "174379", Label: "Short Code"},
		{ID: ParameterIDAccountNumber, Value: "254712345678", Label: "Phone Number"},
	}

	expected := []ParameterID{
		ParameterIDAmount,
		ParameterIDShortCode,
		ParameterIDAccountNumber,
	}

	var result []ParameterID
	for id := range parameterIDs(params) {
		result = append(result, id)
	}

	if len(result) != len(expected) {
		t.Errorf("Expected %d IDs, got %d", len(expected), len(result))
		return
	}

	for i, id := range expected {
		if result[i] != id {
			t.Errorf("Expected ID %s at index %d, got %s", id, i, result[i])
		}
	}
}
