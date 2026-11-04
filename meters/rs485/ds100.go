package rs485

import . "github.com/volkszaehler/mbmd/meters"

func init() {
	Register("DS100", NewDS100Producer)
}

type DS100Producer struct {
	Opcodes
}

func NewDS100Producer() Producer {
	/**
	 * Opcodes as defined by B+G e-tech DS100.
	 * See https://data.xn--stromzhler-v5a.eu/manuals/bg_ds100serie_de.pdf
	 */
	ops := Opcodes{
		VoltageL1:       0x0400,
		VoltageL2:       0x0402,
		VoltageL3:       0x0404,
		VoltageL1_L2:    0x0406,
		VoltageL2_L3:    0x0408,
		VoltageL3_L1:    0x040A,
		Voltage:         0x040C, // average L-N
		VoltageL_L_avg:  0x040E,
		CurrentL1:       0x0410,
		CurrentL2:       0x0412,
		CurrentL3:       0x0414,
		CurrentN:        0x0416,
		Current:         0x0418, // three phase vector current
		PowerL1:         0x041A,
		PowerL2:         0x041C,
		PowerL3:         0x041E,
		Power:           0x0420,
		ApparentPowerL1: 0x0422,
		ApparentPowerL2: 0x0424,
		ApparentPowerL3: 0x0426,
		ApparentPower:   0x0428,
		ReactivePowerL1: 0x042A,
		ReactivePowerL2: 0x042C,
		ReactivePowerL3: 0x042E,
		ReactivePower:   0x0430,
		FrequencyL1:     0x0432,
		FrequencyL2:     0x0433,
		FrequencyL3:     0x0434,
		Frequency:       0x0435,
		CosphiL1:        0x0436,
		CosphiL2:        0x0437,
		CosphiL3:        0x0438,
		Cosphi:          0x0439,
		ImportL1:        0x050A, //A phase forward active energy
		ImportL2:        0x056E, //B phase forward active energy
		ImportL3:        0x05D2, //C phase forward active energy
		Import:          0x010E, //  total forward active energy
		ExportL1:        0x0514, //A phase reverse active energy
		ExportL2:        0x0578, //B phase reverse active energy
		ExportL3:        0x05DC, //C phase reverse active energy
		Export:          0x0118, //  total reverse active energy
		SumL1:           0x0500, //A phase total active energy
		SumL2:           0x0564, //B phase total active energy
		SumL3:           0x05C8, //C phase total active energy
		Sum:             0x0122, //  total total active energy

		ReactiveImportL1: 0x0528, // A phase forward reactive energy
		ReactiveImportL2: 0x058C, // B phase forward reactive energy
		ReactiveImportL3: 0x05F0, // C phase forward reactive energy
		ReactiveImport:   0x012C, // total forward reactive energy
		ReactiveExportL1: 0x0532, // A phase reverse reactive energy
		ReactiveExportL2: 0x0596, // B phase reverse reactive energy
		ReactiveExportL3: 0x05FA, // C phase reverse reactive energy
		ReactiveExport:   0x0136, // total reverse reactive energy
		ReactiveSumL1:    0x051E, // A phase total reactive energy
		ReactiveSumL2:    0x0582, // B phase total reactive energy
		ReactiveSumL3:    0x05E6, // C phase total reactive energy
		ReactiveSum:      0x0140, // total reactive energy

		// demand over the demand period (register 0x1011, default 15 minutes), 0.1 W / 0.1 var
		ImportPowerDemandL1:         0x043A, // A phase forward active demand
		ImportPowerDemandL2:         0x043C,
		ImportPowerDemandL3:         0x043E,
		ImportPowerDemand:           0x0440,
		ExportPowerDemandL1:         0x0442, // A phase reverse active demand
		ExportPowerDemandL2:         0x0444,
		ExportPowerDemandL3:         0x0446,
		ExportPowerDemand:           0x0448,
		PowerDemandL1:               0x044A, // A phase total active demand
		PowerDemandL2:               0x044C,
		PowerDemandL3:               0x044E,
		PowerDemand:                 0x0450,
		ReactiveImportPowerDemandL1: 0x0452, // A phase forward reactive demand
		ReactiveImportPowerDemandL2: 0x0454,
		ReactiveImportPowerDemandL3: 0x0456,
		ReactiveImportPowerDemand:   0x0458,
		ReactiveExportPowerDemandL1: 0x045A, // A phase reverse reactive demand
		ReactiveExportPowerDemandL2: 0x045C,
		ReactiveExportPowerDemandL3: 0x045E,
		ReactiveExportPowerDemand:   0x0460,
		ReactivePowerDemandL1:       0x0462, // A phase total reactive demand
		ReactivePowerDemandL2:       0x0464,
		ReactivePowerDemandL3:       0x0466,
		ReactivePowerDemand:         0x0468,

		// maximum demand since the demand values were last cleared, 0.1 W / 0.1 var
		ImportPowerMaxDemandL1:         0x046A, // A phase forward active maximum demand
		ImportPowerMaxDemandL2:         0x046C,
		ImportPowerMaxDemandL3:         0x046E,
		ImportPowerMaxDemand:           0x0470,
		ExportPowerMaxDemandL1:         0x0472, // A phase reverse active maximum demand
		ExportPowerMaxDemandL2:         0x0474,
		ExportPowerMaxDemandL3:         0x0476,
		ExportPowerMaxDemand:           0x0478,
		PowerMaxDemandL1:               0x047A, // A phase total active maximum demand
		PowerMaxDemandL2:               0x047C,
		PowerMaxDemandL3:               0x047E,
		PowerMaxDemand:                 0x0480,
		ReactiveImportPowerMaxDemandL1: 0x0482, // A phase forward reactive maximum demand
		ReactiveImportPowerMaxDemandL2: 0x0484,
		ReactiveImportPowerMaxDemandL3: 0x0486,
		ReactiveImportPowerMaxDemand:   0x0488,
		ReactiveExportPowerMaxDemandL1: 0x048A, // A phase reverse reactive maximum demand
		ReactiveExportPowerMaxDemandL2: 0x048C,
		ReactiveExportPowerMaxDemandL3: 0x048E,
		ReactiveExportPowerMaxDemand:   0x0490,
		ReactivePowerMaxDemandL1:       0x0492, // A phase total reactive maximum demand
		ReactivePowerMaxDemandL2:       0x0494,
		ReactivePowerMaxDemandL3:       0x0496,
		ReactivePowerMaxDemand:         0x0498,

		// reactive energy per quadrant, 0.01 kvarh
		ReactiveQ1:   0x014A, // first quadrant total reactive energy
		ReactiveQ2:   0x0154,
		ReactiveQ3:   0x015E,
		ReactiveQ4:   0x0168,
		ReactiveQ1L1: 0x053C, // A phase first quadrant total reactive energy
		ReactiveQ2L1: 0x0546,
		ReactiveQ3L1: 0x0550,
		ReactiveQ4L1: 0x055A,
		ReactiveQ1L2: 0x05A0, // B phase first quadrant total reactive energy
		ReactiveQ2L2: 0x05AA,
		ReactiveQ3L2: 0x05B4,
		ReactiveQ4L2: 0x05BE,
		ReactiveQ1L3: 0x0604, // C phase first quadrant total reactive energy
		ReactiveQ2L3: 0x060E,
		ReactiveQ3L3: 0x0618,
		ReactiveQ4L3: 0x0622,

		// resettable energy counters, 0.01 kWh / 0.01 kvarh
		ResettableSum:              0x062C, // active resettable energy
		ResettableImport:           0x062E, // forward active resettable energy
		ResettableExport:           0x0630, // reverse active resettable energy
		ResettableSumL1:            0x0632,
		ResettableImportL1:         0x0634,
		ResettableExportL1:         0x0636,
		ResettableSumL2:            0x0638,
		ResettableImportL2:         0x063A,
		ResettableExportL2:         0x063C,
		ResettableSumL3:            0x063E,
		ResettableImportL3:         0x0640,
		ResettableExportL3:         0x0642,
		ResettableReactiveSum:      0x0644, // reactive resettable energy
		ResettableReactiveImport:   0x0646, // forward reactive resettable energy
		ResettableReactiveExport:   0x0648, // reverse reactive resettable energy
		ResettableReactiveSumL1:    0x064A,
		ResettableReactiveImportL1: 0x064C,
		ResettableReactiveExportL1: 0x064E,
		ResettableReactiveSumL2:    0x0650,
		ResettableReactiveImportL2: 0x0652,
		ResettableReactiveExportL2: 0x0654,
		ResettableReactiveSumL3:    0x0656,
		ResettableReactiveImportL3: 0x0658,
		ResettableReactiveExportL3: 0x065A,
	}
	return &DS100Producer{Opcodes: ops}
}

