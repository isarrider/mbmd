package server_test

import (
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/volkszaehler/mbmd/meters"
)

// TestUIShowsAllMeasurements makes sure every measurement has a row in the web UI
func TestUIShowsAllMeasurements(t *testing.T) {
	html, err := os.ReadFile("../assets/index.html")
	assert.NoError(t, err)

	shown := make(map[string]bool)
	for _, match := range regexp.MustCompile(`m\.(\w+)`).FindAllStringSubmatch(string(html), -1) {
		shown[match[1]] = true
	}

	for _, m := range meters.MeasurementValues() {
		assert.True(t, shown[m.String()], "%s is not shown in assets/index.html", m)
	}
}
