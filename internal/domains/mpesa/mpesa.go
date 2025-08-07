package mpesa

import (
	"context"
)

type ShortCodeConfig struct {
	ShortCodeID       string
	ShortCode         string // business pay bill or buy goods account
	Service           string // service can be either daraja or quikk
	InitiatorName     string // daraja api initiator name
	InitiatorPassword string // daraja api initiator password
	Passphrase        string // (optional) passphrase for c2b transfers
	Key               string // daraja app consumer key or quikk app key
	Secret            string // daraja app consumer secret or quikk app secret
	CallbackURL       string // callback url for shortcode async responses
}

type ShortCodeConfigRepository interface {
	Add(ctx context.Context, shortcode ShortCodeConfig) error
	Find(ctx context.Context, shortcodeID string) (ShortCodeConfig, error)
}
