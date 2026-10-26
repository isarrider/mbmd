package server

import (
	"fmt"
	"log"
	"math"
	"time"

	influxdb "github.com/influxdata/influxdb-client-go/v2"
	influxlog "github.com/influxdata/influxdb-client-go/v2/log"
)

// Influx is an InfluxDB v2 publisher
type Influx struct {
	client      influxdb.Client
	org         string
	database    string
	measurement string
}

// NewInfluxClient creates new publisher for influx.
// maxRetries: 0 means retry indefinitely.
// retryBufferLimit: max number of points to buffer while DB is unavailable.
// maxRetryInterval: cap on exponential back-off between retries, in milliseconds.
func NewInfluxClient(
	url string,
	database string,
	measurement string,
	org string,
	token string,
	user string,
	password string,
	maxRetries uint,
	retryBufferLimit uint,
	maxRetryInterval uint,
) *Influx {
	// InfluxDB v1 compatibility
	if token == "" && user != "" {
		token = fmt.Sprintf("%s:%s", user, password)
	}

	if database == "" {
		log.Fatal("influx: missing database")
	}
	if measurement == "" {
		log.Fatal("influx: missing measurement")
	}

	// 0 → retry indefinitely (use the largest representable uint)
	retries := maxRetries
	if retries == 0 {
		retries = math.MaxUint32
	}

	options := influxdb.DefaultOptions().
		SetPrecision(time.Second).
		SetMaxRetries(retries).
		SetMaxRetryTime(math.MaxUint32).       // no aggregate retry timeout
		SetMaxRetryInterval(maxRetryInterval). // ms; cap per-attempt back-off
		SetRetryBufferLimit(retryBufferLimit)  // max queued points

	client := influxdb.NewClientWithOptions(url, token, options)

	// suppress the default influx-client logger; errors surface via writer.Errors()
	influxlog.Log = nil

	return &Influx{
		client:      client,
		org:         org,
		database:    database,
		measurement: measurement,
	}
}

// Run Influx publisher
func (m *Influx) Run(in <-chan QuerySnip) {
	// WriteAPI is non-blocking: points are queued and flushed in the background,
	// with automatic retry when the DB is temporarily unavailable.
	writer := m.client.WriteAPI(m.org, m.database)

	// log async write errors
	go func() {
		for err := range writer.Errors() {
			log.Printf("influx: write error: %v", err)
		}
	}()

	for snip := range in {
		tags := map[string]string{
			"device": snip.Device,
			"type":   snip.Measurement.String(),
		}

		fields := map[string]any{
			"value": snip.Value,
		}

		// write asynchronously
		p := influxdb.NewPoint(m.measurement, tags, fields, time.Now())
		writer.WritePoint(p)
	}

	m.client.Close()
}
