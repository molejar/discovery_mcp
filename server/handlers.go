package server

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/molejar/discovery-mcp/dwf"
)

// Helper functions for parameter extraction

func argsMap(args any) map[string]interface{} {
	if m, ok := args.(map[string]interface{}); ok {
		return m
	}
	return map[string]interface{}{}
}

func getFloat(args any, key string, def float64) float64 {
	if v, ok := argsMap(args)[key]; ok {
		if f, ok := v.(float64); ok {
			return f
		}
	}
	return def
}

func getInt(args any, key string, def int) int {
	return int(getFloat(args, key, float64(def)))
}

func getBool(args any, key string, def bool) bool {
	if v, ok := argsMap(args)[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return def
}

func getString(args any, key, def string) string {
	if v, ok := argsMap(args)[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return def
}

func jsonResult(v interface{}) *mcp.CallToolResult {
	data, _ := json.Marshal(v)
	return mcp.NewToolResultText(string(data))
}

func errResult(err error) *mcp.CallToolResult {
	return mcp.NewToolResultError(err.Error())
}

// ==================== Device Handlers ====================

func (s *DiscoveryMCPServer) handleDeviceOpen(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	device := getString(req.Params.Arguments, "device", "")
	config := getInt(req.Params.Arguments, "config", 0)

	info, err := s.device.Open(device, config)
	if err != nil {
		return errResult(err), nil
	}
	return jsonResult(info), nil
}

func (s *DiscoveryMCPServer) handleDeviceClose(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := s.device.Close(); err != nil {
		return errResult(err), nil
	}
	return mcp.NewToolResultText("Device closed"), nil
}

func (s *DiscoveryMCPServer) handleDeviceTemperature(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	temp, err := s.device.Temperature()
	if err != nil {
		return errResult(err), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("%.2f °C", temp)), nil
}

// ==================== Oscilloscope Handlers ====================

func (s *DiscoveryMCPServer) handleScopeOpen(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cfg := dwf.ScopeConfig{
		SamplingFrequency: getFloat(req.Params.Arguments, "sampling_frequency", 20e6),
		BufferSize:        getInt(req.Params.Arguments, "buffer_size", 0),
		OffsetVoltage:     getFloat(req.Params.Arguments, "offset_voltage", 0),
		AmplitudeRange:    getFloat(req.Params.Arguments, "amplitude_range", 5),
	}
	if err := s.device.Scope().Open(cfg); err != nil {
		return errResult(err), nil
	}
	return mcp.NewToolResultText("Oscilloscope initialized"), nil
}

func (s *DiscoveryMCPServer) handleScopeMeasure(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ch := getInt(req.Params.Arguments, "channel", 1)
	voltage, err := s.device.Scope().Measure(ch)
	if err != nil {
		return errResult(err), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("%.6f V", voltage)), nil
}

func (s *DiscoveryMCPServer) handleScopeTrigger(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cfg := dwf.TriggerConfig{
		Enable:     getBool(req.Params.Arguments, "enable", true),
		Source:     dwf.TriggerSource(getInt(req.Params.Arguments, "source", 0)),
		Channel:    getInt(req.Params.Arguments, "channel", 1),
		Timeout:    getFloat(req.Params.Arguments, "timeout", 0),
		EdgeRising: getBool(req.Params.Arguments, "edge_rising", true),
		Level:      getFloat(req.Params.Arguments, "level", 0),
	}
	if err := s.device.Scope().SetTrigger(cfg); err != nil {
		return errResult(err), nil
	}
	return mcp.NewToolResultText("Trigger configured"), nil
}

func (s *DiscoveryMCPServer) handleScopeRecord(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ch := getInt(req.Params.Arguments, "channel", 1)
	data, err := s.device.Scope().Record(ch)
	if err != nil {
		return errResult(err), nil
	}
	return jsonResult(map[string]interface{}{
		"channel": ch,
		"samples": len(data),
		"data":    data,
	}), nil
}

func (s *DiscoveryMCPServer) handleScopeClose(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := s.device.Scope().Close(); err != nil {
		return errResult(err), nil
	}
	return mcp.NewToolResultText("Oscilloscope reset"), nil
}

// ==================== Wavegen Handlers ====================

func (s *DiscoveryMCPServer) handleWavegenGenerate(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cfg := dwf.WavegenConfig{
		Channel:   getInt(req.Params.Arguments, "channel", 1),
		Function:  dwf.WavegenFunc(getInt(req.Params.Arguments, "function", 1)),
		Offset:    getFloat(req.Params.Arguments, "offset", 0),
		Frequency: getFloat(req.Params.Arguments, "frequency", 1000),
		Amplitude: getFloat(req.Params.Arguments, "amplitude", 1),
		Symmetry:  getFloat(req.Params.Arguments, "symmetry", 50),
		Wait:      getFloat(req.Params.Arguments, "wait", 0),
		RunTime:   getFloat(req.Params.Arguments, "run_time", 0),
		Repeat:    getInt(req.Params.Arguments, "repeat", 0),
	}
	if err := s.device.Wavegen().Generate(cfg); err != nil {
		return errResult(err), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("Generating waveform on channel %d", cfg.Channel)), nil
}

func (s *DiscoveryMCPServer) handleWavegenEnable(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ch := getInt(req.Params.Arguments, "channel", 1)
	if err := s.device.Wavegen().Enable(ch); err != nil {
		return errResult(err), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("Wavegen channel %d enabled", ch)), nil
}

func (s *DiscoveryMCPServer) handleWavegenDisable(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ch := getInt(req.Params.Arguments, "channel", 1)
	if err := s.device.Wavegen().Disable(ch); err != nil {
		return errResult(err), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("Wavegen channel %d disabled", ch)), nil
}

func (s *DiscoveryMCPServer) handleWavegenClose(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ch := getInt(req.Params.Arguments, "channel", 1)
	if err := s.device.Wavegen().Close(ch); err != nil {
		return errResult(err), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("Wavegen channel %d reset", ch)), nil
}

// ==================== Power Supply Handlers ====================

func (s *DiscoveryMCPServer) handleSuppliesSwitch(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cfg := dwf.SuppliesConfig{
		MasterState:     getBool(req.Params.Arguments, "master_state", false),
		PositiveState:   getBool(req.Params.Arguments, "positive_state", false),
		NegativeState:   getBool(req.Params.Arguments, "negative_state", false),
		State:           getBool(req.Params.Arguments, "state", false),
		PositiveVoltage: getFloat(req.Params.Arguments, "positive_voltage", 0),
		NegativeVoltage: getFloat(req.Params.Arguments, "negative_voltage", 0),
		Voltage:         getFloat(req.Params.Arguments, "voltage", 0),
		PositiveCurrent: getFloat(req.Params.Arguments, "positive_current", 0),
		NegativeCurrent: getFloat(req.Params.Arguments, "negative_current", 0),
		Current:         getFloat(req.Params.Arguments, "current", 0),
	}
	if err := s.device.Supply().Switch(cfg); err != nil {
		return errResult(err), nil
	}
	return mcp.NewToolResultText("Power supplies configured"), nil
}

func (s *DiscoveryMCPServer) handleSuppliesClose(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := s.device.Supply().Close(); err != nil {
		return errResult(err), nil
	}
	return mcp.NewToolResultText("Power supplies reset"), nil
}

// ==================== DMM Handlers ====================

func (s *DiscoveryMCPServer) handleDMMOpen(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := s.device.DMM().Open(); err != nil {
		return errResult(err), nil
	}
	return mcp.NewToolResultText("DMM initialized"), nil
}

func (s *DiscoveryMCPServer) handleDMMMeasure(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	mode := dwf.DMMMode(getInt(req.Params.Arguments, "mode", 1))
	range_ := getFloat(req.Params.Arguments, "range", 0)
	highZ := getBool(req.Params.Arguments, "high_impedance", false)

	value, err := s.device.DMM().Measure(mode, range_, highZ)
	if err != nil {
		return errResult(err), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("%.6f", value)), nil
}

func (s *DiscoveryMCPServer) handleDMMClose(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := s.device.DMM().Close(); err != nil {
		return errResult(err), nil
	}
	return mcp.NewToolResultText("DMM reset"), nil
}

// ==================== Logic Analyzer Handlers ====================

func (s *DiscoveryMCPServer) handleLogicOpen(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cfg := dwf.LogicConfig{
		SamplingFrequency: getFloat(req.Params.Arguments, "sampling_frequency", 100e6),
		BufferSize:        getInt(req.Params.Arguments, "buffer_size", 0),
	}
	if err := s.device.Logic().Open(cfg); err != nil {
		return errResult(err), nil
	}
	return mcp.NewToolResultText("Logic analyzer initialized"), nil
}

func (s *DiscoveryMCPServer) handleLogicTrigger(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cfg := dwf.LogicTriggerConfig{
		Enable:     getBool(req.Params.Arguments, "enable", true),
		Channel:    getInt(req.Params.Arguments, "channel", 0),
		Position:   getInt(req.Params.Arguments, "position", 0),
		Timeout:    getFloat(req.Params.Arguments, "timeout", 0),
		RisingEdge: getBool(req.Params.Arguments, "rising_edge", true),
		LengthMin:  getFloat(req.Params.Arguments, "length_min", 0),
		LengthMax:  getFloat(req.Params.Arguments, "length_max", 20),
		Count:      getInt(req.Params.Arguments, "count", 1),
	}
	if err := s.device.Logic().SetTrigger(cfg); err != nil {
		return errResult(err), nil
	}
	return mcp.NewToolResultText("Logic trigger configured"), nil
}

func (s *DiscoveryMCPServer) handleLogicRecord(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ch := getInt(req.Params.Arguments, "channel", 0)
	data, err := s.device.Logic().Record(ch)
	if err != nil {
		return errResult(err), nil
	}
	return jsonResult(map[string]interface{}{
		"channel": ch,
		"samples": len(data),
		"data":    data,
	}), nil
}

func (s *DiscoveryMCPServer) handleLogicClose(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := s.device.Logic().Close(); err != nil {
		return errResult(err), nil
	}
	return mcp.NewToolResultText("Logic analyzer reset"), nil
}

// ==================== Pattern Generator Handlers ====================

func (s *DiscoveryMCPServer) handlePatternGenerate(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cfg := dwf.PatternConfig{
		Channel:   getInt(req.Params.Arguments, "channel", 0),
		Function:  dwf.DigitalOutType(getInt(req.Params.Arguments, "function", 0)),
		Frequency: getFloat(req.Params.Arguments, "frequency", 1000),
		DutyCycle: getFloat(req.Params.Arguments, "duty_cycle", 50),
		Wait:      getFloat(req.Params.Arguments, "wait", 0),
		Repeat:    getInt(req.Params.Arguments, "repeat", 0),
		RunTime:   getInt(req.Params.Arguments, "run_time", 0),
	}
	if err := s.device.Pattern().Generate(cfg); err != nil {
		return errResult(err), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("Pattern generated on DIO %d", cfg.Channel)), nil
}

func (s *DiscoveryMCPServer) handlePatternEnable(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ch := getInt(req.Params.Arguments, "channel", 0)
	if err := s.device.Pattern().Enable(ch); err != nil {
		return errResult(err), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("Pattern DIO %d enabled", ch)), nil
}

func (s *DiscoveryMCPServer) handlePatternDisable(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ch := getInt(req.Params.Arguments, "channel", 0)
	if err := s.device.Pattern().Disable(ch); err != nil {
		return errResult(err), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("Pattern DIO %d disabled", ch)), nil
}

func (s *DiscoveryMCPServer) handlePatternClose(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := s.device.Pattern().Close(); err != nil {
		return errResult(err), nil
	}
	return mcp.NewToolResultText("Pattern generator reset"), nil
}

// ==================== Static I/O Handlers ====================

func (s *DiscoveryMCPServer) handleStaticSetMode(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ch := getInt(req.Params.Arguments, "channel", 0)
	output := getBool(req.Params.Arguments, "output", false)
	if err := s.device.Static().SetMode(ch, output); err != nil {
		return errResult(err), nil
	}
	mode := "input"
	if output {
		mode = "output"
	}
	return mcp.NewToolResultText(fmt.Sprintf("DIO %d set to %s", ch, mode)), nil
}

func (s *DiscoveryMCPServer) handleStaticGetState(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ch := getInt(req.Params.Arguments, "channel", 0)
	state, err := s.device.Static().GetState(ch)
	if err != nil {
		return errResult(err), nil
	}
	stateStr := "LOW"
	if state {
		stateStr = "HIGH"
	}
	return mcp.NewToolResultText(fmt.Sprintf("DIO %d: %s", ch, stateStr)), nil
}

func (s *DiscoveryMCPServer) handleStaticSetState(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ch := getInt(req.Params.Arguments, "channel", 0)
	value := getBool(req.Params.Arguments, "value", false)
	if err := s.device.Static().SetState(ch, value); err != nil {
		return errResult(err), nil
	}
	stateStr := "LOW"
	if value {
		stateStr = "HIGH"
	}
	return mcp.NewToolResultText(fmt.Sprintf("DIO %d set to %s", ch, stateStr)), nil
}

func (s *DiscoveryMCPServer) handleStaticClose(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := s.device.Static().Close(); err != nil {
		return errResult(err), nil
	}
	return mcp.NewToolResultText("Static I/O reset"), nil
}

// ==================== UART Handlers ====================

func (s *DiscoveryMCPServer) handleUARTOpen(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cfg := dwf.UARTConfig{
		RX:       getInt(req.Params.Arguments, "rx", 0),
		TX:       getInt(req.Params.Arguments, "tx", 1),
		BaudRate: getInt(req.Params.Arguments, "baud_rate", 9600),
		Parity:   getInt(req.Params.Arguments, "parity", 0),
		DataBits: getInt(req.Params.Arguments, "data_bits", 8),
		StopBits: getInt(req.Params.Arguments, "stop_bits", 1),
	}
	if err := s.device.UARTProtocol().Open(cfg); err != nil {
		return errResult(err), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("UART initialized: %d baud, RX=DIO%d, TX=DIO%d", cfg.BaudRate, cfg.RX, cfg.TX)), nil
}

func (s *DiscoveryMCPServer) handleUARTRead(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	data, err := s.device.UARTProtocol().Read()
	if err != nil {
		return errResult(err), nil
	}
	return jsonResult(map[string]interface{}{
		"bytes": len(data),
		"data":  fmt.Sprintf("%x", data),
		"text":  string(data),
	}), nil
}

func (s *DiscoveryMCPServer) handleUARTWrite(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	data := getString(req.Params.Arguments, "data", "")
	if err := s.device.UARTProtocol().Write([]byte(data)); err != nil {
		return errResult(err), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("Sent %d bytes via UART", len(data))), nil
}

func (s *DiscoveryMCPServer) handleUARTClose(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := s.device.UARTProtocol().Close(); err != nil {
		return errResult(err), nil
	}
	return mcp.NewToolResultText("UART reset"), nil
}

// ==================== SPI Handlers ====================

func (s *DiscoveryMCPServer) handleSPIOpen(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cfg := dwf.SPIConfig{
		CS:             getInt(req.Params.Arguments, "cs", 0),
		SCK:            getInt(req.Params.Arguments, "sck", 1),
		MISO:           getInt(req.Params.Arguments, "miso", -1),
		MOSI:           getInt(req.Params.Arguments, "mosi", -1),
		ClockFrequency: getFloat(req.Params.Arguments, "clock_frequency", 1e6),
		Mode:           getInt(req.Params.Arguments, "mode", 0),
		MSBFirst:       getBool(req.Params.Arguments, "msb_first", true),
	}
	if err := s.device.SPIProtocol().Open(cfg); err != nil {
		return errResult(err), nil
	}
	return mcp.NewToolResultText("SPI initialized"), nil
}

func (s *DiscoveryMCPServer) handleSPIRead(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	count := getInt(req.Params.Arguments, "count", 1)
	cs := getInt(req.Params.Arguments, "cs", 0)
	data, err := s.device.SPIProtocol().Read(count, cs)
	if err != nil {
		return errResult(err), nil
	}
	return jsonResult(map[string]interface{}{
		"bytes": len(data),
		"data":  fmt.Sprintf("%x", data),
	}), nil
}

func (s *DiscoveryMCPServer) handleSPIWrite(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dataHex := getString(req.Params.Arguments, "data", "")
	cs := getInt(req.Params.Arguments, "cs", 0)
	data, err := hex.DecodeString(dataHex)
	if err != nil {
		return errResult(fmt.Errorf("invalid hex data: %w", err)), nil
	}
	if err := s.device.SPIProtocol().Write(data, cs); err != nil {
		return errResult(err), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("Sent %d bytes via SPI", len(data))), nil
}

func (s *DiscoveryMCPServer) handleSPIClose(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := s.device.SPIProtocol().Close(); err != nil {
		return errResult(err), nil
	}
	return mcp.NewToolResultText("SPI reset"), nil
}

// ==================== I2C Handlers ====================

func (s *DiscoveryMCPServer) handleI2COpen(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cfg := dwf.I2CConfig{
		SDA:        getInt(req.Params.Arguments, "sda", 0),
		SCL:        getInt(req.Params.Arguments, "scl", 1),
		ClockRate:  getFloat(req.Params.Arguments, "clock_rate", 100e3),
		Stretching: getBool(req.Params.Arguments, "stretching", false),
	}
	if err := s.device.I2CProtocol().Open(cfg); err != nil {
		return errResult(err), nil
	}
	return mcp.NewToolResultText("I2C initialized"), nil
}

func (s *DiscoveryMCPServer) handleI2CRead(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	count := getInt(req.Params.Arguments, "count", 1)
	addr := getInt(req.Params.Arguments, "address", 0)
	data, err := s.device.I2CProtocol().Read(count, addr)
	if err != nil {
		return errResult(err), nil
	}
	return jsonResult(map[string]interface{}{
		"address": fmt.Sprintf("0x%02X", addr),
		"bytes":   len(data),
		"data":    fmt.Sprintf("%x", data),
	}), nil
}

func (s *DiscoveryMCPServer) handleI2CWrite(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dataHex := getString(req.Params.Arguments, "data", "")
	addr := getInt(req.Params.Arguments, "address", 0)
	data, err := hex.DecodeString(dataHex)
	if err != nil {
		return errResult(fmt.Errorf("invalid hex data: %w", err)), nil
	}
	if err := s.device.I2CProtocol().Write(data, addr); err != nil {
		return errResult(err), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("Sent %d bytes to I2C 0x%02X", len(data), addr)), nil
}

func (s *DiscoveryMCPServer) handleI2CClose(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := s.device.I2CProtocol().Close(); err != nil {
		return errResult(err), nil
	}
	return mcp.NewToolResultText("I2C reset"), nil
}
