package wallconnector

import (
	"github.com/prometheus/client_golang/prometheus"
)

// A prometheus.Collector implementation for wallconnector stats.
type vitalsCollector struct {
	client *Client
}

func (c *vitalsCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

var (
	// Vitals
	vitalsCurrentAmps = prometheus.NewDesc(
		"current_amps",
		"Wall connector output in amps",
		[]string{"phase"},
		nil,
	)
	vitalsCurrentVolts = prometheus.NewDesc(
		"current_volts",
		"Wall connector output in volts",
		[]string{"phase"},
		nil,
	)
	vitalsGridFrequency = prometheus.NewDesc(
		"grid_frequency",
		"Grid frequency in hertz",
		nil,
		nil,
	)
	vitalsGridVoltage = prometheus.NewDesc(
		"grid_voltage",
		"Grid voltage in volts",
		nil,
		nil,
	)
	vitalsSessionSeconds = prometheus.NewDesc(
		"session_seconds",
		"Duration of the most recent charging session in seconds",
		nil,
		nil,
	)
	vitalsUptimeSeconds = prometheus.NewDesc(
		"uptime_seconds",
		"Duration of the wall connector uptime in seconds",
		nil,
		nil,
	)
)

func (c *vitalsCollector) Collect(ch chan<- prometheus.Metric) {
	if data, err := c.client.Vitals(); err != nil {
		ch <- prometheus.MustNewConstMetric(vitalsCurrentAmps, prometheus.GaugeValue, data.CurrentA_a, "A")
		ch <- prometheus.MustNewConstMetric(vitalsCurrentAmps, prometheus.GaugeValue, data.CurrentB_a, "B")
		ch <- prometheus.MustNewConstMetric(vitalsCurrentAmps, prometheus.GaugeValue, data.CurrentC_a, "C")
		ch <- prometheus.MustNewConstMetric(vitalsCurrentAmps, prometheus.GaugeValue, data.CurrentN_a, "N")
		ch <- prometheus.MustNewConstMetric(vitalsCurrentVolts, prometheus.GaugeValue, data.VoltageA_v, "A")
		ch <- prometheus.MustNewConstMetric(vitalsCurrentVolts, prometheus.GaugeValue, data.VoltageB_v, "B")
		ch <- prometheus.MustNewConstMetric(vitalsCurrentVolts, prometheus.GaugeValue, data.VoltageC_v, "C")
		ch <- prometheus.MustNewConstMetric(vitalsGridFrequency, prometheus.GaugeValue, data.GridHz)
		ch <- prometheus.MustNewConstMetric(vitalsGridVoltage, prometheus.GaugeValue, data.GridV)
		ch <- prometheus.MustNewConstMetric(vitalsSessionSeconds, prometheus.CounterValue, data.SessionS)
		ch <- prometheus.MustNewConstMetric(vitalsUptimeSeconds, prometheus.CounterValue, data.UptimeS)
	}
}

// NewCollector creates a new collector for wallconnector stats.
func NewVitalsCollector(client *Client) prometheus.Collector {
	return &vitalsCollector{
		client: client,
	}
}

// Ensure we implement the [prometheus.Collector] interface
var _ prometheus.Collector = (*vitalsCollector)(nil)
