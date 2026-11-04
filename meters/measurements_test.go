package meters

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMeasurementsHaveDescriptionAndUnit(t *testing.T) {
	for _, m := range MeasurementValues() {
		details, ok := iec[m]
		if !assert.True(t, ok, "%s has no description", m) {
			continue
		}
		assert.NotEmpty(t, details[0], "%s has an empty description", m)

		// power factor is the only measurement without a unit
		if m != Cosphi && m != CosphiL1 && m != CosphiL2 && m != CosphiL3 {
			assert.Len(t, details, 2, "%s has no unit", m)
		}
	}
}

func TestMeasurementNamesRoundTrip(t *testing.T) {
	for _, m := range MeasurementValues() {
		parsed, err := MeasurementString(m.String())
		assert.NoError(t, err)
		assert.Equal(t, m, parsed)
	}
}
