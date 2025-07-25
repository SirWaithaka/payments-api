package services

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/oklog/ulid/v2"
)

func Int(v int) *int {
	return &v
}

func TestC2BWebHookResult(t *testing.T) {
	type testInput struct {
		resultCode    int
		merchantReqId string
		amount        int
		receiptId     string
	}

	t.Run("test success case", func(t *testing.T) {
		successTestBody := `{"Body":{"stkCallback":{"MerchantRequestID":"%s","CheckoutRequestID":"ws_CO_02072024204225888790902376","ResultCode":%d,"ResultDesc":"The service request is processed successfully.","CallbackMetadata":{"Item":[{"Name":"Amount","Value":%d},{"Name":"MpesaReceiptNumber","Value":"%s"},{"Name":"Balance"},{"Name":"TransactionDate","Value":20240702204236},{"Name":"PhoneNumber","Value":254790902376}]}}}}`

		tcs := []struct {
			input testInput
		}{
			{input: testInput{resultCode: 0, merchantReqId: uuid.Must(uuid.NewV7()).String(), amount: 100, receiptId: ulid.Make().String()}},
			{input: testInput{resultCode: 0, merchantReqId: uuid.Must(uuid.NewV7()).String(), amount: 1000, receiptId: ulid.Make().String()}},
			{input: testInput{resultCode: 0, merchantReqId: uuid.Must(uuid.NewV7()).String(), amount: 3000, receiptId: ulid.Make().String()}},
			{input: testInput{resultCode: 0, merchantReqId: uuid.Must(uuid.NewV7()).String(), amount: 70000, receiptId: ulid.Make().String()}},
			{input: testInput{resultCode: 0, merchantReqId: uuid.Must(uuid.NewV7()).String(), amount: 100000, receiptId: ulid.Make().String()}},
			{input: testInput{resultCode: 0, merchantReqId: uuid.Must(uuid.NewV7()).String(), amount: 150000, receiptId: ulid.Make().String()}},
		}

		for _, tc := range tcs {
			result, err := c2bWebHookResult(strings.NewReader(fmt.Sprintf(successTestBody, tc.input.merchantReqId, tc.input.resultCode, tc.input.amount, tc.input.receiptId)))
			if err != nil {
				t.Errorf("expected nil error, got %v", err)
			}

			if result.OriginationID != tc.input.merchantReqId {
				t.Errorf("expected %v, got %v", tc.input.merchantReqId, result.OriginationID)
			}

			if int(result.ResultCode) != tc.input.resultCode {
				t.Errorf("expected %v, got %v", tc.input.resultCode, result.ResultCode)
			}

			if attributes, ok := result.Attributes.(PaymentAttributes); !ok {
				t.Errorf("attributes type incorrect")
			} else {

				if attributes.Amount != strconv.Itoa(tc.input.amount) {
					t.Errorf("expected %v, got %v", tc.input.amount, attributes.Amount)
				}
			}

		}
	})

	t.Run("test failed case", func(t *testing.T) {
		failedTestBody := `{"Body":{"stkCallback":{"MerchantRequestID":"%s","CheckoutRequestID":"ws_CO_02072024204225888790902376","ResultCode":%d,"ResultDesc":"The service request is processed successfully."}}}`

		tcs := []struct {
			input testInput
		}{
			{input: testInput{resultCode: 1, merchantReqId: uuid.Must(uuid.NewV7()).String()}},
			{input: testInput{resultCode: 1032, merchantReqId: uuid.Must(uuid.NewV7()).String()}},
			{input: testInput{resultCode: 2001, merchantReqId: uuid.Must(uuid.NewV7()).String()}},
		}

		for _, tc := range tcs {
			result, err := c2bWebHookResult(strings.NewReader(fmt.Sprintf(failedTestBody, tc.input.merchantReqId, tc.input.resultCode)))
			if err != nil {
				t.Errorf("expected nil error, got %v", err)
			}

			if result.OriginationID != tc.input.merchantReqId {
				t.Errorf("expected %v, got %v", tc.input.merchantReqId, result.OriginationID)
			}

			if int(result.ResultCode) != tc.input.resultCode {
				t.Errorf("expected %v, got %v", tc.input.resultCode, result.ResultCode)
			}

			// for failed webhook requests, the attributes field should be nil
			if result.Attributes != nil {
				t.Errorf("expected nil attributes, got %v", result.Attributes)
			}
		}

	})
}

