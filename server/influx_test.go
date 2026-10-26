package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/volkszaehler/mbmd/meters"
)

func TestInfluxOptions(t *testing.T) {
	opts := influxOptions(time.Hour)

	if batches := opts.WriteOptions().RetryBufferLimit() / opts.WriteOptions().BatchSize(); batches < 3600 {
		t.Errorf("retry buffer holds %d batches, want >= 3600", batches)
	}
	if got := opts.WriteOptions().MaxRetryTime(); got != uint(time.Hour.Milliseconds()) {
		t.Errorf("max retry time %d, want %d", got, time.Hour.Milliseconds())
	}
	if opts.WriteOptions().MaxRetries() == 0 {
		t.Error("retries disabled")
	}

	if got := influxOptions(0).WriteOptions().MaxRetries(); got != 0 {
		t.Errorf("buffer 0: max retries %d, want 0", got)
	}
}

// fakeInflux counts written lines and fails all writes while down is set
type fakeInflux struct {
	down  atomic.Bool
	mu    sync.Mutex
	lines int
}

func (f *fakeInflux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if f.down.Load() {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	body, _ := io.ReadAll(r.Body)
	f.mu.Lock()
	f.lines += len(strings.Split(strings.TrimSpace(string(body)), "\n"))
	f.mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}

func (f *fakeInflux) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.lines
}

// TestInfluxOutage verifies that points written while the database is down
// are delivered once it is back.
func TestInfluxOutage(t *testing.T) {
	if testing.Short() {
		t.Skip("slow")
	}

	fake := &fakeInflux{}
	srv := httptest.NewServer(fake)
	defer srv.Close()

	influx := NewInfluxClient(srv.URL, "db", "mbmd", "org", "", "", "", time.Hour)

	in := make(chan QuerySnip)
	done := make(chan struct{})
	go func() {
		influx.Run(in)
		close(done)
	}()

	// outage of 20s, longer than the ~10s the client's default buffer covers
	fake.down.Store(true)

	const points = 100
	for i := range points {
		in <- QuerySnip{
			Device:            "test",
			MeasurementResult: meters.MeasurementResult{Measurement: meters.Power, Value: float64(i)},
		}
		time.Sleep(200 * time.Millisecond)
	}

	fake.down.Store(false)

	// keep data flowing, retries are triggered by new writes
	deadline := time.Now().Add(time.Minute)
	sent := points
	for fake.count() < points && time.Now().Before(deadline) {
		in <- QuerySnip{
			Device:            "test",
			MeasurementResult: meters.MeasurementResult{Measurement: meters.Power, Value: -1},
		}
		sent++
		time.Sleep(time.Second)
	}

	close(in)
	<-done

	if got := fake.count(); got != sent {
		t.Errorf("received %d of %d points", got, sent)
	}
}
