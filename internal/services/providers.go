package services

import (
	"github.com/rs/zerolog"

	"github.com/SirWaithaka/payments-api/clients/daraja"
	"github.com/SirWaithaka/payments-api/corehooks"
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
		lg := NewLogger(&l)
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

func (provider Provider) GetWalletApi(bankCode string, reqType payments.RequestType) payments.WalletApi {

	if bankCode == payments.BankMpesa {
		shortcodeCfg, _ := provider.GetShortCodeConfig(reqType)
		// build the daraja client
		client := provider.GetDarajaClient(daraja.SandboxUrl, shortcodeCfg)
		return NewDarajaApi(client, shortcodeCfg, provider.requestsRepo)
	}

	return nil
}

func (provider Provider) GetWebhookClient(service string) requests.WebhookProcessor {
	switch service {
	case "daraja":
		return NewWebhookProcessor()
	default:
		return nil
	}
}

func (provider Provider) GetDarajaClient(endpoint string, cfg ShortCodeConfig) *daraja.Client {

	client := daraja.New(daraja.Config{Endpoint: endpoint, LogLevel: request.LogError})
	client.Hooks.Build.PushBackHook(daraja.Authenticate(client.AuthenticationRequest(cfg.ConsumerKey, cfg.ConsumerSecret)))
	client.Hooks.Build.PushBack(WithLogger())
	client.Hooks.Send.PushFrontHook(corehooks.LogHTTPRequest)

	return &client

}

func (provider Provider) GetShortCodeConfig(name payments.RequestType) (ShortCodeConfig, error) {
	switch name {
	case payments.RequestTypeWalletCharge:
		return ShortCodeConfig{
			ShortCode:         "174379",
			InitiatorName:     "testapi",
			InitiatorPassword: "Safaricom999!*!%",
			Passphrase:        "bfb279f9aa9bdbcf158e97dd71a467cd2e0c893059b10f78e6b72ada1ed2c919",
			ConsumerKey:       "GW0TvN2gUTakps3b1AbAw48no1Yogu92oXI0N55fmlEVK40p",
			ConsumerSecret:    "CtpMOjvk47jm6A5hmCzaQjQTBOWAwK1LM95awGNLSTGawbGNPsy9f8Eabsr1Lg7Q",
			CallbackURL:       "https://webhook.sirwaithaka.space/webhooks/daraja",
		}, nil

	case payments.RequestTypeWalletPayout:
		return ShortCodeConfig{
			ShortCode:         "000000",
			InitiatorName:     "testapi",
			InitiatorPassword: "Safaricom999!*!%",
			ConsumerKey:       "GW0TvN2gUTakps3b1AbAw48no1Yogu92oXI0N55fmlEVK40p",
			ConsumerSecret:    "CtpMOjvk47jm6A5hmCzaQjQTBOWAwK1LM95awGNLSTGawbGNPsy9f8Eabsr1Lg7Q",
			CallbackURL:       "https://webhook.sirwaithaka.space/webhooks/daraja",
		}, nil
	case payments.RequestTypeWalletTransfer:
		return ShortCodeConfig{
			ShortCode:         "000000",
			InitiatorName:     "testapi",
			InitiatorPassword: "Safaricom999!*!",
			ConsumerKey:       "GW0TvN2gUTakps3b1AbAw48no1Yogu92oXI0N55fmlEVK40p",
			ConsumerSecret:    "CtpMOjvk47jm6A5hmCzaQjQTBOWAwK1LM95awGNLSTGawbGNPsy9f8Eabsr1Lg7Q",
			CallbackURL:       "https://webhook.sirwaithaka.space/webhooks/daraja",
		}, nil
	case payments.RequestTypePaymentStatus:
		return ShortCodeConfig{
			ShortCode:         "000000",
			InitiatorName:     "testapi",
			InitiatorPassword: "Safaricom999!*!",
			ConsumerKey:       "GW0TvN2gUTakps3b1AbAw48no1Yogu92oXI0N55fmlEVK40p",
			ConsumerSecret:    "CtpMOjvk47jm6A5hmCzaQjQTBOWAwK1LM95awGNLSTGawbGNPsy9f8Eabsr1Lg7Q",
			CallbackURL:       "https://webhook.sirwaithaka.space/webhooks/daraja",
		}, nil
	default:
		return ShortCodeConfig{}, nil
	}

}
