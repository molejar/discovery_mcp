package dwf

// DeviceController manages device lifecycle and info.
type DeviceController interface {
	// Open connects to a Digilent device.
	// device can be "" (first available), "Analog Discovery 2", "Digital Discovery", etc.
	// config selects the device configuration index (0 for default).
	Open(device string, config int) (*DeviceInfo, error)

	// Close disconnects from the device and frees resources.
	Close() error

	// Temperature returns the board temperature in °C.
	Temperature() (float64, error)
}

// Oscilloscope controls the analog input (scope) instrument.
type Oscilloscope interface {
	// Open initializes the oscilloscope with the given configuration.
	Open(cfg ScopeConfig) error

	// Measure reads a single voltage sample from the specified channel (1-based).
	Measure(channel int) (float64, error)

	// SetTrigger configures the oscilloscope trigger.
	SetTrigger(cfg TriggerConfig) error

	// Record captures a buffer of samples from the specified channel (1-based).
	// Returns the recorded voltage samples.
	Record(channel int) ([]float64, error)

	// Close resets the oscilloscope.
	Close() error
}

// WavegenDriver controls the analog output (wavegen) instrument.
type WavegenDriver interface {
	// Generate starts an analog waveform on the specified channel.
	Generate(cfg WavegenConfig) error

	// Enable starts output on the given channel (1-based).
	Enable(channel int) error

	// Disable stops output on the given channel (1-based).
	Disable(channel int) error

	// Close resets the wavegen for the given channel (1-based).
	Close(channel int) error
}

// PowerSupply controls the onboard power supplies.
type PowerSupply interface {
	// Switch configures and enables/disables the power supplies.
	Switch(cfg SuppliesConfig) error

	// Close resets the power supply instrument.
	Close() error
}

// DigitalMultimeter controls the DMM instrument (available on some devices).
type DigitalMultimeter interface {
	// Open initializes the DMM.
	Open() error

	// Measure performs a measurement in the given mode.
	// range_ specifies the measurement range (0 = auto).
	// highImpedance sets 10GΩ input (true) vs 10MΩ (false) for DC voltage.
	Measure(mode DMMMode, range_ float64, highImpedance bool) (float64, error)

	// Close resets the DMM.
	Close() error
}

// LogicAnalyzer controls the digital input (logic analyzer) instrument.
type LogicAnalyzer interface {
	// Open initializes the logic analyzer with the given configuration.
	Open(cfg LogicConfig) error

	// SetTrigger configures the logic analyzer trigger.
	SetTrigger(cfg LogicTriggerConfig) error

	// Record captures digital samples from the specified DIO channel.
	// Returns the recorded logic values.
	Record(channel int) ([]uint16, error)

	// Close resets the logic analyzer.
	Close() error
}

// PatternGenerator controls the digital output (pattern generator) instrument.
type PatternGenerator interface {
	// Generate starts a digital pattern on the configured channel.
	Generate(cfg PatternConfig) error

	// Enable starts output on the given DIO channel.
	Enable(channel int) error

	// Disable stops output on the given DIO channel.
	Disable(channel int) error

	// Close resets the pattern generator.
	Close() error
}

// StaticIO controls the static digital I/O pins.
type StaticIO interface {
	// SetMode sets a DIO line as input (false) or output (true).
	SetMode(channel int, output bool) error

	// GetState reads the state of a DIO line (true = HIGH).
	GetState(channel int) (bool, error)

	// SetState sets a DIO line high (true) or low (false).
	SetState(channel int, value bool) error

	// SetCurrent limits the DIO output current in mA.
	// Valid values: 2, 4, 6, 8, 12, 16 mA.
	SetCurrent(current float64) error

	// SetPull configures pull-up/pull-down for a DIO channel.
	SetPull(channel int, direction PullDirection) error

	// Close resets the static I/O.
	Close() error
}

// UART controls the UART protocol instrument.
type UART interface {
	// Open initializes UART communication.
	Open(cfg UARTConfig) error

	// Read receives data from the UART RX line.
	Read() ([]byte, error)

	// Write sends data through the UART TX line.
	Write(data []byte) error

	// Close resets the UART interface.
	Close() error
}

// SPI controls the SPI protocol instrument.
type SPI interface {
	// Open initializes SPI communication.
	Open(cfg SPIConfig) error

	// Read receives count bytes from SPI.
	// cs is the chip select DIO line.
	Read(count int, cs int) ([]byte, error)

	// Write sends data through SPI.
	// cs is the chip select DIO line.
	Write(data []byte, cs int) error

	// Exchange simultaneously sends txData and receives rxCount bytes.
	// cs is the chip select DIO line.
	Exchange(txData []byte, rxCount int, cs int) ([]byte, error)

	// Close resets the SPI interface.
	Close() error
}

// I2C controls the I2C protocol instrument.
type I2C interface {
	// Open initializes I2C communication.
	Open(cfg I2CConfig) error

	// Read receives count bytes from the given 7-bit address.
	Read(count int, address int) ([]byte, error)

	// Write sends data to the given 7-bit address.
	Write(data []byte, address int) error

	// Exchange sends txData then receives rxCount bytes from the given address.
	Exchange(txData []byte, rxCount int, address int) ([]byte, error)

	// Close resets the I2C interface.
	Close() error
}

// DiscoveryDevice aggregates all instrument interfaces for a connected device.
type DiscoveryDevice interface {
	DeviceController
	Scope() Oscilloscope
	Wavegen() WavegenDriver
	Supply() PowerSupply
	DMM() DigitalMultimeter
	Logic() LogicAnalyzer
	Pattern() PatternGenerator
	Static() StaticIO
	UARTProtocol() UART
	SPIProtocol() SPI
	I2CProtocol() I2C
}
