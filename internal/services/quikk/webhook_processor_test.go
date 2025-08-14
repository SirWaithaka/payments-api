package quikk

import (
	"fmt"
	"strings"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/oklog/ulid/v2"

	"github.com/SirWaithaka/payments-api/clients/quikk"
	"github.com/SirWaithaka/payments-api/internal/domains/mpesa"
	"github.com/SirWaithaka/payments-api/internal/domains/requests"
	"github.com/SirWaithaka/payments-api/internal/pkg/types"
)

func TestWebhookProcessor_Process(t *testing.T) {
	chargeSuccessBody := `{"data":{"type":"payin","id":"d2d3gu6q54te0s4npr20","attributes":{"sender_type":"msisdn","txn_charge_id":"ws_CO_110820252137464713011722","amount":1,"txn_id":"%s","txn_charge_created_at":"2025-08-11T21:38:03+0300","sender_no":"254713011722","sender_no_sha1":"9639950ec44cdd048d18cba786437ae924fac397","sender_no_sha256":"c641052890dd119252d3cbdefc1e21f33d410f20f394c0e8ae2d785c3f5c7703"}}}`
	chargeFailedBody := `{"data":{"type":"payin","id":"d2d4efmq54t372vkvo40","attributes":{"txn_charge_id":"ws_CO_110820252240476713011722"}},"meta":{"status":"FAIL","code":"1037","detail":"No response from user"}}`
	payoutSuccessBody := `{"data":{"type":"payout","id":"1","attributes":{"txn_id":"%s","response_id":"AG_20190905_00005f0dcb86732c611c","recipient_type":"msisdn","recipient_no":"2547*****024","recipient_no_sha1":"a8150b7b38fa70ae1c0e4b5312f3a79827ceb657","recipient_no_sha256":"bc89782d032e6f28356888aba50dc35a1c180078eae95465c13ab1b82b8492b9","recipient_registered":"Y","amount":10,"txn_created_at":"2022-07-01T10:51:59+0300","balance_utility_ac":4517632.27,"balance_working_ac":299825045,"balance_charges_paid_ac":0}}}`
	payoutFailedBody := `{"data":{"type":"payout","id":"1","attributes":{"txn_id":"NH90HBCXPM","response_id":"AG_20190809_000040b4caf4c7a029c0"}},"meta":{"status":"FAIL","code":"17","detail":"The initiator is not allowed to initiate this request"}}`
	transferSuccessBody := `{"data":{"type":"transfer","id":"1","attributes":{"txn_id":"%s","response_id":"AG_20190809_000040b4caf4c7a029c0","recipient_type":"short_code","recipient_no":"12348","recipient_no_sha1":"ad8f96d2cbd19db18de8e2cc70703ba9c3c6840b","recipient_no_sha256":"015e81eddfab44be16ac53a8653feab50859b4c5508a915679e33c271d2b54df","recipient_name":"some company","short_code":"123456","short_code_name":"thingamagik","amount":10,"txn_created_at":"2022-07-01T10:51:59+0300","balance_working_ac":111745,"balance_utility_ac":111745,"balance_charges_paid_ac":111745,"reference":"7000"}}}`
	transferFailedBody := `{"data":{"type":"transfer","id":"1","attributes":{"txn_id":"NH90HBCXPM","response_id":"AG_20190809_000040b4caf4c7a029c0"}},"meta":{"status":"FAIL","code":"20","detail":"Insufficient balance"}}`
	transactionSearchSuccessBody := `{"data":{"type":"search","id":"6e9a2aad-621e-444c-9fc8-1e08fdaaaa6c","attributes":{"resource_id":"1","response_id":"AG_20190417_000049de14ae0c48","txn_id":"%s","amount":1222.22,"recipient_fee":33,"recipient_name":"Safaricom Disbursement Account","recipient_no":"511382","recipient_type":"short_code","sender_name":"Jane J D","sender_no":"2547*****678","sender_no_sha1":"90fbb07a296a87f476af720ccb89f35822e50182","sender_no_sha256":"7132104d6aae9c3fac82095a42c2817952bca48e09d98d5bf4ac08218982fb90","sender_fee":"33.0,","txn_type":"payin","category":"Paybill","txn_status":"Authorized","txn_created_at":"2022-07-01T10:51:59+0300"}}}`
	balanceSearchSuccessBody := `{"data":{"type":"search","id":"1","attributes":{"response_id":"AG_20190808_000051f18a81f3aee279","txn_id":"NH94HBCXII","balance_working_ac":4761531.1,"balance_utility_ac":4761531,"balance_charges_paid_ac":4761531,"balance_merchant_ac":4761531,"balance_organization_settlement_ac":4761531,"checked_at":"2019-03-18T17:22:09.651011Z"}}}`

	paymentRef := ulid.Make().String()

	testcases := []struct {
		name     string
		input    *requests.WebhookResult
		expected mpesa.OptionsUpdatePayment
	}{
		{
			name:     "test a successful quikk mpesa charge webhook",
			input:    requests.NewWebhookResult("test", quikk.OperationCharge, strings.NewReader(fmt.Sprintf(chargeSuccessBody, paymentRef))),
			expected: mpesa.OptionsUpdatePayment{PaymentReference: &paymentRef, Status: types.Pointer(requests.StatusSucceeded)},
		},
		{
			name:     "test a failed quikk mpesa charge webhook",
			input:    requests.NewWebhookResult("test", quikk.OperationCharge, strings.NewReader(chargeFailedBody)),
			expected: mpesa.OptionsUpdatePayment{Status: types.Pointer(requests.StatusFailed)},
		},
		{
			name:     "test a successful quikk mpesa payout webhook",
			input:    requests.NewWebhookResult("test", quikk.OperationPayout, strings.NewReader(fmt.Sprintf(payoutSuccessBody, paymentRef))),
			expected: mpesa.OptionsUpdatePayment{PaymentReference: &paymentRef, Status: types.Pointer(requests.StatusSucceeded)},
		},
		{
			name:     "test a failed quikk mpesa payout webhook",
			input:    requests.NewWebhookResult("test", quikk.OperationPayout, strings.NewReader(payoutFailedBody)),
			expected: mpesa.OptionsUpdatePayment{Status: types.Pointer(requests.StatusFailed)},
		},
		{
			name:     "test a successful quikk mpesa transfer webhook",
			input:    requests.NewWebhookResult("test", quikk.OperationTransfer, strings.NewReader(fmt.Sprintf(transferSuccessBody, paymentRef))),
			expected: mpesa.OptionsUpdatePayment{PaymentReference: &paymentRef, Status: types.Pointer(requests.StatusSucceeded)},
		},
		{
			name:     "test a failed quikk mpesa transfer webhook",
			input:    requests.NewWebhookResult("test", quikk.OperationTransfer, strings.NewReader(transferFailedBody)),
			expected: mpesa.OptionsUpdatePayment{Status: types.Pointer(requests.StatusFailed)},
		},
		{
			name:     "test a successful transaction status webhook",
			input:    requests.NewWebhookResult("test", quikk.OperationSearch, strings.NewReader(fmt.Sprintf(transactionSearchSuccessBody, paymentRef))),
			expected: mpesa.OptionsUpdatePayment{Status: types.Pointer(requests.StatusSucceeded), PaymentReference: &paymentRef},
		},
		{ // the processor should ignore this webhook and not return an error
			name:     "test a successful balance search webhook",
			input:    requests.NewWebhookResult("test", quikk.OperationSearch, strings.NewReader(balanceSearchSuccessBody)),
			expected: mpesa.OptionsUpdatePayment{},
		},
	}

	processor := NewWebhookProcessor()

	for _, tc := range testcases {

		t.Run(tc.name, func(t *testing.T) {
			opts := mpesa.OptionsUpdatePayment{}
			err := processor.Process(t.Context(), tc.input, &opts)
			if err != nil {
				t.Errorf("expected nil error, got %v", err)
			}

			assert.Equal(t, tc.expected, opts)
		})
	}
}
