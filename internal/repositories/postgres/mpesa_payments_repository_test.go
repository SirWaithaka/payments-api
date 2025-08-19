package postgres_test

import (
	"testing"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"

	"github.com/SirWaithaka/payments-api/internal/domains/mpesa"
	"github.com/SirWaithaka/payments-api/internal/domains/requests"
	"github.com/SirWaithaka/payments-api/internal/pkg/types"
	"github.com/SirWaithaka/payments-api/internal/repositories/postgres"
	"github.com/SirWaithaka/payments-api/internal/testdata"
)

func TestMpesaPaymentRepository_Add(t *testing.T) {
	repo := postgres.NewMpesaPaymentsRepository(inf.Storage.PG)

	t.Run("test that it saves record", func(t *testing.T) {
		defer testdata.ResetTables(inf)

		payment := mpesa.Payment{
			PaymentID:                ulid.Make().String(),
			Type:                     mpesa.PaymentTypeTransfer,
			Status:                   requests.StatusSent,
			ClientTransactionID:      ulid.Make().String(),
			IdempotencyID:            ulid.Make().String(),
			PaymentReference:         ulid.Make().String(),
			Amount:                   "10",
			SourceAccountNumber:      "fake_account",
			DestinationAccountNumber: "fake_account",
			Beneficiary:              "fake_beneficiary",
			Description:              "fake_description",
			ShortCodeID:              ulid.Make().String(),
		}

		err := repo.Add(t.Context(), payment)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		// fetch record and check values
		record := postgres.MpesaPaymentSchema{}
		result := inf.Storage.PG.WithContext(t.Context()).First(&record)
		if err = result.Error; err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		assert.Equal(t, payment, record.ToEntity())
	})

}

func TestMpesaPaymentsRepository_FindOne(t *testing.T) {
	repo := postgres.NewMpesaPaymentsRepository(inf.Storage.PG)

	t.Run("test that it finds records", func(t *testing.T) {
		defer testdata.ResetTables(inf)

		record := mpesa.Payment{
			PaymentID:           ulid.Make().String(),
			ClientTransactionID: ulid.Make().String(),
			IdempotencyID:       ulid.Make().String(),
			PaymentReference:    ulid.Make().String(),
			Status:              requests.StatusSent,
			Description:         "fake_description",
		}
		if err := repo.Add(t.Context(), record); err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		testcases := []struct {
			name  string
			input mpesa.OptionsFindPayment
		}{
			{
				name:  "test that it finds record by payment id",
				input: mpesa.OptionsFindPayment{PaymentID: &record.PaymentID},
			},
			{
				name:  "test that it finds record by client transaction id",
				input: mpesa.OptionsFindPayment{ClientTransactionID: &record.ClientTransactionID},
			},
			{
				name:  "test that it finds record by idempotency id",
				input: mpesa.OptionsFindPayment{IdempotencyID: &record.IdempotencyID},
			},
			{
				name:  "test that it finds record by payment reference",
				input: mpesa.OptionsFindPayment{PaymentReference: &record.PaymentReference},
			},
		}

		for _, tc := range testcases {
			t.Run(tc.name, func(t *testing.T) {
				payment, err := repo.FindOne(t.Context(), tc.input)
				if err != nil {
					t.Errorf("expected nil error, got %v", err)
				}

				assert.Equal(t, record.PaymentID, payment.PaymentID)
				assert.Equal(t, record.ClientTransactionID, payment.ClientTransactionID)
				assert.Equal(t, record.IdempotencyID, payment.IdempotencyID)
				assert.Equal(t, record.PaymentReference, payment.PaymentReference)
			})
		}
	})
}

func TestMpesaPaymentsRepository_Update(t *testing.T) {
	repo := postgres.NewMpesaPaymentsRepository(inf.Storage.PG)

	testcases := []struct {
		name  string
		input mpesa.OptionsUpdatePayment
	}{
		{
			name: "test all values provided",
			input: mpesa.OptionsUpdatePayment{
				Status:           types.Pointer(requests.StatusSucceeded),
				PaymentReference: types.Pointer(ulid.Make().String()),
			},
		},
	}

	for _, tc := range testcases {

		t.Run(tc.name, func(t *testing.T) {
			defer testdata.ResetTables(inf)

			// create and save payment record

			payment := mpesa.Payment{
				PaymentID:           ulid.Make().String(),
				ClientTransactionID: ulid.Make().String(),
				IdempotencyID:       ulid.Make().String(),
				PaymentReference:    ulid.Make().String(),
				Status:              "status",
				Description:         "fake_description",
			}

			err := repo.Add(t.Context(), payment)
			if err != nil {
				t.Errorf("expected nil error, got %v", err)
			}

			err = repo.Update(t.Context(), payment.PaymentID, tc.input)
			if err != nil {
				t.Errorf("expected nil error, got %v", err)
			}

			// pull the record and check values
			record := postgres.MpesaPaymentSchema{}
			r := inf.Storage.PG.Where(postgres.MpesaPaymentSchema{PaymentID: payment.PaymentID}).First(&record)
			if r.Error != nil {
				t.Errorf("expected nil error, got %v", r.Error)
			}

			assert.Equal(t, string(*tc.input.Status), record.Status)
			assert.Equal(t, tc.input.PaymentReference, record.PaymentReference)
		})
	}
}
