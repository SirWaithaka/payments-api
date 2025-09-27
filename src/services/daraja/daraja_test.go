package daraja_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	jsoniter "github.com/json-iterator/go"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"

	"github.com/SirWaithaka/payments-api/src/domains/mpesa"
	"github.com/SirWaithaka/payments-api/src/domains/requests"
	"github.com/SirWaithaka/payments-api/src/repositories/postgres"
	"github.com/SirWaithaka/payments-api/src/services/daraja"
	"github.com/SirWaithaka/payments-api/testdata"
	daraja_sdk "github.com/SirWaithaka/payments/daraja"
)

const (
	key    = "fake_key"
	secret = "fake_secret"
)

func TestDarajaApi_C2B(t *testing.T) {

	testPayment := mpesa.PaymentRequest{
		IdempotencyID:         ulid.Make().String(),
		ClientTransactionID:   ulid.Make().String(),
		Amount:                "15",
		ExternalAccountNumber: "254712345678",
		Beneficiary:           "100200",
		Description:           "test payment",
	}
	shortcode := mpesa.ShortCode{
		ShortCode:         "900999",
		InitiatorName:     "test_name",
		InitiatorPassword: "test_password",
		Passphrase:        "test_passphrase",
		Key:               key,
		Secret:            secret,
	}

	repository := postgres.NewRequestRepository(inf.Storage.PG)

	t.Run("test that request is sent and saved", func(t *testing.T) {
		defer testdata.ResetTables(inf)

		merchantReqID := ulid.Make().String()
		checkoutReqID := ulid.Make().String()
		// create a mock test server
		mux := http.NewServeMux()
		mux.HandleFunc(daraja_sdk.EndpointC2bExpress, func(w http.ResponseWriter, r *http.Request) {
			// parse request body
			var req daraja_sdk.RequestC2BExpress
			if err := jsoniter.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Errorf("expected nil error, got %v", err)
			}
			// assert request values
			assert.Equal(t, shortcode.ShortCode, req.BusinessShortCode)
			assert.Equal(t, daraja_sdk.TypeCustomerPayBillOnline, req.TransactionType)
			assert.Equal(t, testPayment.Amount, req.Amount)
			assert.Equal(t, testPayment.ExternalAccountNumber, req.PartyA)
			assert.Equal(t, shortcode.ShortCode, req.PartyB)

			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(fmt.Sprintf(`{"ResponseMessage":"Success","ResponseCode":"0","MerchantRequestID":"%s","CheckoutRequestID":"%s"}`, merchantReqID, checkoutReqID)))
		})
		server := httptest.NewServer(mux)
		defer server.Close()

		// build daraja client
		client := daraja_sdk.New(daraja_sdk.Config{Endpoint: server.URL})
		// create instance of daraja service
		service := daraja.NewDarajaApi(&client, daraja_sdk.SandboxCertificate, shortcode, repository)

		// make request
		paymentID := ulid.Make().String()
		err := service.C2B(t.Context(), paymentID, testPayment)
		// expect no error
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		// check request is saved
		request, err := repository.FindOne(t.Context(), requests.OptionsFindRequest{ExternalID: &merchantReqID})
		// expect no error
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		// assert request values
		assert.Equal(t, paymentID, request.PaymentID)
		// status should be succeeded for successful requests
		assert.Equal(t, requests.StatusSucceeded, request.Status)
	})

	t.Run("test request fail", func(t *testing.T) {
		defer testdata.ResetTables(inf)

		requestID := ulid.Make().String()
		// create a mock test server
		mux := http.NewServeMux()
		mux.HandleFunc(daraja_sdk.EndpointC2bExpress, func(w http.ResponseWriter, r *http.Request) {
			// mock request failure
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(fmt.Sprintf(`{"requestId":"%s","errorCode":"23","errorMessage":"test failure"}`, requestID)))
		})
		server := httptest.NewServer(mux)
		defer server.Close()

		// build daraja client
		client := daraja_sdk.New(daraja_sdk.Config{Endpoint: server.URL})
		// create instance of daraja service
		service := daraja.NewDarajaApi(&client, daraja_sdk.SandboxCertificate, shortcode, repository)
		// make request
		paymentID := ulid.Make().String()
		err := service.C2B(t.Context(), paymentID, testPayment)
		if err == nil {
			t.Errorf("expected error, got nil")
		}

		// check request is saved
		var record postgres.RequestSchema
		result := inf.Storage.PG.First(&record)
		// expect no error
		if err = result.Error; err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		// assert request values
		assert.Equal(t, paymentID, *record.PaymentID)
		// status should be succeeded for successful requests
		assert.Equal(t, requests.StatusFailed, requests.ToStatus(*record.Status))
	})

}

