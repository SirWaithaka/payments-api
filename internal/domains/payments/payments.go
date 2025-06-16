package payments

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"

	"github.com/SirWaithaka/payments-api/clients/daraja"
	"github.com/SirWaithaka/payments-api/internal/pkg/logger"
)

const (
	darajaEndpoint = "http://localhost:9002/daraja"
)

type Payment struct {
	// unique payment reference
	Reference string
	// source account
	AccountNumber string
	// destination account
	ExternalAccountNumber string
	// (optional)
	BeneficiaryAccountNumber string
	// amount for payment
	Amount string
	// short description for payment
	Description string
	// external payment reference
	ExternalID string
	// external idempotent unique payment reference
	ExternalUID string
}

func NewService() Service {
	return Service{provider: Provider{}}
}

type Service struct {
	provider Provider
}

func (service Service) Transact(ctx context.Context, payment Payment) error {
	l := zerolog.Ctx(ctx)

	shortcode, _ := service.provider.GetShortCodeConfig("C2B")
	client := service.provider.GetDarajaClient(darajaEndpoint, shortcode)

	credential, err := daraja.OpenSSLEncrypt(shortcode.InitiatorPassword, daraja.SandboxCertificate)
	if err != nil {
		l.Error().Err(err).Msg("error encrypting")
		return err
	}

	res, err := client.B2B(ctx, daraja.RequestB2B{
		Initiator:              shortcode.InitiatorName,
		SecurityCredential:     credential,
		CommandID:              daraja.CommandBusinessPayBill,
		SenderIdentifierType:   daraja.IdentifierOrgShortCode,
		RecieverIdentifierType: daraja.IdentifierOrgShortCode,
		Amount:                 payment.Amount,
		PartyA:                 shortcode.ShortCode,
		PartyB:                 payment.ExternalAccountNumber,
		AccountReference:       payment.BeneficiaryAccountNumber,
		Remarks:                fmt.Sprintf("B2B REF %s ID %s", payment.Reference, payment.ExternalID),
		QueueTimeOutURL:        shortcode.CallbackURL,
		ResultURL:              shortcode.CallbackURL,
	})
	if err != nil {
		l.Error().Err(err).Msg("client error")
		return err
	}

	l.Info().Any(logger.LData, res).Msg("response")

	return nil
}
