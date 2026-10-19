package server

import (
	"fmt"
	"log"
	"time"

	influxdb "github.com/influxdata/influxdb-client-go/v2"
	api "github.com/influxdata/influxdb-client-go/v2/api"
)

const (
	// influxFlushInterval is how often points are sent to the database.
	// Each flush creates one batch in the client's retry queue.
	influxFlushInterval = time.Second
	// influxBatchSize is the maximum number of points per write request
	influxBatchSize = 5_000
	// influxMaxRetryInterval caps the backoff between retries, so data
	// is written soon after the database becomes available again.
	influxMaxRetryInterval = 30 * time.Second
)

// Influx is an InfluxDB v2 publisher
type Influx struct {
	client      influxdb.Client
	writer      api.WriteAPI
	measurement string
}

// NewInfluxClient creates new publisher for influx
func NewInfluxClient(
	url string,
	database string,
	measurement string,
	org string,
	token string,
	user string,
	password string,
	buffer time.Duration,
) *Influx {
	// InfluxDB v1 compatibility
	if token == "" && user != "" {
		token = fmt.Sprintf("%s:%s", user, password)
	}

	client := influxdb.NewClientWithOptions(url, token, influxOptions(buffer))

	if database == "" {
		log.Fatal("influx: missing database")
	}
	if measurement == "" {
		log.Fatal("influx: missing measurement")
	}

	return &Influx{
		client:      client,
		measurement: measurement,
		writer:      client.WriteAPI(org, database),
	}
}

// influxOptions configures the client's in-memory retry buffer so that
// points are kept for the given duration while the database is unreachable.
//
// The client's defaults drop data almost immediately: the retry buffer
// limit is divided by the batch size (50,000 / 5,000 = 10 batches), and with
// one batch per flush interval that only holds ~10s of data. In addition,
// batches expire after 3 minutes and are discarded after 5 retries.
//
// A buffer of 0 disables retrying, so failed writes are dropped.
func influxOptions(buffer time.Duration) *influxdb.Options {
	opts := influxdb.DefaultOptions().
		SetFlushInterval(uint(influxFlushInterval.Milliseconds())).
		SetBatchSize(influxBatchSize)

	if buffer <= 0 {
		return opts.SetMaxRetries(0)
	}

	// one batch is created per flush interval
	batches := uint(buffer/influxFlushInterval) + 1

	return opts.
		// the limit is given in points but enforced in batches (limit / batch size)
		SetRetryBufferLimit(batches * influxBatchSize).
		// batches are dropped once older than this
		SetMaxRetryTime(uint(buffer.Milliseconds())).
		SetMaxRetryInterval(uint(influxMaxRetryInterval.Milliseconds())).
		// retries are bounded by MaxRetryTime, not by count. Retry intervals
		// are at least 5s, so this can not be reached before batches expire.
		SetMaxRetries(uint(buffer / time.Second))
}

// Run Influx publisher
func (m *Influx) Run(in <-chan QuerySnip) {
	// log errors
	go func() {
		for err := range m.writer.Errors() {
			log.Printf("influxdb error: %v", err)
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
		m.writer.WritePoint(p)
	}

	m.client.Close()
}
