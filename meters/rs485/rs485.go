package rs485

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/grid-x/modbus"
	"github.com/volkszaehler/mbmd/meters"
)

const (
	ReadHoldingReg = 3
	ReadInputReg   = 4
)

// maxOpFailures is the number of consecutive reads a device may reject for an
// operation before the operation is no longer read
const maxOpFailures = 3

// RS485 implements meters.Device
type RS485 struct {
	typ      string
	id       string
	producer Producer
	plan     []block // bus reads of one query run, nil if it needs to be (re)built
	next     int     // index of the next block in plan
	failures map[opKey]int
	disabled map[opKey]bool      // operations the device does not support
	single   map[opKey]bool      // operations from rejected blocks, read on their own
	identity []IdentityOperation // identity operations not read yet, nil before the first query
	mu       sync.Mutex          // guards details, which Descriptor reads from other goroutines
	details  meters.DeviceDescriptor
}

// opKey identifies an operation independent of its transform
type opKey struct {
	funcCode uint8
	opCode   uint16
	iec      meters.Measurement
}

func keyOf(op Operation) opKey {
	return opKey{funcCode: op.FuncCode, opCode: op.OpCode, iec: op.IEC61850}
}

// block is a single bus read covering one or more operations
type block struct {
	funcCode uint8
	start    uint16
	length   uint16
	ops      []Operation
}

func (b block) String() string {
	return fmt.Sprintf("%s registers 0x%04X-0x%04X", registerType(b.funcCode), b.start, uint32(b.start)+uint32(b.length)-1)
}

// NewDevice creates a device who's type must exist in the producer registry
func NewDevice(typ string) (*RS485, error) {
	for t, factory := range Producers {
		if strings.EqualFold(t, typ) {
			return newDevice(typ, factory()), nil
		}
	}

	return nil, fmt.Errorf("unknown meter type: %s", typ)
}

func newDevice(typ string, producer Producer) *RS485 {
	return &RS485{
		typ:      typ,
		producer: producer,
		failures: make(map[opKey]int),
		disabled: make(map[opKey]bool),
		single:   make(map[opKey]bool),
	}
}

// Initialize prepares the device for usage. Any setup or initialization should be done here.
func (d *RS485) Initialize(client modbus.Client) error {
	return nil
}

// SetID sets the device id used in log messages
func (d *RS485) SetID(id string) {
	d.id = id
}

func (d *RS485) name() string {
	if d.id != "" {
		return d.id
	}
	return d.typ
}

// Producer returns the underlying producer. The producer can be used to understand which operations the device supports.
func (d *RS485) Producer() Producer {
	return d.producer
}

// Descriptor returns the device descriptor. Since this method does not have bus access the descriptor should be
// prepared during initialization.
func (d *RS485) Descriptor() meters.DeviceDescriptor {
	d.mu.Lock()
	defer d.mu.Unlock()

	desc := meters.DeviceDescriptor{
		Type:         d.typ,
		Manufacturer: d.typ,
		Model:        d.producer.Description(),
		Version:      d.details.Version,
		Serial:       d.details.Serial,
	}

	if d.details.Model != "" {
		desc.Model = d.details.Model
	}

	return desc
}

// identify reads the device details of an IdentityProducer. Operations that
// succeed or that the device rejects are not read again, operations that fail
// for other reasons are retried on the next query, up to maxOpFailures times.
func (d *RS485) identify(client modbus.Client) {
	if d.identity == nil {
		ip, ok := d.producer.(IdentityProducer)
		if !ok {
			d.identity = []IdentityOperation{}
			return
		}
		d.identity = ip.Identity()
	}

	if len(d.identity) == 0 {
		return
	}

	// not nil, nil means the identity operations have not been fetched yet
	pending := []IdentityOperation{}
	for _, op := range d.identity {
		// measurement zero is unused, so the key does not collide with an operation
		key := opKey{funcCode: op.FuncCode, opCode: op.OpCode}
		where := fmt.Sprintf("%s register 0x%04X", registerType(op.FuncCode), op.OpCode)

		b, err := readRegisters(client, op.FuncCode, op.OpCode, op.ReadLen)
		if err != nil {
			if isRejected(err) {
				log.Printf("device %s: device details at %s are not supported: %v", d.name(), where, err)
				continue
			}

			d.failures[key]++
			if d.failures[key] < maxOpFailures {
				log.Printf("device %s: reading device details at %s failed (%d/%d): %v", d.name(), where, d.failures[key], maxOpFailures, err)
				pending = append(pending, op)
			} else {
				delete(d.failures, key)
				log.Printf("device %s: reading device details at %s failed %d times, giving up: %v", d.name(), where, maxOpFailures, err)
			}
			continue
		}
		delete(d.failures, key)

		d.mu.Lock()
		op.Apply(b, &d.details)
		d.mu.Unlock()
	}
	d.identity = pending

	if len(pending) == 0 && (d.details.Serial != "" || d.details.Version != "") {
		desc := d.Descriptor()
		details := []string{desc.Model}
		if desc.Serial != "" {
			details = append(details, "serial "+desc.Serial)
		}
		if desc.Version != "" {
			details = append(details, "version "+desc.Version)
		}
		log.Printf("device %s: %s", d.name(), strings.Join(details, ", "))
	}
}

