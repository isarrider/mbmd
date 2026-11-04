package meters

import (
	"fmt"
	"time"
)

// MeasurementResult is the result of modbus read operation
type MeasurementResult struct {
	Measurement
	Value     float64
	Timestamp time.Time
}

func (r MeasurementResult) String() string {
	_, unit := r.DescriptionAndUnit()
	return fmt.Sprintf("%s: %.2f%s", r.Measurement.String(), r.Value, unit)
}

// Measurement is the type of measurement, i.e. the physical property being measued in common notation
type Measurement int

//go:generate enumer -type=Measurement
const (
	_ Measurement = iota

	Frequency
	FrequencyL1
	FrequencyL2
	FrequencyL3

	Current
	CurrentL1
	CurrentL2
	CurrentL3

	// phases and sums
	Voltage
	VoltageL1
	VoltageL2
	VoltageL3
	VoltageL1_L2
	VoltageL2_L3
	VoltageL3_L1
	VoltageL_N_avg
	VoltageL_L_avg

	Power // synonymous ActivePower
	PowerL1
	PowerL2
	PowerL3

	ImportPower
	ImportPowerL1
	ImportPowerL2
	ImportPowerL3

	ExportPower
	ExportPowerL1
	ExportPowerL2
	ExportPowerL3

	ReactivePower
	ReactivePowerL1
	ReactivePowerL2
	ReactivePowerL3

	ApparentPower
	ApparentPowerL1
	ApparentPowerL2
	ApparentPowerL3

	Cosphi
	CosphiL1
	CosphiL2
	CosphiL3

	THD
	THDL1
	THDL2
	THDL3

	ThreePhase_Vec_A

	// energy
	Sum // synonymous ActiveEnergy
	SumT1
	SumT2
	SumL1
	SumL2
	SumL3

	Import
	ImportT1
	ImportT2
	ImportL1
	ImportL2
	ImportL3

	Export
	ExportT1
	ExportT2
	ExportL1
	ExportL2
	ExportL3

	ReactiveSum
	ReactiveSumT1
	ReactiveSumT2
	ReactiveSumL1
	ReactiveSumL2
	ReactiveSumL3

	ReactiveImport
	ReactiveImportT1
	ReactiveImportT2
	ReactiveImportL1
	ReactiveImportL2
	ReactiveImportL3

	ReactiveExport
	ReactiveExportT1
	ReactiveExportT2
	ReactiveExportL1
	ReactiveExportL2
	ReactiveExportL3

	// DC
	DCCurrent
	DCVoltage
	DCPower
	HeatSinkTemp
	CabinetTemp

	// Strings
	DCCurrentS1
	DCVoltageS1
	DCPowerS1
	DCEnergyS1
	DCCurrentS2
	DCVoltageS2
	DCPowerS2
	DCEnergyS2
	DCCurrentS3
	DCVoltageS3
	DCPowerS3
	DCEnergyS3
	DCCurrentS4
	DCVoltageS4
	DCPowerS4
	DCEnergyS4

	// Battery
	ChargeState
	BatteryVoltage

	PhaseAngle

	// Demand: average power over the meter's demand period, and its maximum since the last reset
	ImportPowerDemand
	ImportPowerDemandL1
	ImportPowerDemandL2
	ImportPowerDemandL3
	ExportPowerDemand
	ExportPowerDemandL1
	ExportPowerDemandL2
	ExportPowerDemandL3
	PowerDemand
	PowerDemandL1
	PowerDemandL2
	PowerDemandL3
	ReactiveImportPowerDemand
	ReactiveImportPowerDemandL1
	ReactiveImportPowerDemandL2
	ReactiveImportPowerDemandL3
	ReactiveExportPowerDemand
	ReactiveExportPowerDemandL1
	ReactiveExportPowerDemandL2
	ReactiveExportPowerDemandL3
	ReactivePowerDemand
	ReactivePowerDemandL1
	ReactivePowerDemandL2
	ReactivePowerDemandL3
	ImportPowerMaxDemand
	ImportPowerMaxDemandL1
	ImportPowerMaxDemandL2
	ImportPowerMaxDemandL3
	ExportPowerMaxDemand
	ExportPowerMaxDemandL1
	ExportPowerMaxDemandL2
	ExportPowerMaxDemandL3
	PowerMaxDemand
	PowerMaxDemandL1
	PowerMaxDemandL2
	PowerMaxDemandL3
	ReactiveImportPowerMaxDemand
	ReactiveImportPowerMaxDemandL1
	ReactiveImportPowerMaxDemandL2
	ReactiveImportPowerMaxDemandL3
	ReactiveExportPowerMaxDemand
	ReactiveExportPowerMaxDemandL1
	ReactiveExportPowerMaxDemandL2
	ReactiveExportPowerMaxDemandL3
	ReactivePowerMaxDemand
	ReactivePowerMaxDemandL1
	ReactivePowerMaxDemandL2
	ReactivePowerMaxDemandL3

	// Reactive energy per quadrant
	ReactiveQ1
	ReactiveQ1L1
	ReactiveQ1L2
	ReactiveQ1L3
	ReactiveQ2
	ReactiveQ2L1
	ReactiveQ2L2
	ReactiveQ2L3
	ReactiveQ3
	ReactiveQ3L1
	ReactiveQ3L2
	ReactiveQ3L3
	ReactiveQ4
	ReactiveQ4L1
	ReactiveQ4L2
	ReactiveQ4L3

	// Resettable energy counters
	ResettableSum
	ResettableSumL1
	ResettableSumL2
	ResettableSumL3
	ResettableImport
	ResettableImportL1
	ResettableImportL2
	ResettableImportL3
	ResettableExport
	ResettableExportL1
	ResettableExportL2
	ResettableExportL3
	ResettableReactiveSum
	ResettableReactiveSumL1
	ResettableReactiveSumL2
	ResettableReactiveSumL3
	ResettableReactiveImport
	ResettableReactiveImportL1
	ResettableReactiveImportL2
	ResettableReactiveImportL3
	ResettableReactiveExport
	ResettableReactiveExportL1
	ResettableReactiveExportL2
	ResettableReactiveExportL3
)

