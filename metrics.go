package wallconnector

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/protobuf/proto"
)

// A prometheus.Collector implementation for wallconnector stats.
type collector struct {
	client     *Client
	metricSets map[string]metricSet
}

func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	for _, set := range c.metricSets {
		for _, metric := range set {
			ch <- metric.desc
		}
	}
}

func (c *collector) Collect(ch chan<- prometheus.Metric) {
	for _, set := range c.metricSets {
		for name, metric := range set {
			v, err := c.client.callApi(name)
			if err != nil {
				ch <- prometheus.NewInvalidMetric(metric.desc, err)
				continue
			}

			for _, value := range metric.metric.GetValue() {
				ch <- prometheus.MustNewConstMetric(
					metric.desc,
					metric.typ,
					value,
					metric.labels...,
				)
			}
		}
	}
}

// NewCollector creates a new collector for wallconnector stats.
func NewCollector(client *Client) prometheus.Collector {
	return &collector{
		client: client,
		metricSets: map[string]metricSet{
			vitalsPath: newMetricSet(new(Vitals), "vitals"),
			// newMetricSet(new(Lifetime), "lifetime"),
		},
	}
}

// Ensure we implement the [prometheus.Collector] interface
var _ prometheus.Collector = (*collector)(nil)

type metricData struct {
	typ    prometheus.ValueType
	labels []string
	desc   *prometheus.Desc
	metric *Metric
}

type metricSet map[string]metricData

func newMetricSet(v proto.Message, ns string) metricSet {
	// Parse the Metric proto message off of each field on a Vitals message.
	// It's fine to use reflection here since this is only called once at
	// startup.
	// Use proto reflect to parse the prometheus metric name and description
	// from the proto message.
	set := make(metricSet)
	descs := make(descriptions)

	desc := v.ProtoReflect().Descriptor()
	for i := 0; i < desc.Fields().Len(); i++ {
		field := desc.Fields().Get(i)
		ext, ok := proto.GetExtension(field.Options(), E_Prometheus).(*Metric)
		if !ok || ext.GetName() == "" {
			continue
		}

		metric := metricData{
			metric: ext,
			desc:   descs.getDescription(ext, ns),
			labels: ext.LabelValues(),
		}

		switch ext.GetType() {
		case Metric_COUNTER:
			metric.typ = prometheus.CounterValue
		case Metric_GAUGE:
			metric.typ = prometheus.GaugeValue
		default:
			panic("unknown metric type")
		}

		name := field.JSONName()
		set[name] = metric
	}

	return set
}

type descriptions map[string]*prometheus.Desc

func (m *Metric) LabelKeys() []string {
	keys := make([]string, 0, len(m.GetLabels()))
	for _, label := range m.GetLabels() {
		key, _, ok := strings.Cut(label, ":")
		if !ok {
			panic("invalid label " + label)
		}
		keys = append(keys, key)
	}
	return keys
}

func (m *Metric) LabelValues() []string {
	values := make([]string, 0, len(m.GetLabels()))
	for _, label := range m.GetLabels() {
		_, value, ok := strings.Cut(label, ":")
		if !ok {
			panic("invalid label " + label)
		}
		values = append(values, value)
	}
	return values
}

func (d descriptions) getDescription(v *Metric, ns string) *prometheus.Desc {
	name := prometheus.BuildFQName("wallconnector", ns, v.GetName())
	if desc, ok := d[name]; ok {
		return desc
	}

	desc := prometheus.NewDesc(
		name,
		v.GetHelp(),
		v.LabelKeys(),
		nil,
	)
	d[name] = desc
	return desc
}
