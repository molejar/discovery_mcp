// Package server implements the MCP server that exposes DWF SDK
// instruments as MCP tools.
package server

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/molejar/discovery-mcp/dwf"
)

// DiscoveryMCPServer wraps the MCP server and the Discovery device.
type DiscoveryMCPServer struct {
	mcpServer *server.MCPServer
	device    dwf.DiscoveryDevice
}

// New creates and configures a new DiscoveryMCPServer with all tools registered.
func New() *DiscoveryMCPServer {
	return NewWithDevice(dwf.NewDevice())
}

// NewWithDevice creates a DiscoveryMCPServer using the provided DiscoveryDevice.
// This is useful for testing with mock devices.
func NewWithDevice(dev dwf.DiscoveryDevice) *DiscoveryMCPServer {
	s := &DiscoveryMCPServer{
		device: dev,
	}

	s.mcpServer = server.NewMCPServer(
		"discovery-mcp",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	s.registerTools()
	return s
}

// MCPServer returns the underlying mcp-go server for stdio serving.
func (s *DiscoveryMCPServer) MCPServer() *server.MCPServer {
	return s.mcpServer
}

// DeviceInstance returns the Discovery device instance.
func (s *DiscoveryMCPServer) DeviceInstance() dwf.DiscoveryDevice {
	return s.device
}

func (s *DiscoveryMCPServer) registerTools() {
	// ---- Device ----
	s.mcpServer.AddTool(mcp.NewTool("discovery_enumerate",
		mcp.WithDescription("Enumerate all connected Digilent Discovery devices without opening them"),
	), s.handleEnumerate)

	s.mcpServer.AddTool(mcp.NewTool("discovery_device_get_configs",
		mcp.WithDescription("List available hardware configurations for a device without opening it"),
		mcp.WithNumber("device_index", mcp.Description("Device index from enumeration"), mcp.Required()),
	), s.handleDeviceGetConfigs)

	s.mcpServer.AddTool(mcp.NewTool("discovery_device_open",
		mcp.WithDescription("Open a connection to a Digilent Discovery device"),
		mcp.WithString("device", mcp.Description("Device name (empty for first available): 'Analog Discovery 2', 'Digital Discovery', etc.")),
		mcp.WithNumber("config", mcp.Description("Device configuration index (0 for default)")),
	), s.handleDeviceOpen)

	s.mcpServer.AddTool(mcp.NewTool("discovery_device_close",
		mcp.WithDescription("Close the connection to the Discovery device"),
	), s.handleDeviceClose)

	s.mcpServer.AddTool(mcp.NewTool("discovery_device_temperature",
		mcp.WithDescription("Read the board temperature in °C"),
	), s.handleDeviceTemperature)

	// ---- Oscilloscope ----
	s.mcpServer.AddTool(mcp.NewTool("discovery_scope_open",
		mcp.WithDescription("Initialize the oscilloscope"),
		mcp.WithNumber("sampling_frequency", mcp.Description("Sampling frequency in Hz (default 20MHz)")),
		mcp.WithNumber("buffer_size", mcp.Description("Buffer size in samples (0 = maximum)")),
		mcp.WithNumber("offset_voltage", mcp.Description("Offset voltage in Volts")),
		mcp.WithNumber("amplitude_range", mcp.Description("Amplitude range in Volts (e.g. 5 for ±5V)")),
	), s.handleScopeOpen)

	s.mcpServer.AddTool(mcp.NewTool("discovery_scope_measure",
		mcp.WithDescription("Measure a single voltage from an oscilloscope channel"),
		mcp.WithNumber("channel", mcp.Description("Oscilloscope channel (1-based)"), mcp.Required()),
	), s.handleScopeMeasure)

	s.mcpServer.AddTool(mcp.NewTool("discovery_scope_trigger",
		mcp.WithDescription("Configure the oscilloscope trigger"),
		mcp.WithBoolean("enable", mcp.Description("Enable/disable trigger")),
		mcp.WithNumber("source", mcp.Description("Trigger source (0=none, 2=analog_in, 3=digital_in, 11-14=external)")),
		mcp.WithNumber("channel", mcp.Description("Trigger channel (1-based for analog)")),
		mcp.WithNumber("timeout", mcp.Description("Auto-trigger timeout in seconds")),
		mcp.WithBoolean("edge_rising", mcp.Description("Rising edge (true) or falling edge (false)")),
		mcp.WithNumber("level", mcp.Description("Trigger level in Volts")),
	), s.handleScopeTrigger)

	s.mcpServer.AddTool(mcp.NewTool("discovery_scope_record",
		mcp.WithDescription("Record an analog signal buffer"),
		mcp.WithNumber("channel", mcp.Description("Oscilloscope channel (1-based)"), mcp.Required()),
	), s.handleScopeRecord)

	s.mcpServer.AddTool(mcp.NewTool("discovery_scope_close",
		mcp.WithDescription("Reset the oscilloscope instrument"),
	), s.handleScopeClose)

	// ---- Wavegen ----
	s.mcpServer.AddTool(mcp.NewTool("discovery_wavegen_generate",
		mcp.WithDescription("Generate an analog waveform"),
		mcp.WithNumber("channel", mcp.Description("Wavegen channel (1 or 2)"), mcp.Required()),
		mcp.WithNumber("function", mcp.Description("Wavegen: 0=DC,1=sine,2=square,3=triangle,4=ramp_up,5=ramp_down,6=noise,7=pulse,30=custom"), mcp.Required()),
		mcp.WithNumber("offset", mcp.Description("DC offset in Volts")),
		mcp.WithNumber("frequency", mcp.Description("Frequency in Hz")),
		mcp.WithNumber("amplitude", mcp.Description("Amplitude in Volts")),
		mcp.WithNumber("symmetry", mcp.Description("Symmetry in % (0-100)")),
		mcp.WithNumber("wait", mcp.Description("Wait time before start in seconds")),
		mcp.WithNumber("run_time", mcp.Description("Run time in seconds (0 = infinite)")),
		mcp.WithNumber("repeat", mcp.Description("Repeat count (0 = infinite)")),
	), s.handleWavegenGenerate)

	s.mcpServer.AddTool(mcp.NewTool("discovery_wavegen_enable",
		mcp.WithDescription("Enable a wavegen channel"),
		mcp.WithNumber("channel", mcp.Description("Channel (1-based)"), mcp.Required()),
	), s.handleWavegenEnable)

	s.mcpServer.AddTool(mcp.NewTool("discovery_wavegen_disable",
		mcp.WithDescription("Disable a wavegen channel"),
		mcp.WithNumber("channel", mcp.Description("Channel (1-based)"), mcp.Required()),
	), s.handleWavegenDisable)

	s.mcpServer.AddTool(mcp.NewTool("discovery_wavegen_close",
		mcp.WithDescription("Reset a wavegen channel"),
		mcp.WithNumber("channel", mcp.Description("Channel (1-based)"), mcp.Required()),
	), s.handleWavegenClose)

	// ---- Power Supplies ----
	s.mcpServer.AddTool(mcp.NewTool("discovery_supplies_switch",
		mcp.WithDescription("Configure and switch power supplies on/off"),
		mcp.WithBoolean("master_state", mcp.Description("Master enable/disable")),
		mcp.WithBoolean("positive_state", mcp.Description("Positive supply enable")),
		mcp.WithBoolean("negative_state", mcp.Description("Negative supply enable")),
		mcp.WithBoolean("state", mcp.Description("Digital/6V supply enable")),
		mcp.WithNumber("positive_voltage", mcp.Description("Positive voltage in V")),
		mcp.WithNumber("negative_voltage", mcp.Description("Negative voltage in V")),
		mcp.WithNumber("voltage", mcp.Description("Digital/6V voltage in V")),
		mcp.WithNumber("positive_current", mcp.Description("Positive current limit in A")),
		mcp.WithNumber("negative_current", mcp.Description("Negative current limit in A")),
		mcp.WithNumber("current", mcp.Description("Digital current limit in A")),
	), s.handleSuppliesSwitch)

	s.mcpServer.AddTool(mcp.NewTool("discovery_supplies_close",
		mcp.WithDescription("Reset the power supplies"),
	), s.handleSuppliesClose)

	// ---- DMM ----
	s.mcpServer.AddTool(mcp.NewTool("discovery_dmm_open",
		mcp.WithDescription("Initialize the digital multimeter"),
	), s.handleDMMOpen)

	s.mcpServer.AddTool(mcp.NewTool("discovery_dmm_measure",
		mcp.WithDescription("Measure with the DMM"),
		mcp.WithNumber("mode", mcp.Description("Mode: 0=AC_V,1=DC_V,2=AC_I,3=DC_I,4=resistance,5=continuity,6=diode,7=temp"), mcp.Required()),
		mcp.WithNumber("range", mcp.Description("Measurement range (0 = auto)")),
		mcp.WithBoolean("high_impedance", mcp.Description("High impedance input (10GΩ) for DC voltage")),
	), s.handleDMMMeasure)

	s.mcpServer.AddTool(mcp.NewTool("discovery_dmm_close",
		mcp.WithDescription("Reset the DMM"),
	), s.handleDMMClose)

	// ---- Logic Analyzer ----
	s.mcpServer.AddTool(mcp.NewTool("discovery_logic_open",
		mcp.WithDescription("Initialize the logic analyzer"),
		mcp.WithNumber("sampling_frequency", mcp.Description("Sampling frequency in Hz (default 100MHz)")),
		mcp.WithNumber("buffer_size", mcp.Description("Buffer size (0 = maximum)")),
	), s.handleLogicOpen)

	s.mcpServer.AddTool(mcp.NewTool("discovery_logic_trigger",
		mcp.WithDescription("Configure the logic analyzer trigger"),
		mcp.WithBoolean("enable", mcp.Description("Enable/disable trigger")),
		mcp.WithNumber("channel", mcp.Description("DIO line number")),
		mcp.WithNumber("position", mcp.Description("Prefill size")),
		mcp.WithNumber("timeout", mcp.Description("Auto-trigger timeout in seconds")),
		mcp.WithBoolean("rising_edge", mcp.Description("Rising (true) or falling (false) edge")),
		mcp.WithNumber("length_min", mcp.Description("Min trigger sequence duration in seconds")),
		mcp.WithNumber("length_max", mcp.Description("Max trigger sequence duration in seconds")),
		mcp.WithNumber("count", mcp.Description("Trigger event count")),
	), s.handleLogicTrigger)

	s.mcpServer.AddTool(mcp.NewTool("discovery_logic_record",
		mcp.WithDescription("Record digital signal from a DIO channel"),
		mcp.WithNumber("channel", mcp.Description("DIO line number"), mcp.Required()),
	), s.handleLogicRecord)

	s.mcpServer.AddTool(mcp.NewTool("discovery_logic_close",
		mcp.WithDescription("Reset the logic analyzer"),
	), s.handleLogicClose)

	// ---- Pattern Generator ----
	s.mcpServer.AddTool(mcp.NewTool("discovery_pattern_generate",
		mcp.WithDescription("Generate a digital pattern"),
		mcp.WithNumber("channel", mcp.Description("DIO line number"), mcp.Required()),
		mcp.WithNumber("function", mcp.Description("Type: 0=pulse, 1=custom, 2=random"), mcp.Required()),
		mcp.WithNumber("frequency", mcp.Description("Frequency in Hz"), mcp.Required()),
		mcp.WithNumber("duty_cycle", mcp.Description("Duty cycle % (for pulse)")),
		mcp.WithNumber("wait", mcp.Description("Wait time in seconds")),
		mcp.WithNumber("repeat", mcp.Description("Repeat count (0 = infinite)")),
		mcp.WithNumber("run_time", mcp.Description("Run time in seconds (0=infinite, -1=auto)")),
	), s.handlePatternGenerate)

	s.mcpServer.AddTool(mcp.NewTool("discovery_pattern_enable",
		mcp.WithDescription("Enable a digital output channel"),
		mcp.WithNumber("channel", mcp.Description("DIO line number"), mcp.Required()),
	), s.handlePatternEnable)

	s.mcpServer.AddTool(mcp.NewTool("discovery_pattern_disable",
		mcp.WithDescription("Disable a digital output channel"),
		mcp.WithNumber("channel", mcp.Description("DIO line number"), mcp.Required()),
	), s.handlePatternDisable)

	s.mcpServer.AddTool(mcp.NewTool("discovery_pattern_close",
		mcp.WithDescription("Reset the pattern generator"),
	), s.handlePatternClose)

	// ---- Static I/O ----
	s.mcpServer.AddTool(mcp.NewTool("discovery_static_set_mode",
		mcp.WithDescription("Set a DIO line as input or output"),
		mcp.WithNumber("channel", mcp.Description("DIO channel number"), mcp.Required()),
		mcp.WithBoolean("output", mcp.Description("true=output, false=input"), mcp.Required()),
	), s.handleStaticSetMode)

	s.mcpServer.AddTool(mcp.NewTool("discovery_static_get_state",
		mcp.WithDescription("Read the state of a DIO line"),
		mcp.WithNumber("channel", mcp.Description("DIO channel number"), mcp.Required()),
	), s.handleStaticGetState)

	s.mcpServer.AddTool(mcp.NewTool("discovery_static_set_state",
		mcp.WithDescription("Set a DIO line HIGH or LOW"),
		mcp.WithNumber("channel", mcp.Description("DIO channel number"), mcp.Required()),
		mcp.WithBoolean("value", mcp.Description("true=HIGH, false=LOW"), mcp.Required()),
	), s.handleStaticSetState)

	s.mcpServer.AddTool(mcp.NewTool("discovery_static_close",
		mcp.WithDescription("Reset the static I/O"),
	), s.handleStaticClose)

	// ---- UART ----
	s.mcpServer.AddTool(mcp.NewTool("discovery_uart_open",
		mcp.WithDescription("Initialize UART communication"),
		mcp.WithNumber("rx", mcp.Description("DIO line for RX"), mcp.Required()),
		mcp.WithNumber("tx", mcp.Description("DIO line for TX"), mcp.Required()),
		mcp.WithNumber("baud_rate", mcp.Description("Baud rate (default 9600)")),
		mcp.WithNumber("parity", mcp.Description("Parity: 0=none, 1=odd, 2=even")),
		mcp.WithNumber("data_bits", mcp.Description("Data bits (default 8)")),
		mcp.WithNumber("stop_bits", mcp.Description("Stop bits (default 1)")),
	), s.handleUARTOpen)

	s.mcpServer.AddTool(mcp.NewTool("discovery_uart_read",
		mcp.WithDescription("Read data from UART"),
	), s.handleUARTRead)

	s.mcpServer.AddTool(mcp.NewTool("discovery_uart_write",
		mcp.WithDescription("Write data through UART"),
		mcp.WithString("data", mcp.Description("Data to send (as string)"), mcp.Required()),
	), s.handleUARTWrite)

	s.mcpServer.AddTool(mcp.NewTool("discovery_uart_close",
		mcp.WithDescription("Reset the UART interface"),
	), s.handleUARTClose)

	// ---- SPI ----
	s.mcpServer.AddTool(mcp.NewTool("discovery_spi_open",
		mcp.WithDescription("Initialize SPI communication"),
		mcp.WithNumber("cs", mcp.Description("DIO line for chip select"), mcp.Required()),
		mcp.WithNumber("sck", mcp.Description("DIO line for serial clock"), mcp.Required()),
		mcp.WithNumber("miso", mcp.Description("DIO line for MISO (-1 to skip)")),
		mcp.WithNumber("mosi", mcp.Description("DIO line for MOSI (-1 to skip)")),
		mcp.WithNumber("clock_frequency", mcp.Description("Clock frequency in Hz (default 1MHz)")),
		mcp.WithNumber("mode", mcp.Description("SPI mode 0-3")),
		mcp.WithBoolean("msb_first", mcp.Description("MSB first (true) or LSB first (false)")),
	), s.handleSPIOpen)

	s.mcpServer.AddTool(mcp.NewTool("discovery_spi_read",
		mcp.WithDescription("Read data from SPI"),
		mcp.WithNumber("count", mcp.Description("Number of bytes to read"), mcp.Required()),
		mcp.WithNumber("cs", mcp.Description("Chip select line"), mcp.Required()),
	), s.handleSPIRead)

	s.mcpServer.AddTool(mcp.NewTool("discovery_spi_write",
		mcp.WithDescription("Write data through SPI"),
		mcp.WithString("data", mcp.Description("Data to send (hex string, e.g. 'FF01A2')"), mcp.Required()),
		mcp.WithNumber("cs", mcp.Description("Chip select line"), mcp.Required()),
	), s.handleSPIWrite)

	s.mcpServer.AddTool(mcp.NewTool("discovery_spi_close",
		mcp.WithDescription("Reset the SPI interface"),
	), s.handleSPIClose)

	// ---- I2C ----
	s.mcpServer.AddTool(mcp.NewTool("discovery_i2c_open",
		mcp.WithDescription("Initialize I2C communication"),
		mcp.WithNumber("sda", mcp.Description("DIO line for SDA"), mcp.Required()),
		mcp.WithNumber("scl", mcp.Description("DIO line for SCL"), mcp.Required()),
		mcp.WithNumber("clock_rate", mcp.Description("Clock rate in Hz (default 100kHz)")),
		mcp.WithBoolean("stretching", mcp.Description("Enable clock stretching")),
	), s.handleI2COpen)

	s.mcpServer.AddTool(mcp.NewTool("discovery_i2c_scan",
		mcp.WithDescription("Scan the I2C bus for connected devices (probes addresses 0x08-0x77)"),
	), s.handleI2CScan)

	s.mcpServer.AddTool(mcp.NewTool("discovery_i2c_read",
		mcp.WithDescription("Read data from I2C"),
		mcp.WithNumber("count", mcp.Description("Number of bytes to read"), mcp.Required()),
		mcp.WithNumber("address", mcp.Description("7-bit I2C address"), mcp.Required()),
	), s.handleI2CRead)

	s.mcpServer.AddTool(mcp.NewTool("discovery_i2c_write",
		mcp.WithDescription("Write data to I2C"),
		mcp.WithString("data", mcp.Description("Data to send (hex string, e.g. 'FF01A2')"), mcp.Required()),
		mcp.WithNumber("address", mcp.Description("7-bit I2C address"), mcp.Required()),
	), s.handleI2CWrite)

	s.mcpServer.AddTool(mcp.NewTool("discovery_i2c_close",
		mcp.WithDescription("Reset the I2C interface"),
	), s.handleI2CClose)
}
