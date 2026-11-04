package rs485

import (
	"fmt"

	"github.com/volkszaehler/mbmd/meters"
)

// Operation describes a physical bus operation and its result
type Operation struct {
	FuncCode  uint8
	OpCode    uint16
	ReadLen   uint16
	IEC61850  meters.Measurement
	Transform RTUTransform
}

// Producer is the interface that produces query snips which represent
// modbus operations
type Producer interface {
	// Type returns device description, typically static
	Description() string

	// Produce creates a slice of possible device operations
	Produce() []Operation

	// Produce creates a device operation suited to detect the device during
	// scanning, typically a L1 voltage read operation
	Probe() Operation
}

// BlockProducer is implemented by producers whose devices can read several
// consecutive registers in a single request. Operations are then combined into
// blocks. If the device rejects a block, its operations are read on their own.
type BlockProducer interface {
	// MaxBlockLen returns the maximum number of registers read in one request
	MaxBlockLen() uint16

	// MaxBlockGap returns the maximum number of registers between two operations
	// that are read and discarded to combine the operations into one request
	MaxBlockGap() uint16
}

// IdentityProducer is implemented by producers that can read device details
// like the serial number or firmware version. They are read once, on the
// first query of the device.
type IdentityProducer interface {
	Identity() []IdentityOperation
}

// IdentityOperation reads registers and applies them to the device descriptor
type IdentityOperation struct {
	FuncCode uint8
	OpCode   uint16
	ReadLen  uint16
	Apply    func(b []byte, desc *meters.DeviceDescriptor)
}

// Opcodes map measurements to physical registers
type Opcodes map[meters.Measurement]uint16

// Opcode returns physical register for measurement type
func (o *Opcodes) Opcode(iec meters.Measurement) uint16 {
	if opcode, ok := (*o)[iec]; ok {
		return opcode
	}

	panic(fmt.Sprintf("Undefined opcode for measurement %s", iec.String()))
}
