package services

import (
	"github.com/rs/zerolog"

	clients_daraja "github.com/SirWaithaka/payments-api/clients/daraja"
	clients_quikk "github.com/SirWaithaka/payments-api/clients/quikk"
	"github.com/SirWaithaka/payments-api/corehooks"
	"github.com/SirWaithaka/payments-api/internal/config"
	"github.com/SirWaithaka/payments-api/internal/domains/mpesa"
	"github.com/SirWaithaka/payments-api/internal/domains/requests"
	"github.com/SirWaithaka/payments-api/internal/domains/webhooks"
	"github.com/SirWaithaka/payments-api/internal/services/daraja"
	"github.com/SirWaithaka/payments-api/internal/services/quikk"
	"github.com/SirWaithaka/payments-api/request"
)

// WithLogger fetches the zerolog logger instance from the request context
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

func NewProvider(cfg config.Config, requestsRepo requests.Repository, webhooksRepo webhooks.Repository) *Provider {
	return &Provider{config: cfg, requestsRepo: requestsRepo, webhooksRepo: webhooksRepo}
}

type Provider struct {
	config       config.Config
	requestsRepo requests.Repository
	webhooksRepo webhooks.Repository
}

func (provider Provider) GetMpesaApi(shortcode mpesa.ShortCode) mpesa.API {
	// build client depending on service
	if shortcode.Service == requests.PartnerDaraja {
		certificate := clients_daraja.SandboxCertificate
		if shortcode.Environment == "production" {
			certificate = clients_daraja.ProductionCertificate
		}

		// build the daraja client
		client := provider.GetDarajaClient(shortcode)
		return daraja.NewDarajaApi(client, certificate, shortcode, provider.requestsRepo)
	}
	if shortcode.Service == requests.PartnerQuikk {
		// build the quikk client
		client := provider.GetQuikkClient(shortcode)
		return quikk.NewQuikkApi(client, shortcode, provider.requestsRepo)
	}

	return nil
}

func (provider Provider) GetWebhookProcessor(service requests.Partner) requests.WebhookProcessor {
	switch service {
	case requests.PartnerDaraja:
		return daraja.NewWebhookProcessor()
	case requests.PartnerQuikk:
		return quikk.NewWebhookProcessor()
	default:
		return nil
	}
}

func (provider Provider) GetDarajaClient(shortcode mpesa.ShortCode) *clients_daraja.Client {
	endpoint := clients_daraja.SandboxUrl
	// check the environment the shortcode is configured for
	if shortcode.Environment == "production" {
		endpoint = clients_daraja.ProductionUrl
	}

	// use environment endpoint if set
	if provider.config.Daraja.Endpoint != "" {
		endpoint = provider.config.Daraja.Endpoint
	}

	client := clients_daraja.New(clients_daraja.Config{Endpoint: endpoint, LogLevel: request.LogError})
	client.Hooks.Build.PushFront(WithLogger())
	client.Hooks.Build.PushBackHook(clients_daraja.Authenticate(client.AuthenticationRequest(shortcode.Key, shortcode.Secret)))
	client.Hooks.Send.PushFrontHook(corehooks.LogHTTPRequest)

	return &client

}

func (provider Provider) GetQuikkClient(shortcode mpesa.ShortCode) *clients_quikk.Client {
	endpoint := clients_quikk.SandboxUrl
	// check the environment the shortcode is configured for
	if shortcode.Environment == "production" {
		endpoint = clients_quikk.ProductionUrl
	}

	// use environment endpoint if set
	if provider.config.Quikk.Endpoint != "" {
		endpoint = provider.config.Quikk.Endpoint
	}

	client := clients_quikk.New(clients_quikk.Config{Endpoint: endpoint, LogLevel: request.LogError})
	client.Hooks.Build.PushFront(WithLogger())
	client.Hooks.Build.PushBackHook(clients_quikk.Sign(shortcode.Key, shortcode.Secret))
	client.Hooks.Send.PushFrontHook(corehooks.LogHTTPRequest)

	return &client
}
