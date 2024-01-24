package wallconnector

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

//go:generate protoc --go_out=. --go_opt=paths=source_relative metrics.proto

const (
	vitalsPath   = "/api/1/vitals"
	lifetimePath = "/api/1/lifetime"
	versionPath  = "/api/1/version"
	wifiPath     = "/api/1/wifi_status"
)

type Client struct {
	addr   string
	client *http.Client
}

func NewClient(addr string, opts ...ConnectorConfig) (*Client, error) {
	c := &connectorOpts{
		Transport: http.DefaultTransport,
	}
	for _, opt := range opts {
		opt(c)
	}

	return &Client{
		addr: addr,
		client: &http.Client{
			Transport: c.Transport,
			Timeout:   c.Timeout,
		},
	}, nil
}

func callApi[T any](ctx context.Context, c *Client, path string) (*T, error) {
	req, err := http.NewRequest(http.MethodGet, "http://"+c.addr+path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	v := new(T)

	if err := json.Unmarshal(data, v); err != nil {
		return nil, err
	}
	return v, nil
}

// Vitals returns the current vitals of the wallconnector.
func (c *Client) Vitals(ctx context.Context) (*Vitals, error) {
	return callApi[Vitals](ctx, c, vitalsPath)
}

// Lifetime returns the lifetime stats of the wallconnector.
func (c *Client) Lifetime(ctx context.Context) (*Lifetime, error) {
	return callApi[Lifetime](ctx, c, lifetimePath)
}

// Version returns the version info of the wallconnector.
func (c *Client) Version(ctx context.Context) (*Version, error) {
	return callApi[Version](ctx, c, versionPath)
}

// WifiStatus returns the wifi status of the wallconnector.
func (c *Client) Wifi(ctx context.Context) (*Wifi, error) {
	return callApi[Wifi](ctx, c, wifiPath)
}
