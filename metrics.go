package wallconnector

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
)

// A prometheus.Collector implementation for wallconnector stats.
type collector struct {
	timeout    time.Duration
	client     *Client
	metricSets []metricFetcher
}

func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	for _, set := range c.metricSets {
		set.Describe(ch)
	}
}

func (c *collector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()
	if c.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.timeout)
		defer cancel()
	}
	wait := sync.WaitGroup{}
	for _, set := range c.metricSets {
		wait.Add(1)
		set := set
		go func() {
			set.Collect(ctx, ch)
			wait.Done()
		}()
	}
	wait.Wait()
}

// NewCollector creates a new collector for wallconnector stats.
func NewCollector(client *Client) prometheus.Collector {
	return &collector{
		client: client,
		metricSets: []metricFetcher{
			newMetricSet("vitals", client.Vitals),
			newMetricSet("lifetime", client.Lifetime),
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

type metricFetcher interface {
	Describe(ch chan<- *prometheus.Desc)
	Collect(ctx context.Context, ch chan<- prometheus.Metric)
}

// Mapping of metric JSONName to metric data for a particular endpoint.
type metricSet[T proto.Message] struct {
	metrics map[string]metricData
	fetcher func(context.Context) (T, error)
}

func (m *metricSet[T]) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range m.metrics {
		ch <- metric.desc
	}
}

func (m *metricSet[T]) Collect(ctx context.Context, ch chan<- prometheus.Metric) {
	logger := log.Default()
	v, err := m.fetcher(ctx)
	if err != nil {
		return
	}
	v.ProtoReflect().Range(func(field protoreflect.FieldDescriptor, value protoreflect.Value) bool {
		metric, ok := m.metrics[field.JSONName()]
		if !ok {
			// Ignore unsupported types.
			logger.Printf("unknown key %s(%s)", field.Kind(), field.JSONName())
			return true
		}
		var val float64
		switch field.Kind() {
		case protoreflect.FloatKind, protoreflect.DoubleKind:
			val = value.Float()
		case protoreflect.Int32Kind, protoreflect.Int64Kind:
			val = float64(value.Int())
		default:
			// Ignore unsupported types.
			logger.Printf("unsupported type %s(%s)", field.Kind(), field.JSONName())
			return true
		}

		ch <- prometheus.MustNewConstMetric(
			metric.desc,
			metric.typ,
			val,
			metric.labels...,
		)
		return true
	})
}

func newMetricSet[T proto.Message](ns string, fetcher func(context.Context) (T, error)) metricFetcher {
	// Parse the Metric proto message off of each field on a Vitals message.
	// It's fine to use reflection here since this is only called once at
	// startup.
	// Use proto reflect to parse the prometheus metric name and description
	// from the proto message.
	set := make(map[string]metricData)
	descs := make(descriptions)

	var v T
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

	return &metricSet[T]{
		metrics: set,
		fetcher: fetcher,
	}
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
