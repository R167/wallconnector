package wallconnector

import (
	"context"
	"fmt"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestParsingMetrics(t *testing.T) {
	dummy := func(context.Context) (*Vitals, error) {
		return nil, nil
	}

	metrics := newMetricSet("vitals", dummy)

	ch := make(chan *prometheus.Desc)
	go func() {
		metrics.Describe(ch)
		close(ch)
	}()

	i := 0
	for desc := range ch {
		fmt.Println(desc.String())
		i++
	}
	assert.Equal(t, 25, i)
}
