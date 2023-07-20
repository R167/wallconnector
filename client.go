package wallconnector

import (
	"encoding/json"
	"io"
	"net/http"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

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

func callApi[T any](c *Client, path string) (*T, error) {
	resp, err := c.client.Get("http://" + c.addr + path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	v := new(T)

	switch v.(type) {
	case proto.Message:
		err = protojson.Unmarshal(data, v)
	default:
		err = json.Unmarshal(data, v)
	}

	if err := json.Unmarshal(data, v); err != nil {
		return nil, err
	}
	return v, nil
}

// Vitals returns the current vitals of the wallconnector.
func (c *Client) Vitals() (*Vitals, error) {
	return callApi[Vitals](c, vitalsPath)
}

// Lifetime returns the lifetime stats of the wallconnector.
func (c *Client) Lifetime() (*Lifetime, error) {
	return callApi[Lifetime](c, lifetimePath)
}

// Version returns the version info of the wallconnector.
func (c *Client) Version() (*Version, error) {
	return callApi[Version](c, versionPath)
}

// Lifetime represents the lifetime stats of the wallconnector.
//
// See Wall Monitor FAQ for more details:
// https://wallmonitor.app/faq/explain_lifetime
type Lifetime struct {
	ContactorCycles       int     `json:"contactor_cycles"`
	ContactorCyclesLoaded int     `json:"contactor_cycles_loaded"`
	AlertCount            int     `json:"alert_count"`
	ThermalFoldbacks      int     `json:"thermal_foldbacks"`
	AvgStartupTemp        float64 `json:"avg_startup_temp"` // Is this a float?
	ChargeStarts          int     `json:"charge_starts"`
	EnergyWh              int     `json:"energy_wh"`
	ConnectorCycles       int     `json:"connector_cycles"`
	UptimeS               int     `json:"uptime_s"`
	ChargingTimeS         int     `json:"charging_time_s"`
}

// Version represents the version info of the wallconnector.
type Version struct {
	FirmwareVersion string `json:"firmware_version"`
	PartNumber      string `json:"part_number"`
	SerialNumber    string `json:"serial_number"`
}
