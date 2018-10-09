package logstash

import (
	"fmt"
	"strings"
	"time"

	"github.com/rcrowley/go-metrics"
)

//Measure is a map that represents the contract to be send to measures
type Measure map[string]interface{}

//AddTimer adds a timer to the structure
func (m *Measure) AddTimer(timer metrics.Timer, percentiles []float64) {
	ms := timer.Snapshot()
	t := make(map[string]interface{})

	t["count"] = ms.Count()
	t["max"] = ms.Max()
	t["min"] = ms.Min()
	t["mean"] = ms.Mean()
	t["stddev"] = ms.StdDev()
	t["var"] = ms.Variance()

	for _, p := range percentiles {
		duration := time.Duration(ms.Percentile(p)).Seconds() * 1000
		pStr := strings.Replace(fmt.Sprintf("p%g", p*100), ".", "_", -1)
		t[pStr] = duration
	}

	(*m)["kind"] = "timer"
	(*m)["timer"] = t
}

//AddMeter adds a meter to the structure
func (m *Measure) AddMeter(meter metrics.Meter) {
	ms := meter.Snapshot()
	me := make(map[string]interface{})

	me["count"] = ms.Count()
	me["rate1"] = ms.Rate1()
	me["rate5"] = ms.Rate5()
	me["rate15"] = ms.Rate15()
	me["mean"] = ms.RateMean()

	(*m)["kind"] = "meter"
	(*m)["meter"] = me
}

//AddHistogram adds a histogram to the structure
func (m *Measure) AddHistogram(histogram metrics.Histogram, percentiles []float64) {
	ms := histogram.Snapshot()

	h := make(map[string]interface{})

	h["count"] = ms.Count()
	h["max"] = ms.Max()
	h["min"] = ms.Min()
	h["mean"] = ms.Mean()
	h["stddev"] = ms.StdDev()
	h["var"] = ms.Variance()

	for _, p := range percentiles {
		pStr := strings.Replace(fmt.Sprintf("p%g", p*100), ".", "_", -1)
		h[pStr] = ms.Percentile(p)
	}

	(*m)["kind"] = "histogram"
	(*m)["histogram"] = h
}

//AddGaugeFloat64 adds an int gauge64 to the structure
func (m *Measure) AddGaugeFloat64(gauge metrics.GaugeFloat64) {
	(*m)["kind"] = "gauge64"
	(*m)["gauge64"] = gauge.Value()
}

//AddGauge adds an int gauge to the structure
func (m *Measure) AddGauge(gauge metrics.Gauge) {
	(*m)["kind"] = "gauge"
	(*m)["gauge"] = gauge.Value()
}

//AddCounter adds a counter to the structure
func (m *Measure) AddCounter(counter metrics.Counter) {
	(*m)["kind"] = "counter"
	(*m)["counter"] = counter.Count()
}

//NewMeasure creates a new Measure and adds name and default information to the structure
func NewMeasure(name string, defaultValues map[string]interface{}) *Measure {
	m := Measure{}

	for i, element := range strings.Split(name, ".") {
		m[fmt.Sprintf("identifier%d", i)] = element
	}

	for k, v := range defaultValues {
		m[k] = v
	}

	return &m
}