func (p *DS100Producer) Description() string {
	return "B+G e-tech DS100"
}

// MaxBlockLen implements BlockProducer. The manual allows reading up to 100 registers at once.
func (p *DS100Producer) MaxBlockLen() uint16 {
	return 100
}

// MaxBlockGap implements BlockProducer. It bridges the T1-T4 tariff registers
// between the energy totals, so each energy block is read in one request.
func (p *DS100Producer) MaxBlockGap() uint16 {
	return 8
}

func (p *DS100Producer) snip(iec Measurement, readlen uint16, transform RTUTransform, scaler ...float64) Operation {
	snip := Operation{
		FuncCode:  ReadHoldingReg,
		OpCode:    p.Opcodes[iec],
		ReadLen:   readlen,
		Transform: transform,
		IEC61850:  iec,
	}

	if len(scaler) > 0 {
		snip.Transform = MakeScaledTransform(snip.Transform, scaler[0])
	}

	return snip
}

// snip16u creates modbus operation for single register
func (p *DS100Producer) snip16u(iec Measurement, scaler ...float64) Operation {
	return p.snip(iec, 1, RTUUint16ToFloat64, scaler...)
}

// snip32u creates modbus operation for double register
func (p *DS100Producer) snip32u(iec Measurement, scaler ...float64) Operation {
	return p.snip(iec, 2, RTUUint32ToFloat64, scaler...)
}

