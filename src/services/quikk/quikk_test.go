package quikk_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"

	"github.com/SirWaithaka/payments-api/src/domains/mpesa"
	"github.com/SirWaithaka/payments-api/src/domains/requests"
	"github.com/SirWaithaka/payments-api/src/repositories/postgres"
	"github.com/SirWaithaka/payments-api/src/services/quikk"
	"github.com/SirWaithaka/payments-api/testdata"
	quikk2 "github.com/SirWaithaka/payments/quikk"
)

const (
	key    = "fake_key"
	secret = "fake_secret"
)

var (
	shortcode = mpesa.ShortCode{
		ShortCode: "800888",
		Key:       key,
		Secret:    secret,
	}
)

func TestQuikkApi_C2B(t *testing.T) {

	testPayment := mpesa.PaymentRequest{
		IdempotencyID:         ulid.Make().String(),
		ClientTransactionID:   ulid.Make().String(),
		Amount:                "105",
		ExternalAccountNumber: "254712345678",
		Description:           "test payment",
	}

	repository := postgres.NewRequestRepository(inf.Storage.PG)

	t.Run("test that request is sent and saved", func(t *testing.T) {
		requestID := ulid.Make().String()
		resourceID := ulid.Make().String()

		// create a mock test server
		mux := http.NewServeMux()
		mux.HandleFunc(quikk2.EndpointCharge, func(w http.ResponseWriter, r *http.Request) {
			// parse request body
			var req quikk2.RequestDefault[quikk2.RequestCharge]
			if err := jsoniter.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Errorf("expected nil error, got %v", err)
			}

			assert.Equal(t, "charge", req.Data.Type)
			// TODO: write better test case for checking amount conversions from string to float64
			//assert.EqualValues(t, testPayment.Amount, fmt.Sprintf("%f", req.Data.Attributes.Amount))
			assert.Equal(t, testPayment.ExternalAccountNumber, req.Data.Attributes.CustomerNo)
			assert.Equal(t, shortcode.ShortCode, req.Data.Attributes.ShortCode)
			assert.Equal(t, mpesa.AccountTypeMSISDN.String(), req.Data.Attributes.CustomerType)

			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(fmt.Sprintf(`{"data":{"id":"%s","type":"payin","attributes":{"resource_id":"%s"}}}`, requestID, resourceID)))
		})
		server := httptest.NewServer(mux)
		defer server.Close()

		client := quikk2.New(quikk2.Config{Endpoint: server.URL})
		// create instance of quikk service
		service := quikk.NewQuikkApi(&client, shortcode, repository)
		// make request
		paymentID := ulid.Make().String()
		err := service.C2B(t.Context(), paymentID, testPayment)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		// check request is saved
		request, err := repository.FindOne(t.Context(), requests.OptionsFindRequest{ExternalID: &requestID})
		// expect no error
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		// assert request values
		assert.Equal(t, paymentID, request.PaymentID)
		// status should be succeeded for successful requests
		assert.Equal(t, requests.StatusSucceeded, request.Status)

	})
}

func TestQuikkApi_B2C(t *testing.T) {
	testPayment := mpesa.PaymentRequest{
		IdempotencyID:         ulid.Make().String(),
		ClientTransactionID:   ulid.Make().String(),
		Amount:                "105",
		ExternalAccountNumber: "254712345678",
		Description:           "test payment",
	}

	repository := postgres.NewRequestRepository(inf.Storage.PG)

	t.Run("test that request is sent and saved", func(t *testing.T) {
		requestID := ulid.Make().String()
		resourceID := ulid.Make().String()

		// create a mock test server
		mux := http.NewServeMux()
		mux.HandleFunc(quikk2.EndpointPayout, func(w http.ResponseWriter, r *http.Request) {
			// parse request body
			var req quikk2.RequestDefault[quikk2.RequestPayout]
			if err := jsoniter.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Errorf("expected nil error, got %v", err)
			}

			assert.Equal(t, "payout", req.Data.Type)
			// TODO: write better test case for checking amount conversions from string to float64
			//assert.EqualValues(t, testPayment.Amount, fmt.Sprintf("%f", req.Data.Attributes.Amount))
			assert.Equal(t, testPayment.ExternalAccountNumber, req.Data.Attributes.RecipientNo)
			assert.Equal(t, mpesa.AccountTypeMSISDN.String(), req.Data.Attributes.RecipientType)
			assert.Equal(t, shortcode.ShortCode, req.Data.Attributes.ShortCode)

			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(fmt.Sprintf(`{"data":{"id":"%s","type":"payin","attributes":{"resource_id":"%s"}}}`, requestID, resourceID)))
		})
		server := httptest.NewServer(mux)
		defer server.Close()

		client := quikk2.New(quikk2.Config{Endpoint: server.URL})
		// create instance of quikk service
		service := quikk.NewQuikkApi(&client, shortcode, repository)
		paymentID := ulid.Make().String()
		err := service.B2C(t.Context(), paymentID, testPayment)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		// check request is saved
		request, err := repository.FindOne(t.Context(), requests.OptionsFindRequest{ExternalID: &requestID})
		// expect no error
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		// assert request values
		assert.Equal(t, paymentID, request.PaymentID)
		// status should be succeeded for successful requests
		assert.Equal(t, requests.StatusSucceeded, request.Status)

	})
}

