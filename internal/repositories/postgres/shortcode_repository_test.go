package postgres_test

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"

	"github.com/SirWaithaka/payments-api/internal/domains/mpesa"
	"github.com/SirWaithaka/payments-api/internal/domains/requests"
	pkgerrors "github.com/SirWaithaka/payments-api/internal/pkg/errors"
	"github.com/SirWaithaka/payments-api/internal/repositories/postgres"
	"github.com/SirWaithaka/payments-api/internal/testdata"
)

func TestShortCodeRepository_Add(t *testing.T) {
	repo := postgres.NewShortCodeRepository(inf.Storage.PG)

	t.Run("test that it saves record", func(t *testing.T) {
		defer testdata.ResetTables(inf)

		shortcode := mpesa.ShortCode{
			ShortCodeID:       ulid.Make().String(),
			ShortCode:         "000000",
			Priority:          1,
			Environment:       "sandbox",
			Service:           requests.PartnerDaraja,
			Type:              mpesa.PaymentTypeCharge,
			InitiatorName:     "fake name",
			InitiatorPassword: "fake_password",
			Passphrase:        "fake_passphrase",
			Key:               "fake_key",
			Secret:            "fake_secret",
			CallbackURL:       "fake_url",
		}
		err := repo.Add(t.Context(), shortcode)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		// fetch records
		record := postgres.ShortCodeSchema{}
		result := inf.Storage.PG.First(&record)
		if err = result.Error; err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		assert.Equal(t, shortcode, record.ToEntity())
	})

	t.Run("test that empty string values are not saved on select columns", func(t *testing.T) {
		defer testdata.ResetTables(inf)

		testcases := []struct {
			name  string
			input mpesa.ShortCode
		}{
			{
				name: "test check constraint on shortcode",
				input: mpesa.ShortCode{
					ShortCodeID:       ulid.Make().String(),
					ShortCode:         "",
					Environment:       "sandbox",
					Service:           requests.PartnerDaraja,
					Type:              mpesa.PaymentTypeCharge,
					InitiatorName:     "fake name",
					InitiatorPassword: "fake_password",
					Passphrase:        "fake_passphrase",
					Key:               "fake_key",
					Secret:            "fake_secret",
					CallbackURL:       "fake_url",
				},
			},
			{
				name: "test check constraint on service",
				input: mpesa.ShortCode{
					ShortCodeID: ulid.Make().String(),
					ShortCode:   "000000",
					//Service:           "",
					Environment:       "sandbox",
					Type:              mpesa.PaymentTypeCharge,
					InitiatorName:     "fake name",
					InitiatorPassword: "fake_password",
					Passphrase:        "fake_passphrase",
					Key:               "fake_key",
					Secret:            "fake_secret",
					CallbackURL:       "fake_url",
				},
			},
			{
				name: "test check constraint on key",
				input: mpesa.ShortCode{
					ShortCodeID:       ulid.Make().String(),
					ShortCode:         "000000",
					Environment:       "sandbox",
					Service:           requests.PartnerDaraja,
					Type:              mpesa.PaymentTypeCharge,
					InitiatorName:     "fake name",
					InitiatorPassword: "fake_password",
					Passphrase:        "fake_passphrase",
					Key:               "",
					Secret:            "fake_secret",
					CallbackURL:       "fake_url",
				},
			},
			{
				name: "test check constraint on secret",
				input: mpesa.ShortCode{
					ShortCodeID:       ulid.Make().String(),
					ShortCode:         "000000",
					Environment:       "sandbox",
					Service:           requests.PartnerDaraja,
					Type:              mpesa.PaymentTypeCharge,
					InitiatorName:     "fake name",
					InitiatorPassword: "fake_password",
					Passphrase:        "fake_passphrase",
					Key:               "fake_key",
					Secret:            "",
				},
			},
		}

		for _, tc := range testcases {
			t.Run(tc.name, func(t *testing.T) {
				err := repo.Add(t.Context(), tc.input)
				if err == nil {
					t.Errorf("expected non-nil error")
				}

				// confirm error is a check constraint violation
				pgErr := &pgconn.PgError{}
				if errors.As(err, &pgErr) {
					if pgErr.Code != postgres.PgCodeCheckConstraintViolation {
						t.Errorf("expected check constraint violation error, got %s", pgErr.Code)
					}
				} else {
					t.Errorf("expected pg error, got %T %v", err, err)
				}

			})
		}

	})

	t.Run("test that 2 records with identical service,shortcode,type wont be saved", func(t *testing.T) {
		defer testdata.ResetTables(inf)

		shortcode1 := mpesa.ShortCode{
			ShortCodeID: ulid.Make().String(),
			Environment: "sandbox",
			Priority:    1,
			ShortCode:   "000000",
			Service:     requests.PartnerDaraja,
			Type:        mpesa.PaymentTypeCharge,
			Key:         "fake_key",
			Secret:      "fake_secret",
		}
		shortcode2 := mpesa.ShortCode{
			ShortCodeID: ulid.Make().String(),
			Environment: "sandbox",
			Priority:    2,
			ShortCode:   "000000",                // same shortcode
			Service:     requests.PartnerDaraja,  // same service
			Type:        mpesa.PaymentTypeCharge, // same type
			Key:         "fake_key",
			Secret:      "fake_secret",
		}

		err := repo.Add(t.Context(), shortcode1)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		// adding shortcode2 should return an error
		err = repo.Add(t.Context(), shortcode2)
		if err == nil {
			t.Errorf("expected non-nil error")
		}

		// assert the error implements the duplicate interface
		if e, ok := err.(pkgerrors.Duplicate); !ok || !e.Duplicate() {
			t.Errorf("expected duplicate error, got %T %v", err, err)
		}

	})

	t.Run("test that 2 records with the same priority and type wont be saved", func(t *testing.T) {
		defer testdata.ResetTables(inf)

		shortcode1 := mpesa.ShortCode{
			ShortCodeID: ulid.Make().String(),
			ShortCode:   "000000",
			Environment: "sandbox",
			Priority:    1,
			Service:     requests.PartnerDaraja,
			Type:        mpesa.PaymentTypeCharge,
			Key:         "fake_key",
			Secret:      "fake_secret",
		}
		shortcode2 := mpesa.ShortCode{
			ShortCodeID: ulid.Make().String(),
			ShortCode:   "000001",
			Environment: "sandbox",
			Priority:    1,
			Service:     requests.PartnerDaraja,
			Type:        mpesa.PaymentTypeCharge,
			Key:         "fake_key",
			Secret:      "fake_secret",
		}

		err := repo.Add(t.Context(), shortcode1)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		err = repo.Add(t.Context(), shortcode2)
		if err == nil {
			t.Errorf("expected non-nil error")
		}

		// assert the error implements the duplicate interface
		if e, ok := err.(pkgerrors.Duplicate); !ok || !e.Duplicate() {
			t.Errorf("expected duplicate error, got %T %v", err, err)
		}

	})
}

func TestShortCodeRepository_FindOne(t *testing.T) {
	repo := postgres.NewShortCodeRepository(inf.Storage.PG)

	t.Run("test that it finds records", func(t *testing.T) {
		defer testdata.ResetTables(inf)

		shortcode := mpesa.ShortCode{
			ShortCodeID:       ulid.Make().String(),
			ShortCode:         "000000",
			Environment:       "sandbox",
			Priority:          1,
			Service:           requests.PartnerDaraja,
			Type:              mpesa.PaymentTypeCharge,
			InitiatorName:     "fake name",
			InitiatorPassword: "fake_password",
			Passphrase:        "fake_passphrase",
			Key:               "fake_key",
			Secret:            "fake_secret",
			CallbackURL:       "fake_url",
		}
		err := repo.Add(t.Context(), shortcode)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		// fetch records
		result, err := repo.FindOne(t.Context(), mpesa.OptionsFindShortCodes{ShortCodeID: &shortcode.ShortCodeID})
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		assert.Equal(t, shortcode, result)
	})

}