func TestDarajaApi_B2C(t *testing.T) {

	testPayment := mpesa.PaymentRequest{
		IdempotencyID:         ulid.Make().String(),
		ClientTransactionID:   ulid.Make().String(),
		Amount:                "15",
		ExternalAccountNumber: "254712345678",
		Beneficiary:           "100200",
		Description:           "test payment",
	}
	shortcode := mpesa.ShortCode{
		ShortCode:         "900999",
		InitiatorName:     "test_name",
		InitiatorPassword: "test_password",
		Passphrase:        "test_passphrase",
		Key:               key,
		Secret:            secret,
	}

	repository := postgres.NewRequestRepository(inf.Storage.PG)

	t.Run("test that request is sent and saved", func(t *testing.T) {
		defer testdata.ResetTables(inf)

		conversationID := ulid.Make().String()
		originatorConversationID := ulid.Make().String()

		// create a mock test server
		mux := http.NewServeMux()
		mux.HandleFunc(daraja_sdk.EndpointB2cPayment, func(w http.ResponseWriter, r *http.Request) {
			// parse request body
			var req daraja_sdk.RequestB2C
			if err := jsoniter.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Errorf("expected nil error, got %v", err)
			}
			// assert request values
			assert.Equal(t, testPayment.ClientTransactionID, req.OriginatorConversationID)
			assert.Equal(t, shortcode.InitiatorName, req.InitiatorName)
			assert.Equal(t, daraja_sdk.CommandBusinessPayment, req.CommandID)
			assert.Equal(t, testPayment.Amount, req.Amount)
			assert.Equal(t, shortcode.ShortCode, req.PartyA)
			assert.Equal(t, testPayment.ExternalAccountNumber, req.PartyB)

			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(fmt.Sprintf(`{"ResponseDescription":"Success","ResponseCode":"0","ConversationID":"%s","OriginatorConversationID":"%s"}`, conversationID, originatorConversationID)))
		})
		server := httptest.NewServer(mux)
		defer server.Close()

		// build daraja client
		client := daraja_sdk.New(daraja_sdk.Config{Endpoint: server.URL})
		// create instance of daraja service
		service := daraja.NewDarajaApi(&client, daraja_sdk.SandboxCertificate, shortcode, repository)
		// make request
		paymentID := ulid.Make().String()
		err := service.B2C(t.Context(), paymentID, testPayment)
		// expect no error
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		// check request is saved
		request, err := repository.FindOne(t.Context(), requests.OptionsFindRequest{ExternalID: &originatorConversationID})
		// expect no error
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		// assert request values
		assert.Equal(t, paymentID, request.PaymentID)
		// status should be succeeded for successful requests
		assert.Equal(t, requests.StatusSucceeded, request.Status)

	})

	t.Run("test request fail", func(t *testing.T) {
		defer testdata.ResetTables(inf)

		requestID := ulid.Make().String()
		// create a mock test server
		mux := http.NewServeMux()
		mux.HandleFunc(daraja_sdk.EndpointB2cPayment, func(w http.ResponseWriter, r *http.Request) {
			// mock request failure
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(fmt.Sprintf(`{"requestId":"%s","errorCode":"23","errorMessage":"test failure"}`, requestID)))
		})
		server := httptest.NewServer(mux)
		defer server.Close()

		// build daraja client
		client := daraja_sdk.New(daraja_sdk.Config{Endpoint: server.URL})
		// create instance of daraja service
		service := daraja.NewDarajaApi(&client, daraja_sdk.SandboxCertificate, shortcode, repository)
		// make request
		paymentID := ulid.Make().String()
		err := service.B2C(t.Context(), paymentID, testPayment)
		if err == nil {
			t.Errorf("expected error, got nil")
		}

		// check request is saved
		var record postgres.RequestSchema
		result := inf.Storage.PG.First(&record)
		// expect no error
		if err = result.Error; err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		// assert request values
		assert.Equal(t, paymentID, *record.PaymentID)
		// status should be succeeded for successful requests
		assert.Equal(t, requests.StatusFailed, requests.ToStatus(*record.Status))

	})
}