func TestQuikkApi_B2B(t *testing.T) {
	testPayment := mpesa.PaymentRequest{
		IdempotencyID:         ulid.Make().String(),
		ClientTransactionID:   ulid.Make().String(),
		Amount:                "105",
		ExternalAccountNumber: "254712345678",
		ExternalAccountType:   mpesa.AccountTypePaybill,
		Beneficiary:           "100200",
		Description:           "test payment",
	}

	repository := postgres.NewRequestRepository(inf.Storage.PG)

	t.Run("test b2b to paybill", func(t *testing.T) {
		requestID := ulid.Make().String()
		resourceID := ulid.Make().String()

		// create a mock test server
		mux := http.NewServeMux()
		mux.HandleFunc(quikk2.EndpointTransfer, func(w http.ResponseWriter, r *http.Request) {
			// parse request body
			var req quikk2.RequestDefault[quikk2.RequestTransfer]
			if err := jsoniter.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Errorf("expected nil error, got %v", err)
			}

			assert.Equal(t, "transfer", req.Data.Type)
			// TODO: write better test case for checking amount conversions from string to float64
			//assert.EqualValues(t, testPayment.Amount, fmt.Sprintf("%f", req.Data.Attributes.Amount))
			assert.Equal(t, testPayment.ExternalAccountNumber, req.Data.Attributes.RecipientNo)
			assert.Equal(t, "short_code", req.Data.Attributes.RecipientType)
			assert.Equal(t, "paybill", req.Data.Attributes.RecipientCategory)
			assert.Equal(t, shortcode.ShortCode, req.Data.Attributes.ShortCode)

			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(fmt.Sprintf(`{"data":{"id":"%s","type":"payin","attributes":{"resource_id":"%s"}}}`, requestID, resourceID)))
		})
		server := httptest.NewServer(mux)
		defer server.Close()

		client := quikk2.New(quikk2.Config{Endpoint: server.URL})
		// create instance of quikk service
		service := quikk.NewQuikkApi(&client, shortcode, repository)
		paymentID := ulid.Make().String()
		err := service.B2B(t.Context(), paymentID, testPayment)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		// check request is saved
		request, err := repository.FindOne(t.Context(), requests.OptionsFindRequest{ExternalID: &requestID})
		// expect no error
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		// assert request values
		assert.Equal(t, paymentID, request.PaymentID)
		// status should be succeeded for successful requests
		assert.Equal(t, requests.StatusSucceeded, request.Status)

	})

	t.Run("test b2b to till", func(t *testing.T) {
		requestID := ulid.Make().String()
		resourceID := ulid.Make().String()

		// update test payment to use till account type
		testPayment := testPayment
		testPayment.ExternalAccountType = mpesa.AccountTypeTill

		// create a mock test server
		mux := http.NewServeMux()
		mux.HandleFunc(quikk2.EndpointTransfer, func(w http.ResponseWriter, r *http.Request) {
			// parse request body
			var req quikk2.RequestDefault[quikk2.RequestTransfer]
			if err := jsoniter.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Errorf("expected nil error, got %v", err)
			}

			assert.Equal(t, "transfer", req.Data.Type)
			// TODO: write better test case for checking amount conversions from string to float64
			//assert.EqualValues(t, testPayment.Amount, fmt.Sprintf("%f", req.Data.Attributes.Amount))
			assert.Equal(t, testPayment.ExternalAccountNumber, req.Data.Attributes.RecipientNo)
			assert.Equal(t, "short_code", req.Data.Attributes.RecipientType)
			assert.Equal(t, "till", req.Data.Attributes.RecipientCategory)
			assert.Equal(t, shortcode.ShortCode, req.Data.Attributes.ShortCode)

			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(fmt.Sprintf(`{"data":{"id":"%s","type":"payin","attributes":{"resource_id":"%s"}}}`, requestID, resourceID)))
		})
		server := httptest.NewServer(mux)
		defer server.Close()

		client := quikk2.New(quikk2.Config{Endpoint: server.URL})
		// create instance of quikk service
		service := quikk.NewQuikkApi(&client, shortcode, repository)
		paymentID := ulid.Make().String()
		err := service.B2B(t.Context(), paymentID, testPayment)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		// check request is saved
		request, err := repository.FindOne(t.Context(), requests.OptionsFindRequest{ExternalID: &requestID})
		// expect no error
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		// assert request values
		assert.Equal(t, paymentID, request.PaymentID)
		// status should be succeeded for successful requests
		assert.Equal(t, requests.StatusSucceeded, request.Status)

	})
}