// snip16s creates modbus operation for single register (signed)
func (p *DS100Producer) snip16s(iec Measurement, scaler ...float64) Operation {
	return p.snip(iec, 1, RTUInt16ToFloat64, scaler...)
}

// snip32s creates modbus operation for double register (signed)
func (p *DS100Producer) snip32s(iec Measurement, scaler ...float64) Operation {
	return p.snip(iec, 2, RTUInt32ToFloat64, scaler...)
}

func (p *DS100Producer) Probe() Operation {
	return p.snip32u(Voltage, 1000)
}

func (p *DS100Producer) Produce() (res []Operation) {
	for _, op := range []Measurement{
		VoltageL1, VoltageL2, VoltageL3, Voltage,
		VoltageL1_L2, VoltageL2_L3, VoltageL3_L1, VoltageL_L_avg,
	} {
		res = append(res, p.snip32u(op, 1000))
	}

	// Currents are signed and follow the direction of active power,
	// so they are negative while exporting.
	for _, op := range []Measurement{
		CurrentL1, CurrentL2, CurrentL3, CurrentN, Current,
	} {
		res = append(res, p.snip32s(op, 1000))
	}

	for _, op := range []Measurement{
		PowerL1, PowerL2, PowerL3, Power,
		ReactivePowerL1, ReactivePowerL2, ReactivePowerL3, ReactivePower,
	} {
		res = append(res, p.snip32s(op, 1))
	}

	for _, op := range []Measurement{
		ApparentPowerL1, ApparentPowerL2, ApparentPowerL3, ApparentPower,
	} {
		res = append(res, p.snip32u(op, 1))
	}

	for _, op := range []Measurement{
		ImportL1, ImportL2, ImportL3, Import,
		ExportL1, ExportL2, ExportL3, Export,
		ReactiveImportL1, ReactiveImportL2, ReactiveImportL3, ReactiveImport,
		ReactiveExportL1, ReactiveExportL2, ReactiveExportL3, ReactiveExport,
	} {
		res = append(res, p.snip32u(op, 100))
	}

	// Total energy follows the meter's combination code (register 0x100F).
	// With Import - Export it is signed and turns negative once export exceeds import.
	for _, op := range []Measurement{
		SumL1, SumL2, SumL3, Sum,
		ReactiveSumL1, ReactiveSumL2, ReactiveSumL3, ReactiveSum,
	} {
		res = append(res, p.snip32s(op, 100))
	}

	// reactive energy per quadrant and resettable counters for import and export
	for _, op := range []Measurement{
		ReactiveQ1L1, ReactiveQ1L2, ReactiveQ1L3, ReactiveQ1,
		ReactiveQ2L1, ReactiveQ2L2, ReactiveQ2L3, ReactiveQ2,
		ReactiveQ3L1, ReactiveQ3L2, ReactiveQ3L3, ReactiveQ3,
		ReactiveQ4L1, ReactiveQ4L2, ReactiveQ4L3, ReactiveQ4,
		ResettableImportL1, ResettableImportL2, ResettableImportL3, ResettableImport,
		ResettableExportL1, ResettableExportL2, ResettableExportL3, ResettableExport,
		ResettableReactiveImportL1, ResettableReactiveImportL2, ResettableReactiveImportL3, ResettableReactiveImport,
		ResettableReactiveExportL1, ResettableReactiveExportL2, ResettableReactiveExportL3, ResettableReactiveExport,
	} {
		res = append(res, p.snip32u(op, 100))
	}

	// resettable totals follow the combination code like the other totals
	for _, op := range []Measurement{
		ResettableSumL1, ResettableSumL2, ResettableSumL3, ResettableSum,
		ResettableReactiveSumL1, ResettableReactiveSumL2, ResettableReactiveSumL3, ResettableReactiveSum,
	} {
		res = append(res, p.snip32s(op, 100))
	}

	// demand for import and export
	for _, op := range []Measurement{
		ImportPowerDemandL1, ImportPowerDemandL2, ImportPowerDemandL3, ImportPowerDemand,
		ExportPowerDemandL1, ExportPowerDemandL2, ExportPowerDemandL3, ExportPowerDemand,
		ReactiveImportPowerDemandL1, ReactiveImportPowerDemandL2, ReactiveImportPowerDemandL3, ReactiveImportPowerDemand,
		ReactiveExportPowerDemandL1, ReactiveExportPowerDemandL2, ReactiveExportPowerDemandL3, ReactiveExportPowerDemand,
		ImportPowerMaxDemandL1, ImportPowerMaxDemandL2, ImportPowerMaxDemandL3, ImportPowerMaxDemand,
		ExportPowerMaxDemandL1, ExportPowerMaxDemandL2, ExportPowerMaxDemandL3, ExportPowerMaxDemand,
		ReactiveImportPowerMaxDemandL1, ReactiveImportPowerMaxDemandL2, ReactiveImportPowerMaxDemandL3, ReactiveImportPowerMaxDemand,
		ReactiveExportPowerMaxDemandL1, ReactiveExportPowerMaxDemandL2, ReactiveExportPowerMaxDemandL3, ReactiveExportPowerMaxDemand,
	} {
		res = append(res, p.snip32u(op, 10))
	}

	// total demand is read signed, like the total energy
	for _, op := range []Measurement{
		PowerDemandL1, PowerDemandL2, PowerDemandL3, PowerDemand,
		ReactivePowerDemandL1, ReactivePowerDemandL2, ReactivePowerDemandL3, ReactivePowerDemand,
		PowerMaxDemandL1, PowerMaxDemandL2, PowerMaxDemandL3, PowerMaxDemand,
		ReactivePowerMaxDemandL1, ReactivePowerMaxDemandL2, ReactivePowerMaxDemandL3, ReactivePowerMaxDemand,
	} {
		res = append(res, p.snip32s(op, 10))
	}

	for _, op := range []Measurement{
		FrequencyL1, FrequencyL2, FrequencyL3, Frequency,
	} {
		res = append(res, p.snip16u(op, 10))
	}

	for _, op := range []Measurement{
		CosphiL1, CosphiL2, CosphiL3, Cosphi,
	} {
		res = append(res, p.snip16s(op, 1000))
	}

	return res
}
