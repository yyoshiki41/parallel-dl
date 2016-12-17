package paralleldl

import (
	"errors"
	"net/http"
	"time"
)

var httpClient = &http.Client{}

// Options represents Client's parameters.
type Options struct {
	// Output directory
	Output string

	// http.Client's Timeout.
	// A Timeout of zero means no timeout.
	Timeout time.Duration

	// Maximum number of concurrent requests.
	// A MaxConcurrents of zero means no limit of number of requests.
	MaxConcurrents int64
	// Maximum number of errors before giving up the whole requests.
	// A MaxErrorRequests of zero means try until all requests are done.
	MaxErrorRequests int64

	// Maximum number of retries before giving up a request.
	// A MaxAttempts of zero means try until a request is success.
	MaxAttempts int64
}

// Client represents http.Client with paralleldl options.
type Client struct {
	httpClient *http.Client
	opt        *Options
}

// New returns a new Client struct.
func New(opt *Options) (*Client, error) {
	if httpClient == nil {
		return nil, errors.New("httpClient is nil")
	}
	httpClient.Timeout = opt.Timeout

	return &Client{
		httpClient: httpClient,
		opt:        opt,
	}, nil
}

// SetHTTPClient overrides the default HTTP client.
func SetHTTPClient(client *http.Client) {
	httpClient = client
}
