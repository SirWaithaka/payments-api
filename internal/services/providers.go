package services

import (
	"github.com/rs/zerolog"

	clients_daraja "github.com/SirWaithaka/payments-api/clients/daraja"
	clients_quikk "github.com/SirWaithaka/payments-api/clients/quikk"
	"github.com/SirWaithaka/payments-api/corehooks"
	"github.com/SirWaithaka/payments-api/internal/domains/mpesa"
	"github.com/SirWaithaka/payments-api/internal/domains/requests"
	"github.com/SirWaithaka/payments-api/internal/domains/webhooks"
	"github.com/SirWaithaka/payments-api/internal/services/daraja"
	"github.com/SirWaithaka/payments-api/internal/services/quikk"
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
		client := provider.GetDarajaClient(clients_daraja.SandboxUrl, shortcode)
		return daraja.NewDarajaApi(client, shortcode, provider.requestsRepo)
	}
	if shortcode.Service == requests.PartnerQuikk {
		// build the quikk client
		client := provider.GetQuikkClient(clients_quikk.SandboxUrl, shortcode)
		return quikk.NewQuikkApi(client, shortcode, provider.requestsRepo)
	}

	return nil
}

func (provider Provider) GetWebhookClient(service requests.Partner) requests.WebhookProcessor {
	switch service {
	case requests.PartnerDaraja:
		return daraja.NewWebhookProcessor()
	case requests.PartnerQuikk:
		return quikk.NewWebhookProcessor()
	default:
		return nil
	}
}

func (provider Provider) GetDarajaClient(endpoint string, cfg mpesa.ShortCode) *clients_daraja.Client {

	client := clients_daraja.New(clients_daraja.Config{Endpoint: endpoint, LogLevel: request.LogError})
	client.Hooks.Build.PushFront(WithLogger())
	client.Hooks.Build.PushBackHook(clients_daraja.Authenticate(client.AuthenticationRequest(cfg.Key, cfg.Secret)))
	client.Hooks.Send.PushFrontHook(corehooks.LogHTTPRequest)

	return &client

}

func (provider Provider) GetQuikkClient(endpoint string, shortcode mpesa.ShortCode) *clients_quikk.Client {
	client := clients_quikk.New(clients_quikk.Config{Endpoint: endpoint, LogLevel: request.LogError})
	client.Hooks.Build.PushFront(WithLogger())
	client.Hooks.Build.PushBackHook(clients_quikk.Sign(shortcode.Key, shortcode.Secret))
	client.Hooks.Build.PushFront(request.WithRequestHeader("accept", "application/vnd.api+json"))
	client.Hooks.Build.PushFront(request.WithRequestHeader("content-type", "application/vnd.api+json"))
	client.Hooks.Send.PushFrontHook(corehooks.LogHTTPRequest)

	return &client
}
