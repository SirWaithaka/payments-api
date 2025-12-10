package sdk

type RequestAddShortCode struct {
	Environment       string `json:"environment"`
	Service           string `json:"service"`
	Type              string `json:"type"`
	ShortCode         string `json:"shortcode"`
	InitiatorName     string `json:"initiator_name"`
	InitiatorPassword string `json:"initiator_password"`
	Key               string `json:"key"`
	Secret            string `json:"secret"`
	Passphrase        string `json:"passphrase"`
}
