package services

import (
	"github.com/rs/zerolog"

	"github.com/SirWaithaka/payments-api/clients/daraja"
	"github.com/SirWaithaka/payments-api/corehooks"
	"github.com/SirWaithaka/payments-api/internal/domains/payments"
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

func NewProvider() *Provider {
	return &Provider{}
}

type Provider struct {
	webhooksRepo webhooks.WebhookRepository
}

func (provider Provider) GetWalletApi(request payments.WalletPayment) payments.WalletApi {

	if request.BankCode == payments.BankMpesa {
		shortcodeCfg, _ := provider.GetShortCodeConfig(request.Type)
		// build the daraja client
		client := provider.GetDarajaClient(daraja.SandboxUrl, shortcodeCfg)
		return NewDarajaApi(client, shortcodeCfg)
	}

	return nil
}

func (provider Provider) GetWebhookClient(service string) webhooks.WebhookProcessor {
	switch service {
	case "daraja":
		return NewWebhookProcessor(provider.webhooksRepo)
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

func (provider Provider) GetShortCodeConfig(name string) (ShortCodeConfig, error) {
	switch name {
	case "CHARGE":
		return ShortCodeConfig{
			ShortCode:         "174379",
			InitiatorName:     "testapi",
			InitiatorPassword: "Safaricom999!*!%",
			Passphrase:        "bfb279f9aa9bdbcf158e97dd71a467cd2e0c893059b10f78e6b72ada1ed2c919",
			ConsumerKey:       "GW0TvN2gUTakps3b1AbAw48no1Yogu92oXI0N55fmlEVK40p",
			ConsumerSecret:    "CtpMOjvk47jm6A5hmCzaQjQTBOWAwK1LM95awGNLSTGawbGNPsy9f8Eabsr1Lg7Q",
			CallbackURL:       "https://webhook.sirwaithaka.space/webhooks/daraja",
		}, nil

	case "PAYOUT":
		return ShortCodeConfig{
			ShortCode:         "000000",
			InitiatorName:     "testapi",
			InitiatorPassword: "Safaricom999!*!%",
			ConsumerKey:       "GW0TvN2gUTakps3b1AbAw48no1Yogu92oXI0N55fmlEVK40p",
			ConsumerSecret:    "CtpMOjvk47jm6A5hmCzaQjQTBOWAwK1LM95awGNLSTGawbGNPsy9f8Eabsr1Lg7Q",
			CallbackURL:       "https://webhook.sirwaithaka.space/webhooks/daraja",
		}, nil
	case "TRANSFER":
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