func TestB2CWebhookResult(t *testing.T) {
	//0713334691 - JACINTA MORAA RONO

	type tcInput struct {
		resultCode   int
		originatorID string
		amount       int
		receiptId    string
		receiverName string
	}

	t.Run("test success case", func(t *testing.T) {
		successTestBody := `{"Result":{"ResultType":0,"ResultCode":%d,"ResultDesc":"The service request is processed successfully.","OriginatorConversationID":"%s","ConversationID":"AG_20240703_204072a2b81ca27558c6","TransactionID":"SG31FWI0VP","ResultParameters":{"ResultParameter":[{"Key":"TransactionAmount","Value":%d},{"Key":"TransactionReceipt","Value":"%s"},{"Key":"ReceiverPartyPublicName","Value":"%s"},{"Key":"TransactionCompletedDateTime","Value":"03.07.2024 14:30:21"},{"Key":"B2CUtilityAccountAvailableFunds","Value":4905135},{"Key":"B2CWorkingAccountAvailableFunds","Value":0},{"Key":"B2CRecipientIsRegisteredCustomer","Value":"Y"},{"Key":"B2CChargesPaidAccountAvailableFunds","Value":0}]},"ReferenceData":{"ReferenceItem":{"Key":"QueueTimeoutURL","Value":"http://internalapi.safaricom.co.ke/mpesa/b2cresults/v1/submit"}}}}`

		tcs := []struct {
			input tcInput
		}{
			{input: tcInput{resultCode: 0, originatorID: uuid.Must(uuid.NewV7()).String(), amount: 1, receiptId: ulid.Make().String(), receiverName: "0712345678 - JOHN DOE"}},
			{input: tcInput{resultCode: 0, originatorID: uuid.Must(uuid.NewV7()).String(), amount: 10, receiptId: ulid.Make().String(), receiverName: "0712345678 - JOHN DOE"}},
			{input: tcInput{resultCode: 0, originatorID: uuid.Must(uuid.NewV7()).String(), amount: 100, receiptId: ulid.Make().String(), receiverName: "0712345678 - JOHN DOE"}},
			{input: tcInput{resultCode: 0, originatorID: uuid.Must(uuid.NewV7()).String(), amount: 1000, receiptId: ulid.Make().String(), receiverName: "0712345678 - JOHN DOE"}},
			{input: tcInput{resultCode: 0, originatorID: uuid.Must(uuid.NewV7()).String(), amount: 10000, receiptId: ulid.Make().String(), receiverName: "0712345678 - JOHN DOE"}},
			{input: tcInput{resultCode: 0, originatorID: uuid.Must(uuid.NewV7()).String(), amount: 50000, receiptId: ulid.Make().String(), receiverName: "0712345678 - JOHN DOE"}},
			{input: tcInput{resultCode: 0, originatorID: uuid.Must(uuid.NewV7()).String(), amount: 70000, receiptId: ulid.Make().String(), receiverName: "0712345678 - JOHN DOE"}},
			{input: tcInput{resultCode: 0, originatorID: uuid.Must(uuid.NewV7()).String(), amount: 100000, receiptId: ulid.Make().String(), receiverName: "0712345678 - JOHN DOE"}},
			{input: tcInput{resultCode: 0, originatorID: uuid.Must(uuid.NewV7()).String(), amount: 150000, receiptId: ulid.Make().String(), receiverName: "0712345678 - JOHN DOE"}},
		}

		for _, tc := range tcs {
			result, err := b2cWebhookResult(strings.NewReader(fmt.Sprintf(successTestBody, tc.input.resultCode, tc.input.originatorID, tc.input.amount, tc.input.receiptId, tc.input.receiverName)))
			if err != nil {
				t.Errorf("expected  nil error, got %v", err)
			}

			if result.OriginationID != tc.input.originatorID {
				t.Errorf("expected %s, got %v", tc.input.originatorID, result.OriginationID)
			}

			if attributes, ok := result.Attributes.(PaymentAttributes); !ok {
				t.Errorf("attributes property type incorrect")
			} else {

				// check amount is valid
				amount, err := strconv.ParseFloat(attributes.Amount, 64)
				if err != nil {
					t.Errorf("expected nil error, got %v", err)
				}

				// don't like the float64 cast of expected value
				if amount != float64(tc.input.amount) {
					t.Errorf("expected %d, got %v", tc.input.amount, attributes.Amount)
				}

				if attributes.MpesaReceiptID != tc.input.receiptId {
					t.Errorf("expected %s, got %s", tc.input.receiptId, attributes.MpesaReceiptID)
				}
			}
		}
	})

	t.Run("test failed case", func(t *testing.T) {
		failedTestBody := `{"Result":{"ResultType":0,"ResultCode":%d,"ResultDesc":"The service request is processed successfully.","OriginatorConversationID":"%s","ConversationID":"AG_20240703_204072a2b81ca27558c6","TransactionID":"SG31FWI0VP","ReferenceData":{"ReferenceItem":{"Key":"QueueTimeoutURL","Value":"http://internalapi.safaricom.co.ke/mpesa/b2cresults/v1/submit"}}}}`

		tcs := []struct {
			input tcInput
		}{
			{input: tcInput{resultCode: 1, originatorID: uuid.Must(uuid.NewV7()).String()}},
			{input: tcInput{resultCode: 1032, originatorID: uuid.Must(uuid.NewV7()).String()}},
			{input: tcInput{resultCode: 2001, originatorID: uuid.Must(uuid.NewV7()).String()}},
			{input: tcInput{resultCode: 1037, originatorID: uuid.Must(uuid.NewV7()).String()}},
		}

		for _, tc := range tcs {
			result, err := b2cWebhookResult(strings.NewReader(fmt.Sprintf(failedTestBody, tc.input.resultCode, tc.input.originatorID)))
			if err != nil {
				t.Errorf("expected nil error, got %v", err)
			}

			if result.OriginationID != tc.input.originatorID {
				t.Errorf("expected %s, got %v", tc.input.originatorID, result.OriginationID)
			}

			if result.Attributes != nil {
				t.Errorf("expected attributes property to be nil, got %v", result.Attributes)
			}
		}
	})
}

