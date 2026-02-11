// Package dwf provides Go interfaces and types for controlling Digilent
// DWF SDK instruments (Analog Discovery 2, Digital Discovery, etc.)
// via CGo bindings to libdwf.
package dwf

// WavegenFunc enumerates analog waveform generator function types.
type WavegenFunc int

const (
	FuncDC        WavegenFunc = 0
	FuncSine      WavegenFunc = 1
	FuncSquare    WavegenFunc = 2
	FuncTriangle  WavegenFunc = 3
	FuncRampUp    WavegenFunc = 4
	FuncRampDown  WavegenFunc = 5
	FuncNoise     WavegenFunc = 6
	FuncPulse     WavegenFunc = 7
	FuncTrapezium WavegenFunc = 8
	FuncSinePower WavegenFunc = 9
	FuncCustom    WavegenFunc = 30
)

// TriggerSource enumerates trigger source types.
type TriggerSource int

const (
	TrigSrcNone              TriggerSource = 0
	TrigSrcPC                TriggerSource = 1
	TrigSrcDetectorAnalogIn  TriggerSource = 2
	TrigSrcDetectorDigitalIn TriggerSource = 3
	TrigSrcAnalogIn          TriggerSource = 4
	TrigSrcDigitalIn         TriggerSource = 5
	TrigSrcDigitalOut        TriggerSource = 6
	TrigSrcAnalogOut1        TriggerSource = 7
	TrigSrcAnalogOut2        TriggerSource = 8
	TrigSrcAnalogOut3        TriggerSource = 9
	TrigSrcAnalogOut4        TriggerSource = 10
	TrigSrcExternal1         TriggerSource = 11
	TrigSrcExternal2         TriggerSource = 12
	TrigSrcExternal3         TriggerSource = 13
	TrigSrcExternal4         TriggerSource = 14
)

// DMMMode enumerates digital multimeter measurement modes.
type DMMMode int

const (
	DMMModeACVoltage     DMMMode = 0
	DMMModeDCVoltage     DMMMode = 1
	DMMModeACCurrent     DMMMode = 2
	DMMModeDCCurrent     DMMMode = 3
	DMMModeResistance    DMMMode = 4
	DMMModeContinuity    DMMMode = 5
	DMMModeDiode         DMMMode = 6
	DMMModeTemperature   DMMMode = 7
	DMMModeACLowCurrent  DMMMode = 8
	DMMModeDCLowCurrent  DMMMode = 9
	DMMModeACHighCurrent DMMMode = 10
	DMMModeDCHighCurrent DMMMode = 11
)

// DigitalOutType enumerates pattern generator output types.
type DigitalOutType int

const (
	DigitalOutTypePulse  DigitalOutType = 0
	DigitalOutTypeCustom DigitalOutType = 1
	DigitalOutTypeRandom DigitalOutType = 2
)

// DigitalOutIdle enumerates idle states for digital outputs.
type DigitalOutIdle int

const (
	DigitalOutIdleInit DigitalOutIdle = 0
	DigitalOutIdleLow  DigitalOutIdle = 1
	DigitalOutIdleHigh DigitalOutIdle = 2
	DigitalOutIdleZet  DigitalOutIdle = 3
)

// TriggerSlope enumerates trigger edge types.
type TriggerSlope int

const (
	TriggerSlopeRise   TriggerSlope = 0
	TriggerSlopeFall   TriggerSlope = 1
	TriggerSlopeEither TriggerSlope = 2
)

// PullDirection enumerates pull-up/pull-down directions for Static I/O.
type PullDirection int

const (
	PullUp   PullDirection = 1
	PullDown PullDirection = 0
	PullIdle PullDirection = -1
)

