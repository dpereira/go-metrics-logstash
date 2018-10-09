package logstash

import (
	"encoding/json"
	"log"
	"net"
	"time"

	metrics "github.com/rcrowley/go-metrics"
)

// Reporter represents a metrics registry.
type Reporter struct {
	// Registry map is used to hold metrics that will be sent to logstash.
	Registry metrics.Registry
	// Conn is a UDP connection to logstash.
	Conn *net.UDPConn
	// DefaultValues are the values that will be sent in all submits.
	DefaultValues map[string]interface{}
	Version       string
	// Percentiles to be sent on histograms and timers
	Percentiles []float64
}

// NewReporter creates a new Reporter for the register r, with an UDP client to
// the given logstash address addr and with the given default values. If defaultValues
// is nil, only the metrics will be sent.
func NewReporter(r metrics.Registry, addr string, defaultValues map[string]interface{}) (*Reporter, error) {
	if r == nil {
		r = metrics.DefaultRegistry
	}

	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUDP("udp4", nil, udpAddr)
	if err != nil {
		return nil, err
	}

	return &Reporter{
		Conn:          conn,
		Registry:      r,
		DefaultValues: defaultValues,
		Version:       "1.0.1",

		Percentiles: []float64{0.50, 0.75, 0.95, 0.99, 0.999},
	}, nil
}

// FlushEach is a blocking exporter function which reports metrics in the registry.
// Designed to be used in a goroutine: go reporter.FlushEach()
func (r *Reporter) FlushEach(interval time.Duration) {
	defer func() {
		if rec := recover(); rec != nil {
			handlePanic(rec)
		}
	}()

	for range time.Tick(interval) {
		if err := r.FlushOnce(); err != nil {
			log.Println(err)
		}
	}
}

// FlushOnce submits a snapshot of the registry.
func (r *Reporter) FlushOnce() error {
	var measures []*Measure

	r.Registry.Each(func(name string, i interface{}) {
		measure := NewMeasure(name, r.DefaultValues)
		switch metric := i.(type) {
		case metrics.Counter:
			measure.AddCounter(metric)

		case metrics.Gauge:
			measure.AddGauge(metric)

		case metrics.GaugeFloat64:
			measure.AddGaugeFloat64(metric)

		case metrics.Histogram:
			measure.AddHistogram(metric, r.Percentiles)

		case metrics.Meter:
			measure.AddMeter(metric)

		case metrics.Timer:
			measure.AddTimer(metric, r.Percentiles)
		}
		measures = append(measures, measure)
	})

	for _, measure := range measures {
		data, err := json.Marshal(measure)
		if err == nil {
			_, err := r.Conn.Write(data)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
