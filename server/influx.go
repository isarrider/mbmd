package server

import (
	"fmt"
	"log"
	"math"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	influxlog "github.com/influxdata/influxdb-client-go/v2/log"
)

// Influx is an InfluxDB v2 publisher
type Influx struct {
	client      influxdb2.Client
	org         string
	database    string
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
) *Influx {
	// InfluxDB v1 compatibility
	if token == "" && user != "" {
		token = fmt.Sprintf("%s:%s", user, password)
	}

	options := influxdb2.DefaultOptions().
		SetPrecision(time.Second).
		SetMaxRetries(math.MaxInt).      // retry indefinitely while DB is unavailable
		SetMaxRetryTime(math.MaxUint32). // no overall retry timeout (ms)
		SetMaxRetryInterval(60_000).     // cap individual backoff at 60 s
		SetRetryBufferLimit(100_000)     // buffer up to 100 000 unsent points

	client := influxdb2.NewClientWithOptions(url, token, options)

	// suppress default influx client logger; errors surface via writer.Errors()
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
	writer := m.client.WriteAPI(m.org, m.database)

	// log async write errors
	go func() {
		for err := range writer.Errors() {
			log.Printf("influx: write error: %v", err)
		}
	}()

	for snip := range in {
		p := influxdb2.NewPoint(
			m.measurement,
			map[string]string{
				"device": snip.Device,
				"type":   snip.Measurement.String(),
			},
			map[string]interface{}{"value": snip.Value},
			snip.Timestamp,
		)
		writer.WritePoint(p)
	}

	m.client.Close()
}
