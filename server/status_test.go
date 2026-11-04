package server

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/volkszaehler/mbmd/meters"
)

type testDeviceInfo map[string]meters.DeviceDescriptor

func (d testDeviceInfo) DeviceDescriptorByID(id string) meters.DeviceDescriptor {
	return d[id]
}

func TestStatusShowsDeviceDetails(t *testing.T) {
	info := testDeviceInfo{
		"DS1001.1": {Manufacturer: "DS100", Model: "B+G e-tech DS100-00B", Serial: "202405270001", Version: "V301"},
		"SDM1.2":   {Manufacturer: "SDM", Model: "Eastron SDM630"},
	}
	control := make(chan ControlSnip)
	s := NewStatus(info, control)

	control <- ControlSnip{Device: "DS1001.1", Status: RuntimeInfo{Online: true}}
	control <- ControlSnip{Device: "SDM1.2", Status: RuntimeInfo{Online: true}}

	var res struct{ Meters []map[string]any }
	assert.Eventually(t, func() bool {
		b, err := json.Marshal(s)
		assert.NoError(t, err)
		assert.NoError(t, json.Unmarshal(b, &res))
		return len(res.Meters) == 2
	}, time.Second, 10*time.Millisecond)

	byDevice := make(map[string]map[string]any)
	for _, m := range res.Meters {
		byDevice[m["Device"].(string)] = m
	}

	ds100 := byDevice["DS1001.1"]
	assert.Equal(t, "DS100", ds100["Type"])
	assert.Equal(t, "B+G e-tech DS100-00B", ds100["Model"])
	assert.Equal(t, "202405270001", ds100["Serial"])
	assert.Equal(t, "V301", ds100["Version"])

	// empty details are left out
	sdm := byDevice["SDM1.2"]
	assert.Equal(t, "Eastron SDM630", sdm["Model"])
	assert.NotContains(t, sdm, "Serial")
	assert.NotContains(t, sdm, "Version")
}