// ReadDetails reads the device details of an IdentityProducer right away and
// replaces the ones read before. Scan uses it, because there one device object
// probes every bus address.
func (d *RS485) ReadDetails(client modbus.Client) {
	d.mu.Lock()
	d.details = meters.DeviceDescriptor{}
	d.mu.Unlock()

	// identity failures use measurement zero, see identify
	for key := range d.failures {
		if key.iec == 0 {
			delete(d.failures, key)
		}
	}

	d.identity = nil
	d.identify(client)
}

// readRegisters reads length registers and checks the response size
func readRegisters(client modbus.Client, funcCode uint8, start, length uint16) (b []byte, err error) {
	switch funcCode {
	case ReadHoldingReg:
		b, err = client.ReadHoldingRegisters(start, length)
	case ReadInputReg:
		b, err = client.ReadInputRegisters(start, length)
	default:
		return nil, fmt.Errorf("unknown function code %d", funcCode)
	}

	if err != nil {
		return nil, fmt.Errorf("read failed: %w", err)
	}

	if len(b) < 2*int(length) {
		return nil, fmt.Errorf("read failed: expected %d bytes, got %d", 2*int(length), len(b))
	}

	return b, nil
}

// Probe is called by the handler after preparing the bus by setting the device id
func (d *RS485) Probe(client modbus.Client) (res meters.MeasurementResult, err error) {
	op := d.producer.Probe()

	// check for empty op in case Probe isn't supported
	if op.FuncCode == 0 {
		return res, fmt.Errorf("meter type %s doesn't support Probe", d.producer.Description())
	}

	res, err = d.QueryOp(client, op)
	if err != nil {
		return res, err
	}

	return res, nil
}

// QueryOp executes a single query operation on the bus
func (d *RS485) QueryOp(client modbus.Client, op Operation) (res meters.MeasurementResult, err error) {
	results, err := d.queryBlock(client, block{
		funcCode: op.FuncCode,
		start:    op.OpCode,
		length:   op.ReadLen,
		ops:      []Operation{op},
	})
	if err != nil {
		return res, err
	}

	return results[0], nil
}

// queryBlock reads a block of registers in a single request and converts it to
// the block's measurements
func (d *RS485) queryBlock(client modbus.Client, b block) (res []meters.MeasurementResult, err error) {
	for _, op := range b.ops {
		if op.ReadLen == 0 {
			return res, fmt.Errorf("invalid meter operation %v", op)
		}

		if op.Transform == nil {
			return res, fmt.Errorf("transformation not defined: %v", op)
		}
	}

	bytes, err := readRegisters(client, b.funcCode, b.start, b.length)
	if err != nil {
		return res, err
	}

	now := time.Now()
	for _, op := range b.ops {
		offset := 2 * int(op.OpCode-b.start)
		res = append(res, meters.MeasurementResult{
			Measurement: op.IEC61850,
			Value:       op.Transform(bytes[offset : offset+2*int(op.ReadLen)]),
			Timestamp:   now,
		})
	}

	return res, nil
}

