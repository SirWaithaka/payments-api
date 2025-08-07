package postgres_test

import (
	"errors"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/oklog/ulid/v2"

	"github.com/SirWaithaka/payments-api/internal/domains/mpesa"
	"github.com/SirWaithaka/payments-api/internal/repositories/postgres"
	"github.com/SirWaithaka/payments-api/internal/testdata"
)

func TestShortCodeRepository_Add(t *testing.T) {
	repo := postgres.NewShortCodeRepository(inf.Storage.PG)

	t.Run("test that it saves record", func(t *testing.T) {
		defer testdata.ResetTables(inf)

		shortcode := mpesa.ShortCodeConfig{
			ShortCodeID:       ulid.Make().String(),
			ShortCode:         "000000",
			Service:           "daraja",
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
		record := postgres.ShortcodeSchema{}
		result := inf.Storage.PG.First(&record)
		if err = result.Error; err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		assert.Equal(t, shortcode, record.ToEntity())
	})

	t.Run("test that empty string values are not saved", func(t *testing.T) {
		defer testdata.ResetTables(inf)

		testcases := []struct {
			name  string
			input mpesa.ShortCodeConfig
		}{
			{
				name: "test check constraint on shortcode",
				input: mpesa.ShortCodeConfig{
					ShortCodeID:       ulid.Make().String(),
					ShortCode:         "",
					Service:           "daraja",
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
				input: mpesa.ShortCodeConfig{
					ShortCodeID:       ulid.Make().String(),
					ShortCode:         "000000",
					Service:           "",
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
				input: mpesa.ShortCodeConfig{
					ShortCodeID:       ulid.Make().String(),
					ShortCode:         "000000",
					Service:           "daraja",
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
				input: mpesa.ShortCodeConfig{
					ShortCodeID:       ulid.Make().String(),
					ShortCode:         "000000",
					Service:           "daraja",
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
}

func TestShortCodeRepository_Find(t *testing.T) {
	repo := postgres.NewShortCodeRepository(inf.Storage.PG)

	t.Run("test that it finds records", func(t *testing.T) {
		defer testdata.ResetTables(inf)

		shortcode := mpesa.ShortCodeConfig{
			ShortCodeID:       ulid.Make().String(),
			ShortCode:         "000000",
			Service:           "daraja",
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
		result, err := repo.Find(t.Context(), shortcode.ShortCodeID)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		assert.Equal(t, shortcode, result)
	})

}
