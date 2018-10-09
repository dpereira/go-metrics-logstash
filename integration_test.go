package logstash

import (
	"net"
	"strconv"
	"strings"
	"testing"

	metrics "github.com/rcrowley/go-metrics"
	"github.com/stretchr/testify/assert"
)

type UDPServer struct {
	conn *net.UDPConn
}

func newUDPServer(port int) (*UDPServer, error) {
	serverAddr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		return nil, err
	}
	return &UDPServer{conn}, nil
}

func (us *UDPServer) Read() (string, error) {
	buffer := make([]byte, 4096)
	_, _, err := us.conn.ReadFromUDP(buffer)
	if err != nil {
		return "", err
	}
	resizedStr := strings.Trim(string(buffer), "\x00") // Remove the empty chars at the end of the buffer
	return resizedStr, nil
}

func (us *UDPServer) Close() {
	us.conn.Close()
}

func TestFlushOnceKeepsPreviousValues(t *testing.T) {
	serverAddr := "localhost:1984"
	server, err := newUDPServer(1984)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()

	registry := metrics.NewRegistry()
	reporter, err := NewReporter(registry, serverAddr, map[string]interface{}{
		"client": "dummy-client",
		"metric": "doc",
	})
	assert.NoError(t, err)

	metrics.GetOrRegisterCounter("dummycounter", registry).Inc(6)
	metrics.GetOrRegisterGauge("dummygauge", registry).Update(8)
	err = reporter.FlushOnce()
	assert.NoError(t, err)

	receivedCounter, err := server.Read()
	if err != nil {
		t.Fatal(err)
	}
	receivedGauge, err := server.Read()
	if err != nil {
		t.Fatal(err)
	}

	expectedCounter := `{
		"identifier0":"dummycounter",
		"client": "dummy-client",
		"metric": "doc",
		"kind": "counter",
		"counter": 6
	}`
	expectedGauge := `{
		"identifier0":"dummygauge",
		"client": "dummy-client",
		"metric": "doc",
		"kind": "gauge",
		"gauge": 8
	}`
	assert.JSONEq(t, expectedCounter, receivedCounter)
	assert.JSONEq(t, expectedGauge, receivedGauge)
}

func TestFlushOnceWithDefaultValues(t *testing.T) {
	serverAddr := "localhost:1984"
	server, err := newUDPServer(1984)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()

	registry := metrics.NewRegistry()
	reporter, err := NewReporter(registry, serverAddr, map[string]interface{}{
		"client": "dummy-client",
		"metric": "doc",
	})
	assert.NoError(t, err)

	// Insert metrics
	metrics.GetOrRegisterCounter("test_counter", registry).Inc(6)
	err = reporter.FlushOnce()
	assert.NoError(t, err)

	received, err := server.Read()
	if err != nil {
		t.Fatal(err)
	}

	expected := `{
		"identifier0":"test_counter",
		"client": "dummy-client",
		"metric": "doc",
		"kind": "counter",
		"counter": 6
	}`
	assert.JSONEq(t, expected, received)
}
func TestFlushOnceWithDotSeparatedIndetifiers(t *testing.T) {
	serverAddr := "localhost:1984"
	server, err := newUDPServer(1984)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()

	registry := metrics.NewRegistry()
	reporter, err := NewReporter(registry, serverAddr, map[string]interface{}{
		"client": "dummy-client",
		"metric": "doc",
	})
	assert.NoError(t, err)

	// Insert metrics
	metrics.GetOrRegisterCounter("test_counter.i1.i2.i3", registry).Inc(8)
	err = reporter.FlushOnce()
	assert.NoError(t, err)

	received, err := server.Read()
	if err != nil {
		t.Fatal(err)
	}

	expected := `{
		"identifier0":"test_counter",
		"identifier1":"i1",
		"identifier2":"i2",
		"identifier3":"i3",
		"client": "dummy-client",
		"metric": "doc",
		"kind": "counter",
		"counter": 8
	}`
	assert.JSONEq(t, expected, received)
}

func TestFlushOnceReturnsConnectionError(t *testing.T) {
	serverAddr := "localhost:1984"

	registry := metrics.NewRegistry()
	reporter, err := NewReporter(registry, serverAddr, nil)
	assert.NoError(t, err)

	// Insert metrics
	metrics.GetOrRegisterCounter("test_counter", registry).Inc(6)

	reporter.Conn.Close()
	err = reporter.FlushOnce()
	assert.Error(t, err)
}