var iec = map[Measurement][]string{
	Frequency:        {"Frequency", "Hz"},
	FrequencyL1:      {"L1 Frequency", "Hz"},
	FrequencyL2:      {"L2 Frequency", "Hz"},
	FrequencyL3:      {"L3 Frequency", "Hz"},
	Current:          {"Current", "A"},
	CurrentL1:        {"L1 Current", "A"},
	CurrentL2:        {"L2 Current", "A"},
	CurrentL3:        {"L3 Current", "A"},
	Voltage:          {"Voltage", "V"},
	VoltageL1:        {"L1 Voltage", "V"},
	VoltageL2:        {"L2 Voltage", "V"},
	VoltageL3:        {"L3 Voltage", "V"},
	VoltageL1_L2:     {"L1 to L2 Voltage", "V"},
	VoltageL2_L3:     {"L2 to L3 Voltage", "V"},
	VoltageL3_L1:     {"L3 to L1 Voltage", "V"},
	VoltageL_N_avg:   {"L to N average Voltage", "V"},
	VoltageL_L_avg:   {"L to L average Voltage", "V"},
	Power:            {"Power", "W"},
	PowerL1:          {"L1 Power", "W"},
	PowerL2:          {"L2 Power", "W"},
	PowerL3:          {"L3 Power", "W"},
	ImportPower:      {"Import Power", "W"},
	ImportPowerL1:    {"L1 Import Power", "W"},
	ImportPowerL2:    {"L2 Import Power", "W"},
	ImportPowerL3:    {"L3 Import Power", "W"},
	ExportPower:      {"Export Power", "W"},
	ExportPowerL1:    {"L1 Export Power", "W"},
	ExportPowerL2:    {"L2 Export Power", "W"},
	ExportPowerL3:    {"L3 Export Power", "W"},
	ReactivePower:    {"Reactive Power", "var"},
	ReactivePowerL1:  {"L1 Reactive Power", "var"},
	ReactivePowerL2:  {"L2 Reactive Power", "var"},
	ReactivePowerL3:  {"L3 Reactive Power", "var"},
	ApparentPower:    {"Apparent Power", "VA"},
	ApparentPowerL1:  {"L1 Apparent Power", "VA"},
	ApparentPowerL2:  {"L2 Apparent Power", "VA"},
	ApparentPowerL3:  {"L3 Apparent Power", "VA"},
	Cosphi:           {"Cosphi"},
	CosphiL1:         {"L1 Cosphi"},
	CosphiL2:         {"L2 Cosphi"},
	CosphiL3:         {"L3 Cosphi"},
	THD:              {"Average voltage to neutral THD", "%"},
	THDL1:            {"L1 Voltage to neutral THD", "%"},
	THDL2:            {"L2 Voltage to neutral THD", "%"},
	THDL3:            {"L3 Voltage to neutral THD", "%"},
	ThreePhase_Vec_A: {"Three Phase Vector Current", "%"},
	Sum:              {"Total Sum", "kWh"},
	SumT1:            {"Tariff 1 Sum", "kWh"},
	SumT2:            {"Tariff 2 Sum", "kWh"},
	SumL1:            {"L1 Sum", "kWh"},
	SumL2:            {"L2 Sum", "kWh"},
	SumL3:            {"L3 Sum", "kWh"},
	Import:           {"Total Import", "kWh"},
	ImportT1:         {"Tariff 1 Import", "kWh"},
	ImportT2:         {"Tariff 2 Import", "kWh"},
	ImportL1:         {"L1 Import", "kWh"},
	ImportL2:         {"L2 Import", "kWh"},
	ImportL3:         {"L3 Import", "kWh"},
	Export:           {"Total Export", "kWh"},
	ExportT1:         {"Tariff 1 Export", "kWh"},
	ExportT2:         {"Tariff 2 Export", "kWh"},
	ExportL1:         {"L1 Export", "kWh"},
	ExportL2:         {"L2 Export", "kWh"},
	ExportL3:         {"L3 Export", "kWh"},
	ReactiveSum:      {"Total Reactive", "kvarh"},
	ReactiveSumT1:    {"Tariff 1 Reactive", "kvarh"},
	ReactiveSumT2:    {"Tariff 2 Reactive", "kvarh"},
	ReactiveSumL1:    {"L1 Reactive", "kvarh"},
	ReactiveSumL2:    {"L2 Reactive", "kvarh"},
	ReactiveSumL3:    {"L3 Reactive", "kvarh"},
	ReactiveImport:   {"Reactive Import", "kvarh"},
	ReactiveImportT1: {"Tariff 1 Reactive Import", "kvarh"},
	ReactiveImportT2: {"Tariff 2 Reactive Import", "kvarh"},
	ReactiveImportL1: {"L1 Reactive Import", "kvarh"},
	ReactiveImportL2: {"L2 Reactive Import", "kvarh"},
	ReactiveImportL3: {"L3 Reactive Import", "kvarh"},
	ReactiveExport:   {"Reactive Export", "kvarh"},
	ReactiveExportT1: {"Tariff 1 Reactive Export", "kvarh"},
	ReactiveExportT2: {"Tariff 2 Reactive Export", "kvarh"},
	ReactiveExportL1: {"L1 Reactive Export", "kvarh"},
	ReactiveExportL2: {"L2 Reactive Export", "kvarh"},
	ReactiveExportL3: {"L3 Reactive Export", "kvarh"},
	DCCurrent:        {"DC Current", "A"},
	DCVoltage:        {"DC Voltage", "V"},
	DCPower:          {"DC Power", "W"},
	HeatSinkTemp:     {"Heat Sink Temperature", "°C"},
	CabinetTemp:      {"Cabinet Temperature", "°C"},
	DCCurrentS1:      {"String 1 Current", "A"},
	DCVoltageS1:      {"String 1 Voltage", "V"},
	DCPowerS1:        {"String 1 Power", "W"},
	DCEnergyS1:       {"String 1 Generation", "kWh"},
	DCCurrentS2:      {"String 2 Current", "A"},
	DCVoltageS2:      {"String 2 Voltage", "V"},
	DCPowerS2:        {"String 2 Power", "W"},
	DCEnergyS2:       {"String 2 Generation", "kWh"},
	DCCurrentS3:      {"String 3 Current", "A"},
	DCVoltageS3:      {"String 3 Voltage", "V"},
	DCPowerS3:        {"String 3 Power", "W"},
	DCEnergyS3:       {"String 3 Generation", "kWh"},
	DCCurrentS4:      {"String 4 Current", "A"},
	DCVoltageS4:      {"String 4 Voltage", "V"},
	DCPowerS4:        {"String 4 Power", "W"},
	DCEnergyS4:       {"String 4 Generation", "kWh"},
	ChargeState:      {"Charge State", "%"},
	BatteryVoltage:   {"Battery Voltage", "V"},
	PhaseAngle:       {"Phase Angle", "°"},

	ImportPowerDemand:              {"Import Power Demand", "W"},
	ImportPowerDemandL1:            {"L1 Import Power Demand", "W"},
	ImportPowerDemandL2:            {"L2 Import Power Demand", "W"},
	ImportPowerDemandL3:            {"L3 Import Power Demand", "W"},
	ExportPowerDemand:              {"Export Power Demand", "W"},
	ExportPowerDemandL1:            {"L1 Export Power Demand", "W"},
	ExportPowerDemandL2:            {"L2 Export Power Demand", "W"},
	ExportPowerDemandL3:            {"L3 Export Power Demand", "W"},
	PowerDemand:                    {"Power Demand", "W"},
	PowerDemandL1:                  {"L1 Power Demand", "W"},
	PowerDemandL2:                  {"L2 Power Demand", "W"},
	PowerDemandL3:                  {"L3 Power Demand", "W"},
	ReactiveImportPowerDemand:      {"Reactive Import Power Demand", "var"},
	ReactiveImportPowerDemandL1:    {"L1 Reactive Import Power Demand", "var"},
	ReactiveImportPowerDemandL2:    {"L2 Reactive Import Power Demand", "var"},
	ReactiveImportPowerDemandL3:    {"L3 Reactive Import Power Demand", "var"},
	ReactiveExportPowerDemand:      {"Reactive Export Power Demand", "var"},
	ReactiveExportPowerDemandL1:    {"L1 Reactive Export Power Demand", "var"},
	ReactiveExportPowerDemandL2:    {"L2 Reactive Export Power Demand", "var"},
	ReactiveExportPowerDemandL3:    {"L3 Reactive Export Power Demand", "var"},
	ReactivePowerDemand:            {"Reactive Power Demand", "var"},
	ReactivePowerDemandL1:          {"L1 Reactive Power Demand", "var"},
	ReactivePowerDemandL2:          {"L2 Reactive Power Demand", "var"},
	ReactivePowerDemandL3:          {"L3 Reactive Power Demand", "var"},
	ImportPowerMaxDemand:           {"Import Power Maximum Demand", "W"},
	ImportPowerMaxDemandL1:         {"L1 Import Power Maximum Demand", "W"},
	ImportPowerMaxDemandL2:         {"L2 Import Power Maximum Demand", "W"},
	ImportPowerMaxDemandL3:         {"L3 Import Power Maximum Demand", "W"},
	ExportPowerMaxDemand:           {"Export Power Maximum Demand", "W"},
	ExportPowerMaxDemandL1:         {"L1 Export Power Maximum Demand", "W"},
	ExportPowerMaxDemandL2:         {"L2 Export Power Maximum Demand", "W"},
	ExportPowerMaxDemandL3:         {"L3 Export Power Maximum Demand", "W"},
	PowerMaxDemand:                 {"Power Maximum Demand", "W"},
	PowerMaxDemandL1:               {"L1 Power Maximum Demand", "W"},
	PowerMaxDemandL2:               {"L2 Power Maximum Demand", "W"},
	PowerMaxDemandL3:               {"L3 Power Maximum Demand", "W"},
	ReactiveImportPowerMaxDemand:   {"Reactive Import Power Maximum Demand", "var"},
	ReactiveImportPowerMaxDemandL1: {"L1 Reactive Import Power Maximum Demand", "var"},
	ReactiveImportPowerMaxDemandL2: {"L2 Reactive Import Power Maximum Demand", "var"},
	ReactiveImportPowerMaxDemandL3: {"L3 Reactive Import Power Maximum Demand", "var"},
	ReactiveExportPowerMaxDemand:   {"Reactive Export Power Maximum Demand", "var"},
	ReactiveExportPowerMaxDemandL1: {"L1 Reactive Export Power Maximum Demand", "var"},
	ReactiveExportPowerMaxDemandL2: {"L2 Reactive Export Power Maximum Demand", "var"},
	ReactiveExportPowerMaxDemandL3: {"L3 Reactive Export Power Maximum Demand", "var"},
	ReactivePowerMaxDemand:         {"Reactive Power Maximum Demand", "var"},
	ReactivePowerMaxDemandL1:       {"L1 Reactive Power Maximum Demand", "var"},
	ReactivePowerMaxDemandL2:       {"L2 Reactive Power Maximum Demand", "var"},
	ReactivePowerMaxDemandL3:       {"L3 Reactive Power Maximum Demand", "var"},

	ReactiveQ1:   {"Q1 Reactive", "kvarh"},
	ReactiveQ1L1: {"L1 Q1 Reactive", "kvarh"},
	ReactiveQ1L2: {"L2 Q1 Reactive", "kvarh"},
	ReactiveQ1L3: {"L3 Q1 Reactive", "kvarh"},
	ReactiveQ2:   {"Q2 Reactive", "kvarh"},
	ReactiveQ2L1: {"L1 Q2 Reactive", "kvarh"},
	ReactiveQ2L2: {"L2 Q2 Reactive", "kvarh"},
	ReactiveQ2L3: {"L3 Q2 Reactive", "kvarh"},
	ReactiveQ3:   {"Q3 Reactive", "kvarh"},
	ReactiveQ3L1: {"L1 Q3 Reactive", "kvarh"},
	ReactiveQ3L2: {"L2 Q3 Reactive", "kvarh"},
	ReactiveQ3L3: {"L3 Q3 Reactive", "kvarh"},
	ReactiveQ4:   {"Q4 Reactive", "kvarh"},
	ReactiveQ4L1: {"L1 Q4 Reactive", "kvarh"},
	ReactiveQ4L2: {"L2 Q4 Reactive", "kvarh"},
	ReactiveQ4L3: {"L3 Q4 Reactive", "kvarh"},

	ResettableSum:              {"Resettable Sum", "kWh"},
	ResettableSumL1:            {"L1 Resettable Sum", "kWh"},
	ResettableSumL2:            {"L2 Resettable Sum", "kWh"},
	ResettableSumL3:            {"L3 Resettable Sum", "kWh"},
	ResettableImport:           {"Resettable Import", "kWh"},
	ResettableImportL1:         {"L1 Resettable Import", "kWh"},
	ResettableImportL2:         {"L2 Resettable Import", "kWh"},
	ResettableImportL3:         {"L3 Resettable Import", "kWh"},
	ResettableExport:           {"Resettable Export", "kWh"},
	ResettableExportL1:         {"L1 Resettable Export", "kWh"},
	ResettableExportL2:         {"L2 Resettable Export", "kWh"},
	ResettableExportL3:         {"L3 Resettable Export", "kWh"},
	ResettableReactiveSum:      {"Resettable Reactive Sum", "kvarh"},
	ResettableReactiveSumL1:    {"L1 Resettable Reactive Sum", "kvarh"},
	ResettableReactiveSumL2:    {"L2 Resettable Reactive Sum", "kvarh"},
	ResettableReactiveSumL3:    {"L3 Resettable Reactive Sum", "kvarh"},
	ResettableReactiveImport:   {"Resettable Reactive Import", "kvarh"},
	ResettableReactiveImportL1: {"L1 Resettable Reactive Import", "kvarh"},
	ResettableReactiveImportL2: {"L2 Resettable Reactive Import", "kvarh"},
	ResettableReactiveImportL3: {"L3 Resettable Reactive Import", "kvarh"},
	ResettableReactiveExport:   {"Resettable Reactive Export", "kvarh"},
	ResettableReactiveExportL1: {"L1 Resettable Reactive Export", "kvarh"},
	ResettableReactiveExportL2: {"L2 Resettable Reactive Export", "kvarh"},
	ResettableReactiveExportL3: {"L3 Resettable Reactive Export", "kvarh"},
}

// MarshalText implements encoding.TextMarshaler
func (m *Measurement) MarshalText() (text []byte, err error) {
	return []byte(m.String()), nil
}

// DescriptionAndUnit returns a measurements human-readable name and its unit
func (m *Measurement) DescriptionAndUnit() (string, string) {
	if details, ok := iec[*m]; ok {
		unit := ""
		description := details[0]
		if len(details) > 1 {
			unit = details[1]
		}
		return description, unit
	}
	return m.String(), ""
}

// Description returns a measurements human-readable name
func (m *Measurement) Description() string {
	description, unit := m.DescriptionAndUnit()
	if unit != "" {
		description = description + " (" + unit + ")"
	}
	return description
}
