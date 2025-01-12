package corehooks

import (
	"fmt"
	"regexp"

	"github.com/SirWaithaka/payments-api/request"
)

var schemeRE = regexp.MustCompile("^([^:]+)://")

// AddScheme adds the HTTP or HTTPS schemes to an endpoint URL if there is no
// scheme. If disableSSL is true HTTP will set HTTP instead of the default HTTPS.
//
// If disableSSL is set, it will only set the URL's scheme if the URL does not
// contain a scheme.
func AddScheme(endpoint string, disableSSL bool) string {
	if !schemeRE.MatchString(endpoint) {
		scheme := "https"
		if disableSSL {
			scheme = "http"
		}
		endpoint = fmt.Sprintf("%s://%s", scheme, endpoint)
	}

	return endpoint
}

var ResolveEndpoint = request.Hook{Fn: func(r *request.Request) {
	r.Config.Endpoint = AddScheme(r.Config.Endpoint, r.Config.DisableSSL)
}}