// DeviceInfo holds information about a connected Digilent device.
type DeviceInfo struct {
	// Handle is the internal device handle used for all API calls.
	Handle int
	// Name is the human-readable device name (e.g., "Analog Discovery 2").
	Name string
	// SerialNumber is the unique device serial.
	SerialNumber string
	// Version is the DWF SDK version string.
	Version string
	// AnalogInChannels is the number of oscilloscope channels.
	AnalogInChannels int
	// AnalogOutChannels is the number of waveform generator channels.
	AnalogOutChannels int
	// DigitalInChannels is the number of logic analyzer channels.
	DigitalInChannels int
	// DigitalOutChannels is the number of pattern generator channels.
	DigitalOutChannels int
	// MaxAnalogInBufferSize is the maximum oscilloscope buffer.
	MaxAnalogInBufferSize int
	// MaxAnalogInResolution is the ADC bit resolution.
	MaxAnalogInResolution int
}

// DeviceConfig holds information about one device configuration preset.
// Different configurations trade off resources between instruments.
type DeviceConfig struct {
	// AnalogInChannels is the number of oscilloscope channels.
	AnalogInChannels int
	// AnalogOutChannels is the number of waveform generator channels.
	AnalogOutChannels int
	// AnalogIOChannels is the number of analog I/O channels.
	AnalogIOChannels int
	// DigitalInChannels is the number of logic analyzer channels.
	DigitalInChannels int
	// DigitalOutChannels is the number of pattern generator channels.
	DigitalOutChannels int
	// DigitalIOChannels is the number of digital I/O channels.
	DigitalIOChannels int
	// AnalogInBufferSize is the oscilloscope buffer per channel.
	AnalogInBufferSize int
	// AnalogOutBufferSize is the wavegen buffer per channel.
	AnalogOutBufferSize int
	// DigitalInBufferSize is the logic analyzer buffer.
	DigitalInBufferSize int
	// DigitalOutBufferSize is the pattern generator buffer.
	DigitalOutBufferSize int
}

// EnumDevice holds information about a discovered device before opening.
type EnumDevice struct {
	// Index is the enumeration index used to open this device.
	Index int
	// DeviceName is the product name (e.g. "Analog Discovery 2").
	DeviceName string
	// UserName is the user-assigned name, if set.
	UserName string
	// SerialNumber is the device serial number.
	SerialNumber string
	// IsOpened indicates whether the device is already in use.
	IsOpened bool
}

// ScopeConfig configures the oscilloscope before acquisition.
type ScopeConfig struct {
	// SamplingFrequency in Hz (default 20 MHz).
	SamplingFrequency float64
	// BufferSize in samples; 0 means maximum.
	BufferSize int
	// OffsetVoltage in Volts (default 0).
	OffsetVoltage float64
	// AmplitudeRange in Volts (default ±5 V).
	AmplitudeRange float64
}

// TriggerConfig configures the oscilloscope trigger.
type TriggerConfig struct {
	// Enable enables/disables the trigger.
	Enable bool
	// Source is the trigger source.
	Source TriggerSource
	// Channel is the trigger channel (1-based for analog, 0-based for digital).
	Channel int
	// Timeout is the auto-trigger timeout in seconds; 0 disables.
	Timeout float64
	// EdgeRising selects rising (true) or falling (false) edge.
	EdgeRising bool
	// Level is the trigger level in Volts.
	Level float64
}

// WavegenConfig configures waveform generation on an analog output channel.
type WavegenConfig struct {
	// Channel is the wavegen channel (1 or 2).
	Channel int
	// Function is the waveform type.
	Function WavegenFunc
	// Offset in Volts.
	Offset float64
	// Frequency in Hz.
	Frequency float64
	// Amplitude in Volts.
	Amplitude float64
	// Symmetry as percentage (0-100).
	Symmetry float64
	// Wait time before start in seconds.
	Wait float64
	// RunTime in seconds; 0 means infinite.
	RunTime float64
	// Repeat count; 0 means infinite.
	Repeat int
	// CustomData holds voltages when Function=FuncCustom.
	CustomData []float64
}

