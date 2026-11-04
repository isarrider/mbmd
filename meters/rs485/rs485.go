package rs485

import (
	"errors"
	"fmt"
	"log"
	"strings"
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
	ops      chan Operation
	inflight Operation
	failures map[opKey]int
	disabled map[opKey]bool
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

// NewDevice creates a device who's type must exist in the producer registry
func NewDevice(typ string) (*RS485, error) {
	for t, factory := range Producers {
		if strings.EqualFold(t, typ) {
			device := &RS485{
				typ:      typ,
				producer: factory(),
				failures: make(map[opKey]int),
				disabled: make(map[opKey]bool),
			}
			return device, nil
		}
	}

	return nil, fmt.Errorf("unknown meter type: %s", typ)
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
	return meters.DeviceDescriptor{
		Type:         d.typ,
		Manufacturer: d.typ,
		Model:        d.producer.Description(),
	}
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
	var bytes []byte

	if op.ReadLen == 0 {
		return res, fmt.Errorf("invalid meter operation %v", op)
	}

	if op.Transform == nil {
		return res, fmt.Errorf("transformation not defined: %v", op)
	}

	switch op.FuncCode {
	case ReadHoldingReg:
		bytes, err = client.ReadHoldingRegisters(op.OpCode, op.ReadLen)
	case ReadInputReg:
		bytes, err = client.ReadInputRegisters(op.OpCode, op.ReadLen)
	default:
		return res, fmt.Errorf("unknown function code %d", op.FuncCode)
	}

	if err != nil {
		return res, fmt.Errorf("read failed: %w", err)
	}

	res = meters.MeasurementResult{
		Measurement: op.IEC61850,
		Value:       op.Transform(bytes),
		Timestamp:   time.Now(),
	}

	return res, nil
}

// Query is called by the handler after preparing the bus by setting the device id and waiting for rate limit
func (d *RS485) Query(client modbus.Client) (res []meters.MeasurementResult, err error) {
	res = make([]meters.MeasurementResult, 0)

	if d.ops == nil {
		d.ops = make(chan Operation)

		// ringbuffer of device operations
		go func(d *RS485) {
			for {
				for _, op := range d.producer.Produce() {
					d.ops <- op
				}
			}
		}(d)
	}

	// Query loop will try to read all operations in a single run. It will
	// always start with the current inflight operation. If an error is encountered,
	// the partial results are returned. The loop is terminated after as many
	// operations have been executed as the producer provides in a single run.
	// In case of a flakey connection this guarantees that all registers are
	// read at an equal rate.
	//
	// If the device rejects an operation, e.g. because it does not support the
	// register, the operation is skipped and the remaining operations are read.
	// After maxOpFailures consecutive rejections the operation is no longer read.
	enabled := len(d.producer.Produce()) - len(d.disabled)
	if enabled <= 0 {
		return res, errors.New("device rejected all registers")
	}

	for n := 0; n < enabled; {
		// get next inflight
		if d.inflight.FuncCode == 0 {
			d.inflight = <-d.ops
		}
		op := d.inflight

		if d.disabled[keyOf(op)] {
			d.inflight.FuncCode = 0
			continue
		}

		m, err := d.QueryOp(client, op)
		if err != nil {
			if !isRejected(err) {
				return res, err
			}

			d.reject(op, err)
			d.inflight.FuncCode = 0
			n++
			continue
		}

		delete(d.failures, keyOf(op))

		// mark inflight operation as completed
		d.inflight.FuncCode = 0
		n++

		res = append(res, m)
	}

	return res, nil
}

// reject counts a rejected operation and disables it after maxOpFailures consecutive rejections
func (d *RS485) reject(op Operation, err error) {
	key := keyOf(op)
	d.failures[key]++

	if d.failures[key] < maxOpFailures {
		log.Printf("device %s: reading %s failed (%d/%d): %v", d.name(), describeOp(op), d.failures[key], maxOpFailures, err)
		return
	}

	delete(d.failures, key)
	d.disabled[key] = true
	log.Printf("device %s: %s is not supported, no longer reading it until restart: %v", d.name(), describeOp(op), err)
}

func describeOp(op Operation) string {
	typ := "holding"
	if op.FuncCode == ReadInputReg {
		typ = "input"
	}
	return fmt.Sprintf("%s (%s register 0x%04X)", op.IEC61850, typ, op.OpCode)
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