func TestDarajaApi_B2B(t *testing.T) {
	testPayment := mpesa.PaymentRequest{
		IdempotencyID:         ulid.Make().String(),
		ClientTransactionID:   ulid.Make().String(),
		Amount:                "15",
		ExternalAccountNumber: "254712345678",
		Beneficiary:           "100200",
		Description:           "test payment",
	}
	shortcode := mpesa.ShortCode{
		ShortCode:         "900999",
		InitiatorName:     "test_name",
		InitiatorPassword: "test_password",
		Passphrase:        "test_passphrase",
		Key:               key,
		Secret:            secret,
	}

	repository := postgres.NewRequestRepository(inf.Storage.PG)

	t.Run("test that paybill payment request is sent and saved", func(t *testing.T) {
		defer testdata.ResetTables(inf)

		conversationID := ulid.Make().String()
		originatorConversationID := ulid.Make().String()

		// create a mock test server
		mux := http.NewServeMux()
		mux.HandleFunc(daraja_sdk.EndpointB2bPayment, func(w http.ResponseWriter, r *http.Request) {
			// parse request body
			var req daraja_sdk.RequestB2B
			if err := jsoniter.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Errorf("expected nil error, got %v", err)
			}
			// assert request values
			assert.Equal(t, shortcode.InitiatorName, req.Initiator)
			assert.Equal(t, daraja_sdk.CommandBusinessPayBill, req.CommandID)
			assert.Equal(t, testPayment.Amount, req.Amount)
			assert.Equal(t, shortcode.ShortCode, req.PartyA)
			assert.Equal(t, testPayment.ExternalAccountNumber, req.PartyB)
			assert.Equal(t, testPayment.Beneficiary, req.AccountReference)

			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(fmt.Sprintf(`{"ResponseDescription":"Success","ResponseCode":"0","ConversationID":"%s","OriginatorConversationID":"%s"}`, conversationID, originatorConversationID)))
		})
		server := httptest.NewServer(mux)
		defer server.Close()

		// build daraja client
		client := daraja_sdk.New(daraja_sdk.Config{Endpoint: server.URL})
		// create instance of daraja service
		service := daraja.NewDarajaApi(&client, daraja_sdk.SandboxCertificate, shortcode, repository)
		// make request
		paymentID := ulid.Make().String()
		err := service.B2B(t.Context(), paymentID, testPayment)
		// expect no error
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		// check request is saved
		request, err := repository.FindOne(t.Context(), requests.OptionsFindRequest{ExternalID: &originatorConversationID})
		// expect no error
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		// assert request values
		assert.Equal(t, paymentID, request.PaymentID)
		// status should be succeeded for successful requests
		assert.Equal(t, requests.StatusSucceeded, request.Status)
	})

	t.Run("test that till payment request is sent and saved", func(t *testing.T) {
		defer testdata.ResetTables(inf)

		conversationID := ulid.Make().String()
		originatorConversationID := ulid.Make().String()

		// create a mock test server
		mux := http.NewServeMux()
		mux.HandleFunc(daraja_sdk.EndpointB2bPayment, func(w http.ResponseWriter, r *http.Request) {
			// parse request body
			var req daraja_sdk.RequestB2B
			if err := jsoniter.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Errorf("expected nil error, got %v", err)
			}
			// assert request values
			assert.Equal(t, shortcode.InitiatorName, req.Initiator)
			assert.Equal(t, daraja_sdk.CommandBusinessBuyGoods, req.CommandID)
			assert.Equal(t, testPayment.Amount, req.Amount)
			assert.Equal(t, shortcode.ShortCode, req.PartyA)
			assert.Equal(t, testPayment.ExternalAccountNumber, req.PartyB)
			assert.Equal(t, testPayment.Beneficiary, req.AccountReference)

			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(fmt.Sprintf(`{"ResponseDescription":"Success","ResponseCode":"0","ConversationID":"%s","OriginatorConversationID":"%s"}`, conversationID, originatorConversationID)))
		})
		server := httptest.NewServer(mux)
		defer server.Close()

		// build daraja client
		client := daraja_sdk.New(daraja_sdk.Config{Endpoint: server.URL})
		// create instance of daraja service
		service := daraja.NewDarajaApi(&client, daraja_sdk.SandboxCertificate, shortcode, repository)
		// make request
		testPayment.ExternalAccountType = mpesa.AccountTypeTill
		paymentID := ulid.Make().String()
		err := service.B2B(t.Context(), paymentID, testPayment)
		// expect no error
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		// check request is saved
		request, err := repository.FindOne(t.Context(), requests.OptionsFindRequest{ExternalID: &originatorConversationID})
		// expect no error
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		// assert request values
		assert.Equal(t, paymentID, request.PaymentID)
		// status should be succeeded for successful requests
		assert.Equal(t, requests.StatusSucceeded, request.Status)
	})

	t.Run("test request fail", func(t *testing.T) {
		defer testdata.ResetTables(inf)

		requestID := ulid.Make().String()
		// create a mock test server
		mux := http.NewServeMux()
		mux.HandleFunc(daraja_sdk.EndpointB2bPayment, func(w http.ResponseWriter, r *http.Request) {
			// mock request failure
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(fmt.Sprintf(`{"requestId":"%s","errorCode":"23","errorMessage":"test failure"}`, requestID)))
		})
		server := httptest.NewServer(mux)
		defer server.Close()

		// build daraja client
		client := daraja_sdk.New(daraja_sdk.Config{Endpoint: server.URL})
		// create instance of daraja service
		service := daraja.NewDarajaApi(&client, daraja_sdk.SandboxCertificate, shortcode, repository)
		// make request
		paymentID := ulid.Make().String()
		err := service.B2B(t.Context(), paymentID, testPayment)
		// expect error
		if err == nil {
			t.Errorf("expected nil error, got %v", err)
		}

		// check request is saved
		var record postgres.RequestSchema
		result := inf.Storage.PG.First(&record)
		// expect no error
		if err = result.Error; err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		// assert request values
		assert.Equal(t, paymentID, *record.PaymentID)
		// status should be succeeded for successful requests
		assert.Equal(t, requests.StatusFailed, requests.ToStatus(*record.Status))

	})
}

