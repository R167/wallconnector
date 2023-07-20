package wallconnector

import (
	"encoding/json"
	"io"
	"net/http"
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

// Vitals represents the current state of the wallconnector.
//
// Note: Several fields like SessionS, and UptimeS are probably ints, but
// they're it's impossible to say parsing JSON from the API.
//
// See Wall Monitor FAQ for more details:
// https://wallmonitor.app/faq/explain_technical
type Vitals struct {
	ContactorClosed   bool    `json:"contactor_closed" wc:"contactor_closed_state,gauge"`
	VehicleConnected  bool    `json:"vehicle_connected" wc:"vehicle_connected_state,gauge"`
	SessionS          float64 `json:"session_s" wc:"session_seconds,counter"`
	GridV             float64 `json:"grid_v" wc:"grid_volts,gauge"`
	GridHz            float64 `json:"grid_hz" wc:"grid_hertz,gauge"`
	VehicleCurrentA   float64 `json:"vehicle_current_a" wc:"-"`
	CurrentA_a        float64 `json:"currentA_a" wc:"-"`
	CurrentB_a        float64 `json:"currentB_a" wc:"-"`
	CurrentC_a        float64 `json:"currentC_a" wc:"-"`
	CurrentN_a        float64 `json:"currentN_a" wc:"-"`
	VoltageA_v        float64 `json:"voltageA_v" wc:"-"`
	VoltageB_v        float64 `json:"voltageB_v" wc:"-"`
	VoltageC_v        float64 `json:"voltageC_v" wc:"-"`
	RelayCoilV        float64 `json:"relay_coil_v" wc:",gauge"`
	PcbaTempC         float64 `json:"pcba_temp_c" wc:",gauge"`
	HandleTempC       float64 `json:"handle_temp_c" wc:",gauge"`
	McuTempC          float64 `json:"mcu_temp_c" wc:",gauge"`
	UptimeS           float64 `json:"uptime_s" wc:",counter"`
	InputThermopileUv float64 `json:"input_thermopile_uv" wc:",gauge"`
	ProxV             float64 `json:"prox_v" wc:",gauge"`
	PilotHighV        float64 `json:"pilot_high_v" wc:",gauge"`
	PilotLowV         float64 `json:"pilot_low_v" wc:"pilot_low_volts,gauge"`
	SessionEnergyWh   float64 `json:"session_energy_wh" wc:"session_energy_joules,counter"`
	ConfigStatus      int     `json:"config_status"`
	EvseState         int     `json:"evse_state"`
	CurrentAlerts     []int   `json:"current_alerts"`
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
