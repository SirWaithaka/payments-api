package request

import (
	"net/http"
)

type Config struct {
	// Endpoint is hostname or fully qualified URI of the service being called
	Endpoint string

	// Name of the external service being called
	ServiceName string

	// Set this to `true` to disable SSL when sending requests. Defaults
	// to `false`
	DisableSSL bool

	// The HTTP client to use when sending requests
	HTTPClient *http.Client

	DisableFollowRedirects bool

	LogLevel LogLevel
	// The logger writer interface to write logging messages to. Defaults to
	// standard out.
	Logger Logger

	// Unique ID to trace a request attempt
	RequestID string
}
