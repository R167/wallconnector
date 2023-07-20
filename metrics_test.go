package wallconnector

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsingMetrics(t *testing.T) {
	metrics := newMetricSet(new(Vitals), "vitals")
	assert.Len(t, metrics, 25)

	for _, m := range metrics {
		fmt.Printf("%+v, %v\n", m, m.desc)
	}
}