// SuppliesConfig configures the power supply voltages and states.
type SuppliesConfig struct {
	// MasterState enables/disables all supplies.
	MasterState bool
	// PositiveState enables/disables the positive supply.
	PositiveState bool
	// NegativeState enables/disables the negative supply.
	NegativeState bool
	// State enables/disables the digital/6V supply.
	State bool
	// PositiveVoltage in Volts.
	PositiveVoltage float64
	// NegativeVoltage in Volts.
	NegativeVoltage float64
	// Voltage for the digital/6V rail in Volts.
	Voltage float64
	// PositiveCurrent limit in Amps.
	PositiveCurrent float64
	// NegativeCurrent limit in Amps.
	NegativeCurrent float64
	// Current limit for the digital/6V rail in Amps.
	Current float64
}

// LogicConfig configures the logic analyzer before acquisition.
type LogicConfig struct {
	// SamplingFrequency in Hz (default 100 MHz).
	SamplingFrequency float64
	// BufferSize in samples; 0 means maximum.
	BufferSize int
}

// LogicTriggerConfig configures the logic analyzer trigger.
type LogicTriggerConfig struct {
	// Enable enables/disables the trigger.
	Enable bool
	// Channel is the DIO line number.
	Channel int
	// Position is the prefill size (samples before trigger).
	Position int
	// Timeout is the auto-trigger timeout in seconds.
	Timeout float64
	// RisingEdge selects rising (true) or falling (false) edge.
	RisingEdge bool
	// LengthMin is the minimum trigger sequence duration in seconds.
	LengthMin float64
	// LengthMax is the maximum trigger sequence duration in seconds.
	LengthMax float64
	// Count is the trigger event counter.
	Count int
}

// PatternConfig configures the digital pattern generator.
type PatternConfig struct {
	// Channel is the DIO line number.
	Channel int
	// Function is the output type (pulse, custom, random).
	Function DigitalOutType
	// Frequency in Hz.
	Frequency float64
	// DutyCycle as percentage (for Pulse function).
	DutyCycle float64
	// Data is the custom bit pattern (for Custom function).
	Data []uint16
	// Wait time before start in seconds.
	Wait float64
	// Repeat count; 0 means infinite.
	Repeat int
	// RunTime in seconds; 0 means infinite, -1 means auto.
	RunTime int
	// IdleState for the output when not active.
	IdleState DigitalOutIdle
	// TriggerEnabled includes trigger in repeat cycle.
	TriggerEnabled bool
	// TriggerSource for the pattern generator.
	TriggerSource TriggerSource
	// TriggerEdgeRising selects rising (true) or falling (false) edge.
	TriggerEdgeRising bool
}

// UARTConfig configures UART communication.
type UARTConfig struct {
	// RX is the DIO line for receiving data.
	RX int
	// TX is the DIO line for transmitting data.
	TX int
	// BaudRate in bits/s (default 9600).
	BaudRate int
	// Parity: 0=none, 1=odd, 2=even.
	Parity int
	// DataBits count (default 8).
	DataBits int
	// StopBits count (default 1).
	StopBits int
}

// SPIConfig configures SPI communication.
type SPIConfig struct {
	// CS is the DIO line for chip select.
	CS int
	// SCK is the DIO line for serial clock.
	SCK int
	// MISO is the DIO line for master-in/slave-out (-1 to skip).
	MISO int
	// MOSI is the DIO line for master-out/slave-in (-1 to skip).
	MOSI int
	// ClockFrequency in Hz (default 1 MHz).
	ClockFrequency float64
	// Mode is the SPI mode (0-3).
	Mode int
	// MSBFirst sets bit order; true = MSB first.
	MSBFirst bool
}

// I2CConfig configures I2C communication.
type I2CConfig struct {
	// SDA is the DIO line for data.
	SDA int
	// SCL is the DIO line for clock.
	SCL int
	// ClockRate in Hz (default 100 kHz).
	ClockRate float64
	// Stretching enables/disables clock stretching.
	Stretching bool
}
