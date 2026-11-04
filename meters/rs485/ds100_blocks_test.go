package rs485

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDS100Blocks(t *testing.T) {
	p := NewDS100Producer()
	plan := newDevice("DS100", p).buildPlan()

	ops := 0
	for _, b := range plan {
		assert.LessOrEqual(t, b.length, p.(BlockProducer).MaxBlockLen(), b.String())
		ops += len(b.ops)

		// all instantaneous values are read in one request
		for _, op := range b.ops {
			if op.OpCode >= 0x0400 && op.OpCode <= 0x0439 {
				assert.Equal(t, uint16(0x0400), b.start, "%s in %s", op.IEC61850, b)
			}
		}
	}

	assert.Equal(t, len(p.Produce()), ops, "every value is read exactly once")
	assert.Less(t, len(plan), len(p.Produce())/4)
}
