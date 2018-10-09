package logstash

import (
	"testing"

	"github.com/stretchr/testify/assert"

	metrics "github.com/rcrowley/go-metrics"
)

var percentiles = []float64{0.50, 0.75, 0.95, 0.99, 0.999}

func TestAddCounter(t *testing.T) {
	m := Measure{}
	registry := metrics.NewRegistry()

	expectedValue := Measure{
		"kind":    "counter",
		"counter": int64(8),
	}

	counter := metrics.GetOrRegisterCounter("dummy-counter", registry)
	counter.Inc(8)

	m.AddCounter(counter)
	assert.Equal(t, expectedValue, m)
}

func TestAddGauge(t *testing.T) {
	m := Measure{}
	registry := metrics.NewRegistry()

	expectedValue := Measure{
		"kind":  "gauge",
		"gauge": int64(8),
	}

	gauge := metrics.GetOrRegisterGauge("dummy-gauge", registry)
	gauge.Update(8)

	m.AddGauge(gauge)

	assert.Equal(t, expectedValue, m)
}

func TestAddGauge64(t *testing.T) {
	m := Measure{}
	registry := metrics.NewRegistry()
	expectedValue := Measure{
		"kind":    "gauge64",
		"gauge64": float64(7.7),
	}

	gauge64 := metrics.GetOrRegisterGaugeFloat64("dummy-gauge64", registry)
	gauge64.Update(7.7)

	m.AddGaugeFloat64(gauge64)

	assert.Equal(t, expectedValue, m)
}

func TestAddHistogram(t *testing.T) {
	m := Measure{}
	expectedValueValue := Measure{
		"kind": "histogram",
		"histogram": map[string]interface{}{
			"count":  int64(0),
			"max":    int64(0),
			"min":    int64(0),
			"mean":   float64(0),
			"stddev": float64(0),
			"var":    float64(0),
			"p50":    0.0,
			"p75":    0.0,
			"p95":    0.0,
			"p99":    0.0,
			"p99_9":  0.0,
		},
	}

	histogram := metrics.NilHistogram{}
	histogram.Update(10)

	m.AddHistogram(histogram, percentiles)

	assert.Equal(t, expectedValueValue, m)
}

func TestAddTimer(t *testing.T) {
	m := Measure{}
	expectedValue := Measure{
		"kind": "timer",
		"timer": map[string]interface{}{
			"count":  int64(0),
			"max":    int64(0),
			"min":    int64(0),
			"mean":   float64(0),
			"stddev": float64(0),
			"var":    float64(0),
			"p50":    0.0,
			"p75":    0.0,
			"p95":    0.0,
			"p99":    0.0,
			"p99_9":  0.0,
		},
	}
	timer := metrics.NilTimer{}

	m.AddTimer(timer, percentiles)

	assert.Equal(t, expectedValue, m)
}

func TestAddMeter(t *testing.T) {
	m := Measure{}
	expectedValue := Measure{
		"kind": "meter",
		"meter": map[string]interface{}{
			"count":  int64(0),
			"rate1":  float64(0),
			"rate5":  float64(0),
			"rate15": float64(0),
			"mean":   float64(0),
		},
	}

	meter := metrics.NilMeter{}
	m.AddMeter(meter)

	assert.Equal(t, expectedValue, m)
}