func TestTransactionStatusWebhookResult(t *testing.T) {
	type tcInput struct {
		resultCode      int
		originatorID    string
		debitPartyName  string
		status          string
		amount          int
		receiptID       string
		transactionDate int
	}

	t.Run("test success case", func(t *testing.T) {
		successTestBody := `{"Result":{"ResultType":0,"ResultCode":%d,"ResultDesc":"The service request is processed successfully.","OriginatorConversationID":"9f6c92024b39880271","ConversationID":"AG_20240702_20303738b089a9fb5233","TransactionID":"SG20000000","ResultParameters":{"ResultParameter":[{"Key":"DebitPartyName","Value":"%s"},{"Key":"CreditPartyName","Value":"490742 - FINGO  AFRICA BY ECOBANK"},{"Key":"OriginatorConversationID","Value":"%s"},{"Key":"InitiatedTime","Value":%d},{"Key":"CreditPartyCharges"},{"Key":"DebitAccountType","Value":"MMF Account For Customer"},{"Key":"TransactionReason"},{"Key":"ReasonType","Value":"Pay Bill Online"},{"Key":"TransactionStatus","Value":"%s"},{"Key":"FinalisedTime","Value":20240629184302},{"Key":"Amount","Value":%d},{"Key":"ConversationID","Value":"AG_20240629_2040165691b903e92930"},{"Key":"ReceiptNo","Value":"%s"}]},"ReferenceData":{"ReferenceItem":{"Key":"Occasion","Value":"OK"}}}}`

		tcs := []struct {
			input tcInput
		}{
			{input: tcInput{resultCode: 0, originatorID: uuid.Must(uuid.NewV7()).String(), debitPartyName: "JOHN DOE", status: "Completed", amount: 1, receiptID: ulid.Make().String(), transactionDate: 20240629184302}},
			{input: tcInput{resultCode: 0, originatorID: uuid.Must(uuid.NewV7()).String(), debitPartyName: "JOHN DOE", status: "Completed", amount: 10, receiptID: ulid.Make().String(), transactionDate: 20240629384302}},
			{input: tcInput{resultCode: 0, originatorID: uuid.Must(uuid.NewV7()).String(), debitPartyName: "JOHN DOE", status: "Completed", amount: 100, receiptID: ulid.Make().String(), transactionDate: 20240629384302}},
			{input: tcInput{resultCode: 0, originatorID: uuid.Must(uuid.NewV7()).String(), debitPartyName: "JOHN DOE", status: "Completed", amount: 1000, receiptID: ulid.Make().String(), transactionDate: 20240629384302}},
			{input: tcInput{resultCode: 0, originatorID: uuid.Must(uuid.NewV7()).String(), debitPartyName: "JOHN DOE", status: "Completed", amount: 10000, receiptID: ulid.Make().String(), transactionDate: 20240629384302}},
			{input: tcInput{resultCode: 0, originatorID: uuid.Must(uuid.NewV7()).String(), debitPartyName: "JOHN DOE", status: "Completed", amount: 100000, receiptID: ulid.Make().String(), transactionDate: 20240629384302}},
			{input: tcInput{resultCode: 0, originatorID: uuid.Must(uuid.NewV7()).String(), debitPartyName: "JOHN DOE", status: "Completed", amount: 1000000, receiptID: ulid.Make().String(), transactionDate: 20240629384302}},
		}

		for _, tc := range tcs {
			result, err := transactionStatusWebhookResult(strings.NewReader(fmt.Sprintf(successTestBody, tc.input.resultCode, tc.input.debitPartyName, tc.input.originatorID, tc.input.transactionDate, tc.input.status, tc.input.amount, tc.input.receiptID)))
			if err != nil {
				t.Errorf("expected nil error, got %v", err)
			}

			// result status should always be completed for success case
			if result.Status != "completed" {
				t.Errorf("expected status 'completed', got %s", result.Status)
			}

			if result.OriginationID != tc.input.originatorID {
				t.Errorf("expected %s, got %s", tc.input.originatorID, result.OriginationID)
			}

			var attributes PaymentAttributes
			var ok bool
			if attributes, ok = result.Attributes.(PaymentAttributes); !ok {
				t.Errorf("attributes property type incorrect")
			}

			if attributes.MpesaReceiptID != tc.input.receiptID {
				t.Errorf("expected %s, got %s", tc.input.receiptID, attributes.MpesaReceiptID)
			}

			amount, err := strconv.ParseFloat(attributes.Amount, 64)
			if err != nil {
				t.Errorf("expected nil error, got %v", err)
			}

			if amount != float64(tc.input.amount) {
				t.Errorf("expected %v, got %v", tc.input.amount, amount)
			}

			if attributes.SenderName != tc.input.debitPartyName {
				t.Errorf("expected %v, got %v", tc.input.debitPartyName, attributes.SenderName)
			}

			transactionDate := fmt.Sprintf("%v", tc.input.transactionDate)
			if attributes.TransactionDate != transactionDate {
				t.Errorf("expected %v, got %v", transactionDate, attributes.TransactionDate)
			}

			if result.OriginationID != tc.input.originatorID {
				t.Errorf("expected %s, got %s", tc.input.originatorID, result.OriginationID)
			}
		}
	})

	t.Run("test failed case", func(t *testing.T) {
		failedTestBody := `{"Result":{"ResultType":0,"ResultCode":%d,"ResultDesc":"The service request is processed successfully.","OriginatorConversationID":"e3b6-440d-b59f-9f6c92024b39880271","ConversationID":"AG_20240702_20303738b089a9fb5233","TransactionID":"SG20000000","ReferenceData":{"ReferenceItem":{"Key":"Occasion","Value":"OK"}}}}`

		tcs := []struct {
			input tcInput
		}{
			{input: tcInput{resultCode: 1}},
			{input: tcInput{resultCode: 1032}},
			{input: tcInput{resultCode: 2001}},
		}

		for _, tc := range tcs {
			result, err := transactionStatusWebhookResult(strings.NewReader(fmt.Sprintf(failedTestBody, tc.input.resultCode)))
			if err != nil {
				t.Errorf("expected nil error, got %v", err)
			}

			// if result type is non-success, status should not be set
			if result.Status != "" {
				t.Errorf("expected empty value, got %s", result.Status)
			}

			if result.Attributes != nil {
				t.Errorf("expected nil value, got %v", result.Attributes)
			}
		}

	})
}
