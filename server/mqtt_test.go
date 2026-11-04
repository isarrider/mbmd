package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/volkszaehler/mbmd/meters"
)

func TestTopicFromMeasurement(t *testing.T) {
	for m, topic := range map[meters.Measurement]string{
		meters.Power:                  "Power",
		meters.PowerL1:                "Power/L1",
		meters.SumT1:                  "Sum/T1",
		meters.ImportPowerMaxDemand:   "ImportPowerMaxDemand",
		meters.ImportPowerMaxDemandL1: "ImportPowerMaxDemand/L1",
		meters.ReactiveQ1:             "ReactiveQ1",
		meters.ReactiveQ1L2:           "ReactiveQ1/L2",
		meters.ResettableSum:          "ResettableSum",
		meters.ResettableSumL3:        "ResettableSum/L3",
	} {
		assert.Equal(t, topic, topicFromMeasurement(m), m.String())
	}
}