func TestQuikkApi_Status(t *testing.T) {
	defer testdata.ResetTables(inf)

	responseID := ulid.Make().String()
	resourceID := ulid.Make().String()
	paymentID := ulid.Make().String()

	// save a fake record
	repository := postgres.NewRequestRepository(inf.Storage.PG)
	record := requests.Request{
		RequestID: resourceID,
		PaymentID: paymentID,
		Status:    "success",
		Partner:   "test",
		Latency:   100 * time.Millisecond,
		CreatedAt: time.Now(),
	}
	err := repository.Add(t.Context(), record)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}

	t.Run("test that if payment has reference txn_id type is used", func(t *testing.T) {
		payment := mpesa.Payment{PaymentReference: ulid.Make().String()}

		// create a mock test server
		mux := http.NewServeMux()
		mux.HandleFunc(quikk2.EndpointTransactionSearch, func(w http.ResponseWriter, r *http.Request) {
			// parse request body
			var req quikk2.RequestDefault[quikk2.RequestTransactionStatus]
			if err := jsoniter.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Errorf("expected nil error, got %v", err)
			}

			assert.Equal(t, "search", req.Data.Type)
			assert.Equal(t, payment.PaymentReference, req.Data.Attributes.Reference)
			assert.Equal(t, "txn_id", req.Data.Attributes.ReferenceType)

			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(fmt.Sprintf(`{"data":{"id":"%s","type":"payout","attributes":{"resource_id":"%s"}}}`, responseID, resourceID)))
		})
		server := httptest.NewServer(mux)
		defer server.Close()

		client := quikk2.New(quikk2.Config{Endpoint: server.URL})
		// create instance of quikk service
		service := quikk.NewQuikkApi(&client, shortcode, repository)
		err = service.Status(t.Context(), payment)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

	})

	t.Run("test that if payment has no reference response_id is used", func(t *testing.T) {
		// make sure PaymentReference is empty
		payment := mpesa.Payment{PaymentID: paymentID}

		// create a mock test server
		mux := http.NewServeMux()
		mux.HandleFunc(quikk2.EndpointTransactionSearch, func(w http.ResponseWriter, r *http.Request) {
			// parse request body
			var req quikk2.RequestDefault[quikk2.RequestTransactionStatus]
			if err := jsoniter.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Errorf("expected nil error, got %v", err)
			}

			assert.Equal(t, "search", req.Data.Type)
			assert.Equal(t, record.ExternalID, req.Data.Attributes.Reference)
			assert.Equal(t, "response_id", req.Data.Attributes.ReferenceType)

			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			// it doesn't matter what we return here
			_, _ = w.Write([]byte(`{"data":{"id":"0000","type":"payout","attributes":{"resource_id":"0000"}}}`))
		})
		server := httptest.NewServer(mux)
		defer server.Close()

		client := quikk2.New(quikk2.Config{Endpoint: server.URL})
		// create instance of quikk service
		service := quikk.NewQuikkApi(&client, shortcode, repository)
		err = service.Status(t.Context(), payment)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

	})
}
