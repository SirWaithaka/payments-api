package payments

import (
	"github.com/SirWaithaka/payments-api/clients/daraja"
	"github.com/SirWaithaka/payments-api/corehooks"
	"github.com/SirWaithaka/payments-api/request"
)

type ShortCodeConfig struct {
	ShortCode         string // business pay bill or buy goods account
	InitiatorName     string // daraja api initiator name
	InitiatorPassword string // daraja api initiator password
	Passphrase        string // (optional) passphrase for c2b transfers
	ConsumerKey       string // daraja app consumer key
	ConsumerSecret    string // daraja app consumer secret
	CallbackURL       string // callback url for shortcode async responses
}

type Provider struct {
}

func (provider Provider) GetDarajaClient(endpoint string, cfg ShortCodeConfig) *daraja.Client {

	client := daraja.New(daraja.Config{Endpoint: endpoint, LogLevel: request.LogDebugWithHTTPBody})
	client.Hooks.Build.PushBackHook(daraja.Authenticate(endpoint, cfg.InitiatorName, cfg.InitiatorPassword))
	client.Hooks.Send.PushBackHook(corehooks.LogHTTPRequest)

	return &client

}

func (p Provider) GetShortCodeConfig(name string) (ShortCodeConfig, error) {
	switch name {
	case "C2B":
		return ShortCodeConfig{
			ShortCode:         "000000",
			InitiatorName:     "testapi",
			InitiatorPassword: "Safaricom999!*!%",
			Passphrase:        "bfb279f9aa9bdbcf158e97dd71a467cd2e0c893059b10f78e6b72ada1ed2c919%",
			ConsumerKey:       "GW0TvN2gUTakps3b1AbAw48no1Yogu92oXI0N55fmlEVK40e",
			ConsumerSecret:    "CtpMOjvk47jm6A5hmCzaQjQTBOWAwK1LM95awGNLSTGawbGNPsy9f8Eabsr1Lg7P",
		}, nil

	case "B2C":
		return ShortCodeConfig{
			ShortCode:         "001001",
			InitiatorName:     "testapi",
			InitiatorPassword: "Safaricom999!*!%",
			Passphrase:        "bfb279f9aa9bdbcf158e97dd71a467cd2e0c893059b10f78e6b72ada1ed2c919%",
			ConsumerKey:       "GW0TvN2gUTakps3b1AbAw48no1Yogu92oXI0N55fmlEVK40e",
			ConsumerSecret:    "CtpMOjvk47jm6A5hmCzaQjQTBOWAwK1LM95awGNLSTGawbGNPsy9f8Eabsr1Lg7P",
		}, nil
	case "B2B":
		return ShortCodeConfig{
			ShortCode:         "005005",
			InitiatorName:     "testapi",
			InitiatorPassword: "Safaricom999!*!%",
			Passphrase:        "bfb279f9aa9bdbcf158e97dd71a467cd2e0c893059b10f78e6b72ada1ed2c919%",
			ConsumerKey:       "GW0TvN2gUTakps3b1AbAw48no1Yogu92oXI0N55fmlEVK40e",
			ConsumerSecret:    "CtpMOjvk47jm6A5hmCzaQjQTBOWAwK1LM95awGNLSTGawbGNPsy9f8Eabsr1Lg7P",
		}, nil
	default:
		return ShortCodeConfig{}, nil
	}

}
