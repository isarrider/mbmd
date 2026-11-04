package rs485

import (
	"testing"

	"github.com/stretchr/testify/assert"
	. "github.com/volkszaehler/mbmd/meters"
)

func TestDS100Identity(t *testing.T) {
	var desc DeviceDescriptor
	ops := NewDS100Producer().(IdentityProducer).Identity()

	// example responses from the manual, section 3.3 and 3.4
	ops[0].Apply([]byte{0x20, 0x24, 0x05, 0x27, 0x00, 0x01}, &desc)
	ops[1].Apply([]byte{0x01, 0x2D, 0x01, 0x2D, 0x5B, 0x61}, &desc)

	assert.Equal(t, "202405270001", desc.Serial)
	assert.Equal(t, "V301 (hardware V301)", desc.Version)
	assert.Equal(t, "B+G e-tech DS100-00B", desc.Model)

	// unknown checksum keeps the default model
	desc = DeviceDescriptor{}
	ops[1].Apply([]byte{0x01, 0x2E, 0x01, 0x2D, 0x12, 0x34}, &desc)
	assert.Equal(t, "V302 (hardware V301)", desc.Version)
	assert.Empty(t, desc.Model)
}
