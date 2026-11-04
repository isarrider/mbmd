package rs485

import (
	"errors"
	"fmt"
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

// testBlockProducer additionally allows block reads
type testBlockProducer struct {
	testProducer
	maxLen, maxGap uint16
}

func (p *testBlockProducer) MaxBlockLen() uint16 { return p.maxLen }
func (p *testBlockProducer) MaxBlockGap() uint16 { return p.maxGap }

// testClient answers every register with its own address, unless an error is
// configured for one of the requested registers
type testClient struct {
	modbus.Client
	errors   map[uint16]error
	reads    map[uint16]int // requests by start address
	requests int
}

func newTestClient(errs map[uint16]error) *testClient {
	if errs == nil {
		errs = make(map[uint16]error)
	}
	return &testClient{errors: errs, reads: make(map[uint16]int)}
}

func (c *testClient) ReadHoldingRegisters(address, quantity uint16) ([]byte, error) {
	c.requests++
	c.reads[address]++

	b := make([]byte, 0, 2*quantity)
	for reg := address; reg < address+quantity; reg++ {
		if err, ok := c.errors[reg]; ok {
			return nil, err
		}
		b = append(b, byte(reg>>8), byte(reg))
	}
	return b, nil
}

func newTestDevice(ops ...Operation) *RS485 {
	return newDevice("test", &testProducer{ops: ops})
}

func newTestBlockDevice(maxLen, maxGap uint16, ops ...Operation) *RS485 {
	return newDevice("test", &testBlockProducer{testProducer: testProducer{ops: ops}, maxLen: maxLen, maxGap: maxGap})
}

var rejected = &modbus.Error{FunctionCode: 0x83, ExceptionCode: modbus.ExceptionCodeIllegalDataAddress}

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
	client := newTestClient(map[uint16]error{2: rejected})

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
	client := newTestClient(map[uint16]error{1: rejected})

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
		client := newTestClient(map[uint16]error{2: err})

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
	client := newTestClient(map[uint16]error{1: rejected})

	for i := 0; i < maxOpFailures; i++ {
		_, err := d.Query(client)
		assert.NoError(t, err)
	}

	_, err := d.Query(client)
	assert.Error(t, err)
	assert.Equal(t, maxOpFailures, client.reads[1])
}

func TestBuildPlan(t *testing.T) {
	op32 := func(iec Measurement, opcode uint16) Operation {
		op := newTestOp(iec, opcode)
		op.ReadLen = 2
		return op
	}
	input := newTestOp(Frequency, 3)
	input.FuncCode = ReadInputReg

	d := newTestBlockDevice(10, 2,
		op32(VoltageL1, 1), // 1-2
		newTestOp(VoltageL2, 3),
		op32(VoltageL3, 6),       // gap of 2 after 3, block 1-7
		newTestOp(Voltage, 10),   // gap of 2 after 7, but block would be 10 registers: 1-10
		newTestOp(CurrentL1, 11), // block would exceed 10 registers
		newTestOp(CurrentL2, 15), // gap of 3, new block
		input,                    // other function code, own block
	)

	var blocks []string
	for _, b := range d.buildPlan() {
		blocks = append(blocks, b.String())
	}
	assert.Equal(t, []string{
		"holding registers 0x0001-0x000A",
		"holding registers 0x000B-0x000B",
		"holding registers 0x000F-0x000F",
		"input registers 0x0003-0x0003",
	}, blocks)

	// without BlockProducer every operation is read on its own
	assert.Len(t, newTestDevice(newTestOp(VoltageL1, 1), newTestOp(VoltageL2, 2)).buildPlan(), 2)
}

func TestQueryBlocks(t *testing.T) {
	op32 := newTestOp(Power, 4)
	op32.ReadLen = 2
	op32.Transform = RTUUint32ToFloat64

	d := newTestBlockDevice(100, 0,
		newTestOp(VoltageL1, 1),
		newTestOp(VoltageL2, 2),
		newTestOp(VoltageL3, 3),
		op32,
		newTestOp(CurrentL1, 20),
	)
	client := newTestClient(nil)

	res, err := d.Query(client)
	assert.NoError(t, err)
	assert.Equal(t, 2, client.requests)
	assert.Equal(t, map[Measurement]float64{
		VoltageL1: 1, VoltageL2: 2, VoltageL3: 3,
		Power:     4<<16 | 5,
		CurrentL1: 20,
	}, values(res))
}

func TestQuerySplitsRejectedBlock(t *testing.T) {
	// register 2 lies in a gap and the device does not support it
	d := newTestBlockDevice(100, 5,
		newTestOp(VoltageL1, 1),
		newTestOp(VoltageL2, 3),
		newTestOp(VoltageL3, 4),
	)
	client := newTestClient(map[uint16]error{2: rejected})

	// the block read is rejected, so this run returns nothing but does not fail
	res, err := d.Query(client)
	assert.NoError(t, err)
	assert.Empty(t, res)

	// from now on the operations are read on their own
	for i := 0; i < 3; i++ {
		client.requests = 0
		res, err = d.Query(client)
		assert.NoError(t, err)
		assert.Equal(t, 3, client.requests)
		assert.Equal(t, map[Measurement]float64{VoltageL1: 1, VoltageL2: 3, VoltageL3: 4}, values(res))
	}
}

