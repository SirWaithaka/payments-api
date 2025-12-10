package mpesa

import (
	"context"
	"errors"

	"github.com/SirWaithaka/payments-api/src/domains/requests"
)

func NewServiceShortCode(repository ShortCodeRepository) ServiceShortCode {
	return ServiceShortCode{repository: repository}
}

type ServiceShortCode struct {
	repository ShortCodeRepository
}

func (service ServiceShortCode) Add(ctx context.Context, shortcode ShortCode) error {
	if !shortcode.Type.Valid() {
		return errors.New("unknown type on shortcode")
	}

	if shortcode.Type == PaymentTypeCharge && shortcode.Passphrase == "" {
		return errors.New("missing passphrase for charge shortcode")
	}

	// default service to daraja
	if shortcode.Service == requests.PartnerUnknown {
		shortcode.Service = requests.PartnerDaraja
	}

	// default new shortcodes to priority 10
	if shortcode.Priority == 0 {
		shortcode.Priority = 10
	}

	return service.repository.Add(ctx, shortcode)
}
