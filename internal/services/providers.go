package services

import (
	"github.com/rs/zerolog"

	"github.com/SirWaithaka/payments-api/clients/daraja"
	"github.com/SirWaithaka/payments-api/corehooks"
	"github.com/SirWaithaka/payments-api/internal/domains/mpesa"
	"github.com/SirWaithaka/payments-api/internal/domains/payments"
	"github.com/SirWaithaka/payments-api/internal/domains/requests"
	"github.com/SirWaithaka/payments-api/internal/domains/webhooks"
	"github.com/SirWaithaka/payments-api/request"
)

// WithLogger fetches the zerolog logger instance from request context
// and passes it to the request config
func WithLogger() request.Option {
	return func(r *request.Request) {
		l := zerolog.Ctx(r.Context()).With().CallerWithSkipFrameCount(3).Logger()
		lg := NewLogger(&l, r.Config.LogLevel)
		r.Config.Logger = lg
	}
}

type ShortCodeConfig struct {
	ShortCode         string // business pay bill or buy goods account
	InitiatorName     string // daraja api initiator name
	InitiatorPassword string // daraja api initiator password
	Passphrase        string // (optional) passphrase for c2b transfers
	ConsumerKey       string // daraja app consumer key
	ConsumerSecret    string // daraja app consumer secret
	CallbackURL       string // callback url for shortcode async responses
}

func NewProvider(requestsRepo requests.Repository, webhooksRepo webhooks.Repository) *Provider {
	return &Provider{requestsRepo: requestsRepo, webhooksRepo: webhooksRepo}
}

type Provider struct {
	requestsRepo requests.Repository
	webhooksRepo webhooks.Repository
}

//func (provider Provider) GetWalletApi(bankCode string, reqType payments.RequestType) payments.WalletApi {
//
//	if bankCode == payments.BankMpesa {
//		shortcodeCfg, _ := provider.GetShortCodeConfig(reqType)
//		// build the daraja client
//		client := provider.GetDarajaClient(daraja.SandboxUrl, shortcodeCfg)
//		return NewDarajaApi(client, shortcodeCfg, provider.requestsRepo)
//	}
//
//	return nil
//}

func (provider Provider) GetMpesaApi(shortcode mpesa.ShortCode) mpesa.API {
	// build client depending on service
	if shortcode.Service == requests.PartnerDaraja {
		// build the daraja client
		client := provider.GetDarajaClient(daraja.SandboxUrl, shortcode)
		return NewDarajaApi(client, shortcode, provider.requestsRepo)
	}

	return nil
}

func (provider Provider) GetWebhookClient(service requests.Partner) requests.WebhookProcessor {
	switch service {
	case requests.PartnerDaraja:
		return NewWebhookProcessor()
	default:
		return nil
	}
}

func (provider Provider) GetDarajaClient(endpoint string, cfg mpesa.ShortCode) *daraja.Client {

	client := daraja.New(daraja.Config{Endpoint: endpoint, LogLevel: request.LogError})
	client.Hooks.Build.PushFront(WithLogger())
	client.Hooks.Build.PushBackHook(daraja.Authenticate(client.AuthenticationRequest(cfg.Key, cfg.Secret)))
	client.Hooks.Send.PushFrontHook(corehooks.LogHTTPRequest)

	return &client

}

func (provider Provider) GetShortCodeConfig(name payments.RequestType) (ShortCodeConfig, error) {
	switch name {
	case payments.RequestTypeWalletCharge:
		return ShortCodeConfig{
			ShortCode:         "174379",
			InitiatorName:     "testapi",
			InitiatorPassword: "Safaricom123!!",
			Passphrase:        "bfb279f9aa9bdbcf158e97dd71a467cd2e0c893059b10f78e6b72ada1ed2c919",
			ConsumerKey:       "7nRVPmgCrfIEseRTTmkLDDqAYAKhhS9KWx0AfYLGj9NVE2C2",
			ConsumerSecret:    "Cyq7VrtT1vzQAmPQV1zlrC9MZ2n6py6qqaLYzNgFAx6uDG8sTKYSVoCh8sdplZF7",
			CallbackURL:       "https://webhook.sirwaithaka.space/webhooks/daraja",
		}, nil

	case payments.RequestTypeWalletPayout:
		return ShortCodeConfig{
			ShortCode:         "600991",
			InitiatorName:     "testapi",
			InitiatorPassword: "Safaricom123!!",
			ConsumerKey:       "7nRVPmgCrfIEseRTTmkLDDqAYAKhhS9KWx0AfYLGj9NVE2C2",
			ConsumerSecret:    "Cyq7VrtT1vzQAmPQV1zlrC9MZ2n6py6qqaLYzNgFAx6uDG8sTKYSVoCh8sdplZF7",
			CallbackURL:       "https://webhook.sirwaithaka.space/webhooks/daraja",
		}, nil
	case payments.RequestTypeWalletTransfer:
		return ShortCodeConfig{
			ShortCode:         "600979",
			InitiatorName:     "testapi",
			InitiatorPassword: "Safaricom123!!",
			ConsumerKey:       "7nRVPmgCrfIEseRTTmkLDDqAYAKhhS9KWx0AfYLGj9NVE2C2",
			ConsumerSecret:    "Cyq7VrtT1vzQAmPQV1zlrC9MZ2n6py6qqaLYzNgFAx6uDG8sTKYSVoCh8sdplZF7",
			CallbackURL:       "https://webhook.sirwaithaka.space/webhooks/daraja",
		}, nil
	case payments.RequestTypePaymentStatus:
		return ShortCodeConfig{
			ShortCode:         "000000",
			InitiatorName:     "testapi",
			InitiatorPassword: "Safaricom123!!",
			ConsumerKey:       "7nRVPmgCrfIEseRTTmkLDDqAYAKhhS9KWx0AfYLGj9NVE2C2",
			ConsumerSecret:    "Cyq7VrtT1vzQAmPQV1zlrC9MZ2n6py6qqaLYzNgFAx6uDG8sTKYSVoCh8sdplZF7",
			CallbackURL:       "https://webhook.sirwaithaka.space/webhooks/daraja",
		}, nil
	default:
		return ShortCodeConfig{}, nil
	}

}
