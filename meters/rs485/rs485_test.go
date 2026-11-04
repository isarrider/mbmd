package rs485

import (
	"errors"
	"testing"

	"github.com/grid-x/modbus"
	"github.com/stretchr/testify/assert"
	. "github.com/volkszaehler/mbmd/meters"
)

// testProducer reads one register per measurement
type testProducer struct {
	ops []Operation
}

func (p *testProducer) Description() string  { return "test" }
func (p *testProducer) Produce() []Operation { return p.ops }
func (p *testProducer) Probe() Operation     { return p.ops[0] }
func newTestOp(iec Measurement, opcode uint16) Operation {
	return Operation{
		FuncCode:  ReadHoldingReg,
		OpCode:    opcode,
		ReadLen:   1,
		IEC61850:  iec,
		Transform: RTUUint16ToFloat64,
	}
}

// testClient answers every holding register read with the register address,
// unless an error is configured for the address
type testClient struct {
	modbus.Client
	errors map[uint16]error
	reads  map[uint16]int
}

func (c *testClient) ReadHoldingRegisters(address, quantity uint16) ([]byte, error) {
	c.reads[address]++
	if err, ok := c.errors[address]; ok {
		return nil, err
	}
	return []byte{byte(address >> 8), byte(address)}, nil
}

func newTestDevice(ops ...Operation) *RS485 {
	return &RS485{
		typ:      "test",
		producer: &testProducer{ops: ops},
		failures: make(map[opKey]int),
		disabled: make(map[opKey]bool),
	}
}

func values(res []MeasurementResult) map[Measurement]float64 {
	m := make(map[Measurement]float64)
	for _, r := range res {
		m[r.Measurement] = r.Value
	}
	return m
}

func TestQuerySkipsRejectedRegister(t *testing.T) {
	d := newTestDevice(
		newTestOp(VoltageL1, 1),
		newTestOp(VoltageL2, 2),
		newTestOp(VoltageL3, 3),
	)
	client := &testClient{
		errors: map[uint16]error{
			2: &modbus.Error{FunctionCode: 0x83, ExceptionCode: modbus.ExceptionCodeIllegalDataAddress},
		},
		reads: make(map[uint16]int),
	}

	// a rejected register does not fail the query, the other values are still read
	for i := 1; i <= maxOpFailures; i++ {
		res, err := d.Query(client)
		assert.NoError(t, err)
		assert.Equal(t, map[Measurement]float64{VoltageL1: 1, VoltageL3: 3}, values(res), "run %d", i)
	}

	// after maxOpFailures rejections the register is no longer read
	assert.True(t, d.disabled[keyOf(newTestOp(VoltageL2, 2))])
	for i := 0; i < 3; i++ {
		res, err := d.Query(client)
		assert.NoError(t, err)
		assert.Equal(t, map[Measurement]float64{VoltageL1: 1, VoltageL3: 3}, values(res))
	}
	assert.Equal(t, maxOpFailures, client.reads[2])
	assert.Equal(t, maxOpFailures+3, client.reads[1])
}

func TestQueryResetsFailuresOnSuccess(t *testing.T) {
	op := newTestOp(VoltageL1, 1)
	d := newTestDevice(op, newTestOp(VoltageL2, 2))
	rejected := &modbus.Error{FunctionCode: 0x83, ExceptionCode: modbus.ExceptionCodeIllegalDataAddress}
	client := &testClient{errors: map[uint16]error{1: rejected}, reads: make(map[uint16]int)}

	// fewer than maxOpFailures rejections, then a successful read
	for i := 1; i < maxOpFailures; i++ {
		_, err := d.Query(client)
		assert.NoError(t, err)
	}
	delete(client.errors, 1)
	_, err := d.Query(client)
	assert.NoError(t, err)

	// a later rejection starts counting from zero again
	client.errors[1] = rejected
	for i := 1; i < maxOpFailures; i++ {
		_, err := d.Query(client)
		assert.NoError(t, err)
	}
	assert.False(t, d.disabled[keyOf(op)])
}

func TestQueryFailsOnTransientError(t *testing.T) {
	for name, err := range map[string]error{
		"timeout": errors.New("serial: timeout"),
		"busy":    &modbus.Error{FunctionCode: 0x83, ExceptionCode: modbus.ExceptionCodeServerDeviceBusy},
	} {
		d := newTestDevice(newTestOp(VoltageL1, 1), newTestOp(VoltageL2, 2))
		client := &testClient{errors: map[uint16]error{2: err}, reads: make(map[uint16]int)}

		// transient errors fail the query like before, so the handler retries
		// and eventually marks the device offline
		for i := 0; i < 2*maxOpFailures; i++ {
			_, qerr := d.Query(client)
			assert.Error(t, qerr, name)
		}
		assert.Empty(t, d.disabled, name)
	}
}

func TestQueryAllRegistersRejected(t *testing.T) {
	d := newTestDevice(newTestOp(VoltageL1, 1))
	rejected := &modbus.Error{FunctionCode: 0x83, ExceptionCode: modbus.ExceptionCodeIllegalDataAddress}
	client := &testClient{errors: map[uint16]error{1: rejected}, reads: make(map[uint16]int)}

	for i := 0; i < maxOpFailures; i++ {
		_, err := d.Query(client)
		assert.NoError(t, err)
	}

	_, err := d.Query(client)
	assert.Error(t, err)
	assert.Equal(t, maxOpFailures, client.reads[1])
}