func TestDarajaApi_Status(t *testing.T) {
	shortcode := mpesa.ShortCode{
		ShortCode:         "900999",
		InitiatorName:     "test_name",
		InitiatorPassword: "test_password",
		Passphrase:        "test_passphrase",
		Key:               key,
		Secret:            secret,
	}

	repository := postgres.NewRequestRepository(inf.Storage.PG)

	conversationID := ulid.Make().String()
	originatorConversationID := ulid.Make().String()
	// create a mock test server
	mux := http.NewServeMux()
	mux.HandleFunc(daraja_sdk.EndpointTransactionStatus, func(w http.ResponseWriter, r *http.Request) {
		// mock request failure
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(fmt.Sprintf(`{"ResponseDescription":"Success","ResponseCode":"0","ConversationID":"%s","OriginatorConversationID":"%s"}`, conversationID, originatorConversationID)))

	})
	server := httptest.NewServer(mux)
	defer server.Close()

	// build daraja client
	client := daraja_sdk.New(daraja_sdk.Config{Endpoint: server.URL})
	// create instance of daraja service
	service := daraja.NewDarajaApi(&client, daraja_sdk.SandboxCertificate, shortcode, repository)

	t.Run("test transaction status parameters", func(t *testing.T) {
		testcases := []struct {
			name  string
			input mpesa.Payment
		}{
			{name: "test on payment reference", input: mpesa.Payment{PaymentReference: ulid.Make().String()}},
			{name: "test on client transaction id", input: mpesa.Payment{ClientTransactionID: ulid.Make().String()}},
		}

		for _, tc := range testcases {
			t.Run(tc.name, func(t *testing.T) {
				err := service.Status(t.Context(), tc.input)
				if err != nil {
					t.Errorf("expected nil error, got %v", err)
				}

			})
		}

	})

}
