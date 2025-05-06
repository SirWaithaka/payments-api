package daraja

import (
	"log"

	"github.com/SirWaithaka/payments-api/request"
)

// SetBaseUrl checks the environment passed and uses the appropriate base url
// otherwise if the baseUrl is not empty, it sets the Request endpoint to it.
//
// Should be used as a build hook
func SetBaseUrl(baseUrl, environment string) request.Hook {
	// set default base url as the sandbox url
	url := sandboxUrl

	if environment == ENV_PRODUCTION {
		url = productionUrl
	}

	// lastly, if base url is not empty, use that instead
	if baseUrl != "" {
		url = baseUrl
	}

	return request.Hook{Fn: func(r *request.Request) {
		log.Println("SetBaseUrl", url)
		r.Config.Endpoint = url
	}}
}