// buildPlan groups the producer's operations into bus reads. Producers that
// implement BlockProducer get adjacent operations combined into one read,
// all others get one read per operation.
func (d *RS485) buildPlan() []block {
	var maxLen, maxGap uint16
	if bp, ok := d.producer.(BlockProducer); ok {
		maxLen, maxGap = bp.MaxBlockLen(), bp.MaxBlockGap()
	}

	var plan []block
	var funcCodes []uint8
	grouped := make(map[uint8][]Operation)

	for _, op := range d.producer.Produce() {
		key := keyOf(op)
		if d.disabled[key] {
			continue
		}

		if maxLen == 0 || d.single[key] || op.ReadLen >= maxLen {
			plan = append(plan, block{funcCode: op.FuncCode, start: op.OpCode, length: op.ReadLen, ops: []Operation{op}})
			continue
		}

		if _, ok := grouped[op.FuncCode]; !ok {
			funcCodes = append(funcCodes, op.FuncCode)
		}
		grouped[op.FuncCode] = append(grouped[op.FuncCode], op)
	}

	for _, funcCode := range funcCodes {
		ops := grouped[funcCode]
		sort.SliceStable(ops, func(i, j int) bool { return ops[i].OpCode < ops[j].OpCode })

		var cur *block
		for _, op := range ops {
			start, end := uint32(op.OpCode), uint32(op.OpCode)+uint32(op.ReadLen)

			if cur != nil {
				curStart, curEnd := uint32(cur.start), uint32(cur.start)+uint32(cur.length)

				if start <= curEnd+uint32(maxGap) && max(end, curEnd)-curStart <= uint32(maxLen) {
					cur.length = uint16(max(end, curEnd) - curStart)
					cur.ops = append(cur.ops, op)
					continue
				}

				plan = append(plan, *cur)
			}

			cur = &block{funcCode: funcCode, start: op.OpCode, length: op.ReadLen, ops: []Operation{op}}
		}

		if cur != nil {
			plan = append(plan, *cur)
		}
	}

	return plan
}

// Query is called by the handler after preparing the bus by setting the device id and waiting for rate limit
func (d *RS485) Query(client modbus.Client) (res []meters.MeasurementResult, err error) {
	res = make([]meters.MeasurementResult, 0)

	d.identify(client)

	if d.plan == nil {
		d.plan = d.buildPlan()
		d.next = 0
	}

	if len(d.plan) == 0 {
		return res, errors.New("device rejected all registers")
	}

	// Query loop will try to read all blocks in a single run. It will always
	// start with the block that was not completed in the previous run. If an
	// error is encountered, the partial results are returned. The loop is
	// terminated after every block has been read once. In case of a flakey
	// connection this guarantees that all registers are read at an equal rate.
	//
	// If the device rejects a block, its operations are read on their own from
	// the next run on. If the device rejects an operation read on its own, e.g.
	// because it does not support the register, the operation is skipped and
	// the remaining operations are read. After maxOpFailures consecutive
	// rejections the operation is no longer read.
	plan := d.plan
	for range plan {
		if d.next >= len(plan) {
			d.next = 0
		}
		b := plan[d.next]

		m, err := d.queryBlock(client, b)
		if err != nil {
			if !isRejected(err) {
				return res, err
			}

			d.reject(b, err)
			d.next++
			continue
		}

		for _, op := range b.ops {
			delete(d.failures, keyOf(op))
		}

		d.next++
		res = append(res, m...)
	}

	return res, nil
}

// reject handles a block the device rejected. A block with several operations
// is split so that its operations are read on their own. A single operation is
// disabled after maxOpFailures consecutive rejections.
func (d *RS485) reject(b block, err error) {
	if len(b.ops) > 1 {
		for _, op := range b.ops {
			d.single[keyOf(op)] = true
		}
		d.plan = nil
		log.Printf("device %s: reading %s was rejected, reading its %d values one at a time: %v", d.name(), b, len(b.ops), err)
		return
	}

	op := b.ops[0]
	key := keyOf(op)
	d.failures[key]++

	if d.failures[key] < maxOpFailures {
		log.Printf("device %s: reading %s failed (%d/%d): %v", d.name(), describeOp(op), d.failures[key], maxOpFailures, err)
		return
	}

	delete(d.failures, key)
	d.disabled[key] = true
	d.plan = nil
	log.Printf("device %s: %s is not supported, no longer reading it until restart: %v", d.name(), describeOp(op), err)
}

func registerType(funcCode uint8) string {
	if funcCode == ReadInputReg {
		return "input"
	}
	return "holding"
}

func describeOp(op Operation) string {
	return fmt.Sprintf("%s (%s register 0x%04X)", op.IEC61850, registerType(op.FuncCode), op.OpCode)
}

// isRejected returns true if the device answered with an exception saying it
// cannot serve the request. Timeouts and transient exceptions like a busy
// device are not rejections.
func isRejected(err error) bool {
	var mbErr *modbus.Error
	if !errors.As(err, &mbErr) {
		return false
	}

	switch mbErr.ExceptionCode {
	case modbus.ExceptionCodeIllegalFunction,
		modbus.ExceptionCodeIllegalDataAddress,
		modbus.ExceptionCodeIllegalDataValue:
		return true
	default:
		return false
	}
}