func TestQueryDisablesUnsupportedOpInBlock(t *testing.T) {
	d := newTestBlockDevice(100, 0,
		newTestOp(VoltageL1, 1),
		newTestOp(VoltageL2, 2),
		newTestOp(VoltageL3, 3),
	)
	client := newTestClient(map[uint16]error{2: rejected})

	// first run splits the block, then VoltageL2 is rejected maxOpFailures times
	for i := 0; i < 1+maxOpFailures; i++ {
		_, err := d.Query(client)
		assert.NoError(t, err)
	}
	assert.True(t, d.disabled[keyOf(newTestOp(VoltageL2, 2))])

	// the remaining operations are read on their own, without VoltageL2
	client.requests = 0
	res, err := d.Query(client)
	assert.NoError(t, err)
	assert.Equal(t, 2, client.requests)
	assert.Equal(t, map[Measurement]float64{VoltageL1: 1, VoltageL3: 3}, values(res))
}

func TestQueryRetriesBlockAfterTransientError(t *testing.T) {
	d := newTestBlockDevice(100, 0,
		newTestOp(VoltageL1, 1),
		newTestOp(VoltageL2, 2),
		newTestOp(CurrentL1, 10),
	)
	client := newTestClient(map[uint16]error{10: errors.New("serial: timeout")})

	res, err := d.Query(client)
	assert.Error(t, err)
	assert.Len(t, res, 2)

	// the next run starts with the block that failed
	delete(client.errors, 10)
	client.reads = make(map[uint16]int)
	res, err = d.Query(client)
	assert.NoError(t, err)
	assert.Len(t, res, 3)
	assert.Equal(t, 1, client.reads[10])
	assert.Empty(t, d.single)
}

// identityClient serves fixed responses by start address
type identityClient struct {
	testClient
	data map[uint16][]byte
}

func (c *identityClient) ReadHoldingRegisters(address, quantity uint16) ([]byte, error) {
	if b, ok := c.data[address]; ok {
		c.reads[address]++
		if err, ok := c.errors[address]; ok {
			return nil, err
		}
		return b, nil
	}
	return c.testClient.ReadHoldingRegisters(address, quantity)
}

type testIdentityProducer struct {
	testProducer
}

func (p *testIdentityProducer) Identity() []IdentityOperation {
	return []IdentityOperation{
		{FuncCode: ReadHoldingReg, OpCode: 0x1000, ReadLen: 1, Apply: func(b []byte, desc *DeviceDescriptor) {
			desc.Serial = fmt.Sprintf("%X", b)
		}},
		{FuncCode: ReadHoldingReg, OpCode: 0x2000, ReadLen: 1, Apply: func(b []byte, desc *DeviceDescriptor) {
			desc.Version = fmt.Sprintf("%X", b)
		}},
	}
}

func TestQueryReadsIdentityOnce(t *testing.T) {
	d := newDevice("test", &testIdentityProducer{testProducer{ops: []Operation{newTestOp(VoltageL1, 1)}}})
	client := &identityClient{
		testClient: *newTestClient(map[uint16]error{0x2000: errors.New("serial: timeout")}),
		data:       map[uint16][]byte{0x1000: {0x12, 0x34}, 0x2000: {0x03, 0x01}},
	}

	// the identity is read before the first query, a timeout does not fail the query
	res, err := d.Query(client)
	assert.NoError(t, err)
	assert.Len(t, res, 1)
	assert.Equal(t, "1234", d.Descriptor().Serial)
	assert.Equal(t, "", d.Descriptor().Version)
	assert.Equal(t, "test", d.Descriptor().Model)

	// only the failed part is read again
	delete(client.errors, 0x2000)
	_, err = d.Query(client)
	assert.NoError(t, err)
	assert.Equal(t, "0301", d.Descriptor().Version)

	_, err = d.Query(client)
	assert.NoError(t, err)
	assert.Equal(t, 1, client.reads[0x1000])
	assert.Equal(t, 2, client.reads[0x2000])
}

func TestQueryIdentityRejected(t *testing.T) {
	d := newDevice("test", &testIdentityProducer{testProducer{ops: []Operation{newTestOp(VoltageL1, 1)}}})
	client := &identityClient{
		testClient: *newTestClient(map[uint16]error{0x1000: rejected, 0x2000: rejected}),
		data:       map[uint16][]byte{0x1000: {0x12, 0x34}, 0x2000: {0x03, 0x01}},
	}

	// a rejected identity is not read again and does not affect the query
	for i := 0; i < 3; i++ {
		res, err := d.Query(client)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	}
	assert.Equal(t, 1, client.reads[0x1000])
	assert.Equal(t, "", d.Descriptor().Serial)
}

func TestQueryIdentityGivesUpAfterTimeouts(t *testing.T) {
	d := newDevice("test", &testIdentityProducer{testProducer{ops: []Operation{newTestOp(VoltageL1, 1)}}})
	client := &identityClient{
		testClient: *newTestClient(map[uint16]error{0x2000: errors.New("serial: timeout")}),
		data:       map[uint16][]byte{0x1000: {0x12, 0x34}, 0x2000: {0x03, 0x01}},
	}

	for i := 0; i < 2*maxOpFailures; i++ {
		res, err := d.Query(client)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	}

	// the version timed out on every try and is not read after maxOpFailures tries
	assert.Equal(t, maxOpFailures, client.reads[0x2000])
	assert.Equal(t, "1234", d.Descriptor().Serial)
	assert.Empty(t, d.failures)
}
