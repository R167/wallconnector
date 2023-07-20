package wallconnector

import (
	"net/http"
	"time"
)

type ConnectorConfig func(*connectorOpts)

type connectorOpts struct {
	// http.RoundTripper to use for requests to the wallconnector API.
	Transport http.RoundTripper

	// Timeout for requests to the wallconnector API.
	Timeout time.Duration
}

func WithTransport(t http.RoundTripper) func(*connectorOpts) {
	return func(opts *connectorOpts) {
		opts.Transport = t
	}
}

func WithTimeout(t time.Duration) func(*connectorOpts) {
	return func(opts *connectorOpts) {
		opts.Timeout = t
	}
}
