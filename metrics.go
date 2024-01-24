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
	"google.golang.org/protobuf/types/descriptorpb"
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
			newMetricSet("wifi", client.Wifi),
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
		if metric.metric.GetSkip() {
			continue
		}
		ch <- metric.desc
	}
}

func (m *metricSet[T]) Collect(ctx context.Context, ch chan<- prometheus.Metric) {
	logger := log.Default()
	v, err := m.fetcher(ctx)
	if err != nil {
		return
	}
	fields := v.ProtoReflect().Descriptor().Fields()
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		metric, ok := m.metrics[field.JSONName()]
		if !ok {
			// Ignore unsupported types.
			logger.Printf("unknown key %s(%s)", field.JSONName(), field.Kind())
			continue
		}
		if metric.metric.GetSkip() {
			continue
		}

		value := v.ProtoReflect().Get(field)

		var val float64
		switch field.Kind() {
		case protoreflect.FloatKind, protoreflect.DoubleKind:
			val = value.Float()
		case protoreflect.Int32Kind, protoreflect.Int64Kind:
			val = float64(value.Int())
		case protoreflect.BoolKind:
			val = 0
			if value.Bool() {
				val = 1
			}
		default:
			// Ignore unsupported types.
			logger.Printf("unsupported type %s(%s)", field.JSONName(), field.Kind())
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			metric.desc,
			metric.typ,
			metric.metric.ConvertValue(val),
			metric.labels...,
		)
	}
}

func newMetricSet[T proto.Message](ns string, fetcher func(context.Context) (T, error)) metricFetcher {
	set := make(map[string]metricData)
	descs := make(descriptions)

	var v T
	desc := v.ProtoReflect().Descriptor()
	for i := 0; i < desc.Fields().Len(); i++ {
		field := desc.Fields().Get(i)
		opts := field.Options().(*descriptorpb.FieldOptions)
		ext, ok := proto.GetExtension(opts, E_Prometheus).(*Metric)
		if !ok || ext.GetName() == "" {
			continue
		}
		if ext.GetSkip() {
			set[field.JSONName()] = metricData{metric: ext}
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

const (
	whToJoules = 3600
)

func (m *Metric) ConvertValue(v float64) float64 {
	switch m.GetConversion() {
	case Conversion_NONE:
		return v
	case Conversion_INVERSE:
		return 1 / v
	case Conversion_WH_TO_J:
		return v * whToJoules
	default:
		panic("unknown conversion")
	}
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
