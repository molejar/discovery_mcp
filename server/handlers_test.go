package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/molejar/discovery-mcp/dwf"
)

// ============================= Mocks =============================

// mockScope implements dwf.Oscilloscope for testing.
type mockScope struct {
	openCfg    dwf.ScopeConfig
	openErr    error
	measureVal float64
	measureErr error
	triggerCfg dwf.TriggerConfig
	triggerErr error
	recordData []float64
	recordErr  error
	closeErr   error
}

func (m *mockScope) Open(cfg dwf.ScopeConfig) error {
	m.openCfg = cfg
	return m.openErr
}
func (m *mockScope) Measure(channel int) (float64, error) { return m.measureVal, m.measureErr }
func (m *mockScope) SetTrigger(cfg dwf.TriggerConfig) error {
	m.triggerCfg = cfg
	return m.triggerErr
}
func (m *mockScope) Record(channel int) ([]float64, error) { return m.recordData, m.recordErr }
func (m *mockScope) Close() error                          { return m.closeErr }

// mockWavegen implements dwf.WavegenDriver for testing.
type mockWavegen struct {
	generateCfg dwf.WavegenConfig
	generateErr error
	enableErr   error
	disableErr  error
	closeErr    error
}

func (m *mockWavegen) Generate(cfg dwf.WavegenConfig) error {
	m.generateCfg = cfg
	return m.generateErr
}
func (m *mockWavegen) Enable(channel int) error  { return m.enableErr }
func (m *mockWavegen) Disable(channel int) error { return m.disableErr }
func (m *mockWavegen) Close(channel int) error   { return m.closeErr }

// mockSupply implements dwf.PowerSupply for testing.
type mockSupply struct {
	switchCfg dwf.SuppliesConfig
	switchErr error
	closeErr  error
}

func (m *mockSupply) Switch(cfg dwf.SuppliesConfig) error {
	m.switchCfg = cfg
	return m.switchErr
}
func (m *mockSupply) Close() error { return m.closeErr }

// mockDMM implements dwf.DigitalMultimeter for testing.
type mockDMM struct {
	openErr    error
	measureVal float64
	measureErr error
	closeErr   error
}

func (m *mockDMM) Open() error { return m.openErr }
func (m *mockDMM) Measure(mode dwf.DMMMode, range_ float64, highImpedance bool) (float64, error) {
	return m.measureVal, m.measureErr
}
func (m *mockDMM) Close() error { return m.closeErr }

// mockLogic implements dwf.LogicAnalyzer for testing.
type mockLogic struct {
	openCfg    dwf.LogicConfig
	openErr    error
	triggerCfg dwf.LogicTriggerConfig
	triggerErr error
	recordData []uint16
	recordErr  error
	closeErr   error
}

func (m *mockLogic) Open(cfg dwf.LogicConfig) error {
	m.openCfg = cfg
	return m.openErr
}
func (m *mockLogic) SetTrigger(cfg dwf.LogicTriggerConfig) error {
	m.triggerCfg = cfg
	return m.triggerErr
}
func (m *mockLogic) Record(channel int) ([]uint16, error) { return m.recordData, m.recordErr }
func (m *mockLogic) Close() error                         { return m.closeErr }

// mockPattern implements dwf.PatternGenerator for testing.
type mockPattern struct {
	generateCfg dwf.PatternConfig
	generateErr error
	enableErr   error
	disableErr  error
	closeErr    error
}

func (m *mockPattern) Generate(cfg dwf.PatternConfig) error {
	m.generateCfg = cfg
	return m.generateErr
}
func (m *mockPattern) Enable(channel int) error  { return m.enableErr }
func (m *mockPattern) Disable(channel int) error { return m.disableErr }
func (m *mockPattern) Close() error              { return m.closeErr }

// mockStaticIO implements dwf.StaticIO for testing.
type mockStaticIO struct {
	setModeErr    error
	getStateVal   bool
	getStateErr   error
	setStateErr   error
	setCurrentErr error
	setPullErr    error
	closeErr      error
}

func (m *mockStaticIO) SetMode(channel int, output bool) error { return m.setModeErr }
func (m *mockStaticIO) GetState(channel int) (bool, error)     { return m.getStateVal, m.getStateErr }
func (m *mockStaticIO) SetState(channel int, value bool) error { return m.setStateErr }
func (m *mockStaticIO) SetCurrent(current float64) error       { return m.setCurrentErr }
func (m *mockStaticIO) SetPull(channel int, direction dwf.PullDirection) error {
	return m.setPullErr
}
func (m *mockStaticIO) Close() error { return m.closeErr }

// mockUART implements dwf.UART for testing.
type mockUART struct {
	openCfg  dwf.UARTConfig
	openErr  error
	readData []byte
	readErr  error
	writeErr error
	closeErr error
}

func (m *mockUART) Open(cfg dwf.UARTConfig) error {
	m.openCfg = cfg
	return m.openErr
}
func (m *mockUART) Read() ([]byte, error)   { return m.readData, m.readErr }
func (m *mockUART) Write(data []byte) error { return m.writeErr }
func (m *mockUART) Close() error            { return m.closeErr }

// mockSPI implements dwf.SPI for testing.
type mockSPI struct {
	openCfg      dwf.SPIConfig
	openErr      error
	readData     []byte
	readErr      error
	writeErr     error
	exchangeData []byte
	exchangeErr  error
	closeErr     error
}

func (m *mockSPI) Open(cfg dwf.SPIConfig) error           { m.openCfg = cfg; return m.openErr }
func (m *mockSPI) Read(count int, cs int) ([]byte, error) { return m.readData, m.readErr }
func (m *mockSPI) Write(data []byte, cs int) error        { return m.writeErr }
func (m *mockSPI) Exchange(txData []byte, rxCount int, cs int) ([]byte, error) {
	return m.exchangeData, m.exchangeErr
}
func (m *mockSPI) Close() error { return m.closeErr }

// mockI2C implements dwf.I2C for testing.
type mockI2C struct {
	openCfg      dwf.I2CConfig
	openErr      error
	scanData     []int
	scanErr      error
	readData     []byte
	readErr      error
	writeErr     error
	exchangeData []byte
	exchangeErr  error
	closeErr     error
}

func (m *mockI2C) Open(cfg dwf.I2CConfig) error                { m.openCfg = cfg; return m.openErr }
func (m *mockI2C) Scan() ([]int, error)                        { return m.scanData, m.scanErr }
func (m *mockI2C) Read(count int, address int) ([]byte, error) { return m.readData, m.readErr }
func (m *mockI2C) Write(data []byte, address int) error        { return m.writeErr }
func (m *mockI2C) Exchange(txData []byte, rxCount int, address int) ([]byte, error) {
	return m.exchangeData, m.exchangeErr
}
func (m *mockI2C) Close() error { return m.closeErr }

// mockDevice implements dwf.DiscoveryDevice, aggregating all mock instruments.
type mockDevice struct {
	enumDevices    []dwf.EnumDevice
	enumDevicesErr error
	enumConfigs    []dwf.DeviceConfig
	enumConfigsErr error
	openInfo       *dwf.DeviceInfo
	openErr        error
	closeErr       error
	temperature    float64
	tempErr        error
	scope          *mockScope
	wavegen        *mockWavegen
	supply         *mockSupply
	dmm            *mockDMM
	logic          *mockLogic
	pattern        *mockPattern
	staticIO       *mockStaticIO
	uart           *mockUART
	spi            *mockSPI
	i2c            *mockI2C
}

func (d *mockDevice) EnumDevices() ([]dwf.EnumDevice, error) {
	return d.enumDevices, d.enumDevicesErr
}
func (d *mockDevice) EnumConfigs(deviceIndex int) ([]dwf.DeviceConfig, error) {
	return d.enumConfigs, d.enumConfigsErr
}
func (d *mockDevice) Open(device string, config int) (*dwf.DeviceInfo, error) {
	return d.openInfo, d.openErr
}
func (d *mockDevice) Close() error                  { return d.closeErr }
func (d *mockDevice) Temperature() (float64, error) { return d.temperature, d.tempErr }
func (d *mockDevice) Scope() dwf.Oscilloscope       { return d.scope }
func (d *mockDevice) Wavegen() dwf.WavegenDriver    { return d.wavegen }
func (d *mockDevice) Supply() dwf.PowerSupply       { return d.supply }
func (d *mockDevice) DMM() dwf.DigitalMultimeter    { return d.dmm }
func (d *mockDevice) Logic() dwf.LogicAnalyzer      { return d.logic }
func (d *mockDevice) Pattern() dwf.PatternGenerator { return d.pattern }
func (d *mockDevice) Static() dwf.StaticIO          { return d.staticIO }
func (d *mockDevice) UARTProtocol() dwf.UART        { return d.uart }
func (d *mockDevice) SPIProtocol() dwf.SPI          { return d.spi }
func (d *mockDevice) I2CProtocol() dwf.I2C          { return d.i2c }

// newTestServer creates a DiscoveryMCPServer with a fully mocked device.
func newTestServer() (*DiscoveryMCPServer, *mockDevice) {
	dev := &mockDevice{
		scope:    &mockScope{},
		wavegen:  &mockWavegen{},
		supply:   &mockSupply{},
		dmm:      &mockDMM{},
		logic:    &mockLogic{},
		pattern:  &mockPattern{},
		staticIO: &mockStaticIO{},
		uart:     &mockUART{},
		spi:      &mockSPI{},
		i2c:      &mockI2C{},
	}
	s := NewWithDevice(dev)
	return s, dev
}

// makeReq builds a CallToolRequest with the given argument map.
func makeReq(args map[string]interface{}) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: args,
		},
	}
}

// ============================= Helper Tests =============================

func TestArgsMap(t *testing.T) {
	t.Run("valid map", func(t *testing.T) {
		input := map[string]interface{}{"key": "value"}
		result := argsMap(input)
		if result["key"] != "value" {
			t.Errorf("expected 'value', got %v", result["key"])
		}
	})

	t.Run("nil input", func(t *testing.T) {
		result := argsMap(nil)
		if len(result) != 0 {
			t.Errorf("expected empty map, got %v", result)
		}
	})

	t.Run("wrong type", func(t *testing.T) {
		result := argsMap("not a map")
		if len(result) != 0 {
			t.Errorf("expected empty map, got %v", result)
		}
	})
}

func TestGetFloat(t *testing.T) {
	args := map[string]interface{}{"voltage": 3.3, "label": "test"}

	if v := getFloat(args, "voltage", 0); v != 3.3 {
		t.Errorf("expected 3.3, got %f", v)
	}
	if v := getFloat(args, "missing", 5.0); v != 5.0 {
		t.Errorf("expected default 5.0, got %f", v)
	}
	if v := getFloat(args, "label", 1.0); v != 1.0 {
		t.Errorf("expected default 1.0 for wrong type, got %f", v)
	}
}

func TestGetInt(t *testing.T) {
	args := map[string]interface{}{"channel": float64(2)}

	if v := getInt(args, "channel", 0); v != 2 {
		t.Errorf("expected 2, got %d", v)
	}
	if v := getInt(args, "missing", 7); v != 7 {
		t.Errorf("expected default 7, got %d", v)
	}
}

func TestGetBool(t *testing.T) {
	args := map[string]interface{}{"enable": true, "count": float64(1)}

	if v := getBool(args, "enable", false); v != true {
		t.Error("expected true")
	}
	if v := getBool(args, "missing", true); v != true {
		t.Error("expected default true")
	}
	if v := getBool(args, "count", false); v != false {
		t.Error("expected default false for wrong type")
	}
}

func TestGetString(t *testing.T) {
	args := map[string]interface{}{"name": "AD2", "num": float64(42)}

	if v := getString(args, "name", ""); v != "AD2" {
		t.Errorf("expected 'AD2', got %q", v)
	}
	if v := getString(args, "missing", "default"); v != "default" {
		t.Errorf("expected 'default', got %q", v)
	}
	if v := getString(args, "num", "fallback"); v != "fallback" {
		t.Errorf("expected 'fallback' for wrong type, got %q", v)
	}
}

func TestJsonResult(t *testing.T) {
	result := jsonResult(map[string]int{"a": 1})
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	// Verify it contains valid JSON
	text := result.Content[0].(mcp.TextContent).Text
	var parsed map[string]int
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if parsed["a"] != 1 {
		t.Errorf("expected a=1, got %v", parsed["a"])
	}
}

func TestErrResult(t *testing.T) {
	result := errResult(errors.New("test error"))
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !result.IsError {
		t.Error("expected IsError = true")
	}
	text := result.Content[0].(mcp.TextContent).Text
	if text != "test error" {
		t.Errorf("expected 'test error', got %q", text)
	}
}

// ============================= Server Construction =============================

func TestNewWithDevice(t *testing.T) {
	s, _ := newTestServer()
	if s.MCPServer() == nil {
		t.Fatal("MCPServer() should not be nil")
	}
	if s.DeviceInstance() == nil {
		t.Fatal("DeviceInstance() should not be nil")
	}
}

// ============================= Device Handlers =============================

func TestHandleEnumerate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		s, dev := newTestServer()
		dev.enumDevices = []dwf.EnumDevice{
			{Index: 0, DeviceName: "Analog Discovery 2", SerialNumber: "SN123", IsOpened: false},
			{Index: 1, DeviceName: "Digital Discovery", SerialNumber: "SN456", IsOpened: true},
		}
		result, err := s.handleEnumerate(context.Background(), makeReq(nil))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		text := result.Content[0].(mcp.TextContent).Text
		if !strings.Contains(text, "Analog Discovery 2") {
			t.Errorf("expected device name in result, got %q", text)
		}
		if !strings.Contains(text, "Digital Discovery") {
			t.Errorf("expected second device in result, got %q", text)
		}
	})

	t.Run("no devices", func(t *testing.T) {
		s, _ := newTestServer()
		result, err := s.handleEnumerate(context.Background(), makeReq(nil))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		text := result.Content[0].(mcp.TextContent).Text
		if text != "null" {
			t.Errorf("expected 'null' for empty list, got %q", text)
		}
	})

	t.Run("error", func(t *testing.T) {
		s, dev := newTestServer()
		dev.enumDevicesErr = errors.New("enum failed")
		result, err := s.handleEnumerate(context.Background(), makeReq(nil))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Error("expected error result")
		}
	})
}

func TestHandleDeviceGetConfigs(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		s, dev := newTestServer()
		dev.enumConfigs = []dwf.DeviceConfig{
			{AnalogInChannels: 2, DigitalInChannels: 16},
			{AnalogInChannels: 4, DigitalInChannels: 8},
		}
		result, err := s.handleDeviceGetConfigs(context.Background(), makeReq(map[string]any{"device_index": 0}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		text := result.Content[0].(mcp.TextContent).Text
		if !strings.Contains(text, `"AnalogInChannels":2`) {
			t.Errorf("expected 2 analog channels, got %q", text)
		}
		if !strings.Contains(text, `"AnalogInChannels":4`) {
			t.Errorf("expected 4 analog channels, got %q", text)
		}
	})

	t.Run("error", func(t *testing.T) {
		s, dev := newTestServer()
		dev.enumConfigsErr = errors.New("enum config failed")
		result, err := s.handleDeviceGetConfigs(context.Background(), makeReq(map[string]any{"device_index": 0}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Error("expected error result")
		}
	})
}

func TestHandleDeviceOpen(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		s, dev := newTestServer()
		dev.openInfo = &dwf.DeviceInfo{Name: "Analog Discovery 2", Handle: 1}
		result, err := s.handleDeviceOpen(context.Background(), makeReq(map[string]interface{}{
			"device": "Analog Discovery 2",
			"config": float64(0),
		}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		text := result.Content[0].(mcp.TextContent).Text
		if !strings.Contains(text, "Analog Discovery 2") {
			t.Errorf("expected device name in result, got %q", text)
		}
	})

	t.Run("error", func(t *testing.T) {
		s, dev := newTestServer()
		dev.openErr = errors.New("no device")
		result, err := s.handleDeviceOpen(context.Background(), makeReq(nil))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Error("expected error result")
		}
	})
}

func TestHandleDeviceClose(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		s, _ := newTestServer()
		result, err := s.handleDeviceClose(context.Background(), makeReq(nil))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		text := result.Content[0].(mcp.TextContent).Text
		if !strings.Contains(text, "closed") {
			t.Errorf("expected 'closed' in result, got %q", text)
		}
	})

	t.Run("error", func(t *testing.T) {
		s, dev := newTestServer()
		dev.closeErr = errors.New("close fail")
		result, err := s.handleDeviceClose(context.Background(), makeReq(nil))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Error("expected error result")
		}
	})
}

func TestHandleDeviceTemperature(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		s, dev := newTestServer()
		dev.temperature = 42.5
		result, err := s.handleDeviceTemperature(context.Background(), makeReq(nil))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		text := result.Content[0].(mcp.TextContent).Text
		if !strings.Contains(text, "42.50") {
			t.Errorf("expected temperature in result, got %q", text)
		}
	})

	t.Run("error", func(t *testing.T) {
		s, dev := newTestServer()
		dev.tempErr = errors.New("sensor fail")
		result, err := s.handleDeviceTemperature(context.Background(), makeReq(nil))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Error("expected error result")
		}
	})
}

// ============================= Scope Handlers =============================

func TestHandleScopeOpen(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		s, _ := newTestServer()
		result, err := s.handleScopeOpen(context.Background(), makeReq(map[string]interface{}{
			"sampling_frequency": 10e6,
			"buffer_size":        float64(4096),
		}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		text := result.Content[0].(mcp.TextContent).Text
		if !strings.Contains(text, "initialized") {
			t.Errorf("expected 'initialized', got %q", text)
		}
	})

	t.Run("error", func(t *testing.T) {
		s, dev := newTestServer()
		dev.scope.openErr = errors.New("scope fail")
		result, _ := s.handleScopeOpen(context.Background(), makeReq(nil))
		if !result.IsError {
			t.Error("expected error result")
		}
	})
}

func TestHandleScopeMeasure(t *testing.T) {
	s, dev := newTestServer()
	dev.scope.measureVal = 1.234567
	result, err := s.handleScopeMeasure(context.Background(), makeReq(map[string]interface{}{
		"channel": float64(1),
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "1.234567") {
		t.Errorf("expected voltage value, got %q", text)
	}
}

func TestHandleScopeTrigger(t *testing.T) {
	s, _ := newTestServer()
	result, err := s.handleScopeTrigger(context.Background(), makeReq(map[string]interface{}{
		"enable":      true,
		"source":      float64(2),
		"channel":     float64(1),
		"edge_rising": true,
		"level":       1.5,
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "configured") {
		t.Errorf("expected 'configured', got %q", text)
	}
}

func TestHandleScopeRecord(t *testing.T) {
	s, dev := newTestServer()
	dev.scope.recordData = []float64{0.1, 0.2, 0.3}
	result, err := s.handleScopeRecord(context.Background(), makeReq(map[string]interface{}{
		"channel": float64(1),
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, `"samples":3`) {
		t.Errorf("expected samples count, got %q", text)
	}
}

func TestHandleScopeClose(t *testing.T) {
	s, _ := newTestServer()
	result, err := s.handleScopeClose(context.Background(), makeReq(nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "reset") {
		t.Errorf("expected 'reset', got %q", text)
	}
}

// ============================= Wavegen Handlers =============================

func TestHandleWavegenGenerate(t *testing.T) {
	s, _ := newTestServer()
	result, err := s.handleWavegenGenerate(context.Background(), makeReq(map[string]any{
		"channel":   float64(1),
		"function":  float64(1),
		"frequency": 1000.0,
		"amplitude": 2.5,
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "channel 1") {
		t.Errorf("expected 'channel 1', got %q", text)
	}
}

func TestHandleWavegenEnable(t *testing.T) {
	s, _ := newTestServer()
	result, err := s.handleWavegenEnable(context.Background(), makeReq(map[string]any{
		"channel": float64(2),
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "enabled") {
		t.Errorf("expected 'enabled', got %q", text)
	}
}

func TestHandleWavegenDisable(t *testing.T) {
	s, _ := newTestServer()
	result, err := s.handleWavegenDisable(context.Background(), makeReq(map[string]any{
		"channel": float64(1),
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "disabled") {
		t.Errorf("expected 'disabled', got %q", text)
	}
}

func TestHandleWavegenClose(t *testing.T) {
	s, _ := newTestServer()
	result, err := s.handleWavegenClose(context.Background(), makeReq(map[string]any{
		"channel": float64(1),
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "reset") {
		t.Errorf("expected 'reset', got %q", text)
	}
}

// ============================= Supply Handlers =============================

func TestHandleSuppliesSwitch(t *testing.T) {
	s, _ := newTestServer()
	result, err := s.handleSuppliesSwitch(context.Background(), makeReq(map[string]any{
		"master_state":     true,
		"positive_state":   true,
		"positive_voltage": 5.0,
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "configured") {
		t.Errorf("expected 'configured', got %q", text)
	}
}

func TestHandleSuppliesClose(t *testing.T) {
	s, _ := newTestServer()
	result, err := s.handleSuppliesClose(context.Background(), makeReq(nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "reset") {
		t.Errorf("expected 'reset', got %q", text)
	}
}

// ============================= DMM Handlers =============================

func TestHandleDMMOpen(t *testing.T) {
	s, _ := newTestServer()
	result, err := s.handleDMMOpen(context.Background(), makeReq(nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "initialized") {
		t.Errorf("expected 'initialized', got %q", text)
	}
}

func TestHandleDMMMeasure(t *testing.T) {
	s, dev := newTestServer()
	dev.dmm.measureVal = 12.345678
	result, err := s.handleDMMMeasure(context.Background(), makeReq(map[string]any{
		"mode":           float64(1),
		"range":          0.0,
		"high_impedance": true,
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "12.345678") {
		t.Errorf("expected measurement value, got %q", text)
	}
}

func TestHandleDMMClose(t *testing.T) {
	s, _ := newTestServer()
	result, err := s.handleDMMClose(context.Background(), makeReq(nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "reset") {
		t.Errorf("expected 'reset', got %q", text)
	}
}

// ============================= Logic Handlers =============================

func TestHandleLogicOpen(t *testing.T) {
	s, _ := newTestServer()
	result, err := s.handleLogicOpen(context.Background(), makeReq(map[string]any{
		"sampling_frequency": 100e6,
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "initialized") {
		t.Errorf("expected 'initialized', got %q", text)
	}
}

func TestHandleLogicTrigger(t *testing.T) {
	s, _ := newTestServer()
	result, err := s.handleLogicTrigger(context.Background(), makeReq(map[string]any{
		"enable":      true,
		"channel":     float64(0),
		"rising_edge": true,
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "configured") {
		t.Errorf("expected 'configured', got %q", text)
	}
}

func TestHandleLogicRecord(t *testing.T) {
	s, dev := newTestServer()
	dev.logic.recordData = []uint16{0, 1, 0, 1}
	result, err := s.handleLogicRecord(context.Background(), makeReq(map[string]any{
		"channel": float64(0),
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, `"samples":4`) {
		t.Errorf("expected samples count, got %q", text)
	}
}

func TestHandleLogicClose(t *testing.T) {
	s, _ := newTestServer()
	result, err := s.handleLogicClose(context.Background(), makeReq(nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "reset") {
		t.Errorf("expected 'reset', got %q", text)
	}
}

// ============================= Pattern Handlers =============================

func TestHandlePatternGenerate(t *testing.T) {
	s, _ := newTestServer()
	result, err := s.handlePatternGenerate(context.Background(), makeReq(map[string]any{
		"channel":   float64(0),
		"function":  float64(0),
		"frequency": 1000.0,
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "DIO 0") {
		t.Errorf("expected 'DIO 0', got %q", text)
	}
}

func TestHandlePatternEnable(t *testing.T) {
	s, _ := newTestServer()
	result, err := s.handlePatternEnable(context.Background(), makeReq(map[string]any{
		"channel": float64(3),
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "enabled") {
		t.Errorf("expected 'enabled', got %q", text)
	}
}

func TestHandlePatternDisable(t *testing.T) {
	s, _ := newTestServer()
	result, err := s.handlePatternDisable(context.Background(), makeReq(map[string]any{
		"channel": float64(3),
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "disabled") {
		t.Errorf("expected 'disabled', got %q", text)
	}
}

func TestHandlePatternClose(t *testing.T) {
	s, _ := newTestServer()
	result, err := s.handlePatternClose(context.Background(), makeReq(nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "reset") {
		t.Errorf("expected 'reset', got %q", text)
	}
}

// ============================= Static I/O Handlers =============================

func TestHandleStaticSetMode(t *testing.T) {
	t.Run("output", func(t *testing.T) {
		s, _ := newTestServer()
		result, err := s.handleStaticSetMode(context.Background(), makeReq(map[string]any{
			"channel": float64(0),
			"output":  true,
		}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		text := result.Content[0].(mcp.TextContent).Text
		if !strings.Contains(text, "output") {
			t.Errorf("expected 'output', got %q", text)
		}
	})

	t.Run("input", func(t *testing.T) {
		s, _ := newTestServer()
		result, err := s.handleStaticSetMode(context.Background(), makeReq(map[string]any{
			"channel": float64(0),
			"output":  false,
		}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		text := result.Content[0].(mcp.TextContent).Text
		if !strings.Contains(text, "input") {
			t.Errorf("expected 'input', got %q", text)
		}
	})
}

func TestHandleStaticGetState(t *testing.T) {
	t.Run("high", func(t *testing.T) {
		s, dev := newTestServer()
		dev.staticIO.getStateVal = true
		result, err := s.handleStaticGetState(context.Background(), makeReq(map[string]any{
			"channel": float64(0),
		}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		text := result.Content[0].(mcp.TextContent).Text
		if !strings.Contains(text, "HIGH") {
			t.Errorf("expected 'HIGH', got %q", text)
		}
	})

	t.Run("low", func(t *testing.T) {
		s, dev := newTestServer()
		dev.staticIO.getStateVal = false
		result, err := s.handleStaticGetState(context.Background(), makeReq(map[string]any{
			"channel": float64(0),
		}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		text := result.Content[0].(mcp.TextContent).Text
		if !strings.Contains(text, "LOW") {
			t.Errorf("expected 'LOW', got %q", text)
		}
	})
}

func TestHandleStaticSetState(t *testing.T) {
	s, _ := newTestServer()
	result, err := s.handleStaticSetState(context.Background(), makeReq(map[string]any{
		"channel": float64(5),
		"value":   true,
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "HIGH") {
		t.Errorf("expected 'HIGH', got %q", text)
	}
}

func TestHandleStaticClose(t *testing.T) {
	s, _ := newTestServer()
	result, err := s.handleStaticClose(context.Background(), makeReq(nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "reset") {
		t.Errorf("expected 'reset', got %q", text)
	}
}

// ============================= UART Handlers =============================

func TestHandleUARTOpen(t *testing.T) {
	s, _ := newTestServer()
	result, err := s.handleUARTOpen(context.Background(), makeReq(map[string]any{
		"rx":        float64(0),
		"tx":        float64(1),
		"baud_rate": float64(115200),
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "115200") {
		t.Errorf("expected baud rate in result, got %q", text)
	}
}

func TestHandleUARTRead(t *testing.T) {
	s, dev := newTestServer()
	dev.uart.readData = []byte("hello")
	result, err := s.handleUARTRead(context.Background(), makeReq(nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "hello") {
		t.Errorf("expected 'hello' in result, got %q", text)
	}
}

func TestHandleUARTWrite(t *testing.T) {
	s, _ := newTestServer()
	result, err := s.handleUARTWrite(context.Background(), makeReq(map[string]any{
		"data": "test",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "4 bytes") {
		t.Errorf("expected '4 bytes', got %q", text)
	}
}

func TestHandleUARTClose(t *testing.T) {
	s, _ := newTestServer()
	result, err := s.handleUARTClose(context.Background(), makeReq(nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "reset") {
		t.Errorf("expected 'reset', got %q", text)
	}
}

// ============================= SPI Handlers =============================

func TestHandleSPIOpen(t *testing.T) {
	s, _ := newTestServer()
	result, err := s.handleSPIOpen(context.Background(), makeReq(map[string]any{
		"cs":  float64(0),
		"sck": float64(1),
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "initialized") {
		t.Errorf("expected 'initialized', got %q", text)
	}
}

func TestHandleSPIRead(t *testing.T) {
	s, dev := newTestServer()
	dev.spi.readData = []byte{0xAB, 0xCD}
	result, err := s.handleSPIRead(context.Background(), makeReq(map[string]any{
		"count": float64(2),
		"cs":    float64(0),
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "abcd") {
		t.Errorf("expected hex data, got %q", text)
	}
}

func TestHandleSPIWrite(t *testing.T) {
	t.Run("valid hex", func(t *testing.T) {
		s, _ := newTestServer()
		result, err := s.handleSPIWrite(context.Background(), makeReq(map[string]any{
			"data": "FF01A2",
			"cs":   float64(0),
		}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		text := result.Content[0].(mcp.TextContent).Text
		if !strings.Contains(text, "3 bytes") {
			t.Errorf("expected '3 bytes', got %q", text)
		}
	})

	t.Run("invalid hex", func(t *testing.T) {
		s, _ := newTestServer()
		result, err := s.handleSPIWrite(context.Background(), makeReq(map[string]any{
			"data": "ZZZZ",
			"cs":   float64(0),
		}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Error("expected error result for invalid hex")
		}
	})
}

func TestHandleSPIClose(t *testing.T) {
	s, _ := newTestServer()
	result, err := s.handleSPIClose(context.Background(), makeReq(nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "reset") {
		t.Errorf("expected 'reset', got %q", text)
	}
}

// ============================= I2C Handlers =============================

func TestHandleI2COpen(t *testing.T) {
	s, _ := newTestServer()
	result, err := s.handleI2COpen(context.Background(), makeReq(map[string]any{
		"sda": float64(0),
		"scl": float64(1),
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "initialized") {
		t.Errorf("expected 'initialized', got %q", text)
	}
}

func TestHandleI2CScan(t *testing.T) {
	t.Run("found devices", func(t *testing.T) {
		s, dev := newTestServer()
		dev.i2c.scanData = []int{0x20, 0x50, 0x68}
		result, err := s.handleI2CScan(context.Background(), makeReq(nil))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		text := result.Content[0].(mcp.TextContent).Text
		if !strings.Contains(text, "0x20") {
			t.Errorf("expected '0x20' in result, got %q", text)
		}
		if !strings.Contains(text, "0x50") {
			t.Errorf("expected '0x50' in result, got %q", text)
		}
		if !strings.Contains(text, `"count":3`) {
			t.Errorf("expected count 3, got %q", text)
		}
	})

	t.Run("no devices", func(t *testing.T) {
		s, _ := newTestServer()
		result, err := s.handleI2CScan(context.Background(), makeReq(nil))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		text := result.Content[0].(mcp.TextContent).Text
		if !strings.Contains(text, `"count":0`) {
			t.Errorf("expected count 0, got %q", text)
		}
	})

	t.Run("error", func(t *testing.T) {
		s, dev := newTestServer()
		dev.i2c.scanErr = errors.New("scan failed")
		result, err := s.handleI2CScan(context.Background(), makeReq(nil))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Error("expected error result")
		}
	})
}

func TestHandleI2CRead(t *testing.T) {
	s, dev := newTestServer()
	dev.i2c.readData = []byte{0x42, 0x43}
	result, err := s.handleI2CRead(context.Background(), makeReq(map[string]any{
		"count":   float64(2),
		"address": float64(0x50),
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "0x50") {
		t.Errorf("expected address 0x50, got %q", text)
	}
}

func TestHandleI2CWrite(t *testing.T) {
	t.Run("valid hex", func(t *testing.T) {
		s, _ := newTestServer()
		result, err := s.handleI2CWrite(context.Background(), makeReq(map[string]any{
			"data":    "AABB",
			"address": float64(0x50),
		}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		text := result.Content[0].(mcp.TextContent).Text
		if !strings.Contains(text, "2 bytes") {
			t.Errorf("expected '2 bytes', got %q", text)
		}
	})

	t.Run("invalid hex", func(t *testing.T) {
		s, _ := newTestServer()
		result, err := s.handleI2CWrite(context.Background(), makeReq(map[string]any{
			"data":    "XXYY",
			"address": float64(0x50),
		}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Error("expected error result for invalid hex")
		}
	})
}

func TestHandleI2CClose(t *testing.T) {
	s, _ := newTestServer()
	result, err := s.handleI2CClose(context.Background(), makeReq(nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "reset") {
		t.Errorf("expected 'reset', got %q", text)
	}
}

// ============================= Error Propagation =============================

func TestHandlerErrorPropagation(t *testing.T) {
	// Verify that instrument errors propagate as tool errors (IsError=true),
	// not as Go function errors.
	tests := []struct {
		name    string
		setup   func(*mockDevice)
		handler func(*DiscoveryMCPServer) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)
	}{
		{
			name:  "scope measure error",
			setup: func(d *mockDevice) { d.scope.measureErr = errors.New("scope fail") },
			handler: func(s *DiscoveryMCPServer) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				return s.handleScopeMeasure
			},
		},
		{
			name:  "wavegen generate error",
			setup: func(d *mockDevice) { d.wavegen.generateErr = errors.New("wavegen fail") },
			handler: func(s *DiscoveryMCPServer) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				return s.handleWavegenGenerate
			},
		},
		{
			name:  "supply switch error",
			setup: func(d *mockDevice) { d.supply.switchErr = errors.New("supply fail") },
			handler: func(s *DiscoveryMCPServer) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				return s.handleSuppliesSwitch
			},
		},
		{
			name:  "dmm measure error",
			setup: func(d *mockDevice) { d.dmm.measureErr = errors.New("dmm fail") },
			handler: func(s *DiscoveryMCPServer) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				return s.handleDMMMeasure
			},
		},
		{
			name:  "logic record error",
			setup: func(d *mockDevice) { d.logic.recordErr = errors.New("logic fail") },
			handler: func(s *DiscoveryMCPServer) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				return s.handleLogicRecord
			},
		},
		{
			name:  "pattern generate error",
			setup: func(d *mockDevice) { d.pattern.generateErr = errors.New("pattern fail") },
			handler: func(s *DiscoveryMCPServer) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				return s.handlePatternGenerate
			},
		},
		{
			name:  "static get state error",
			setup: func(d *mockDevice) { d.staticIO.getStateErr = errors.New("static fail") },
			handler: func(s *DiscoveryMCPServer) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				return s.handleStaticGetState
			},
		},
		{
			name:  "uart read error",
			setup: func(d *mockDevice) { d.uart.readErr = errors.New("uart fail") },
			handler: func(s *DiscoveryMCPServer) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				return s.handleUARTRead
			},
		},
		{
			name:  "spi read error",
			setup: func(d *mockDevice) { d.spi.readErr = errors.New("spi fail") },
			handler: func(s *DiscoveryMCPServer) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				return s.handleSPIRead
			},
		},
		{
			name:  "i2c read error",
			setup: func(d *mockDevice) { d.i2c.readErr = errors.New("i2c fail") },
			handler: func(s *DiscoveryMCPServer) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				return s.handleI2CRead
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, dev := newTestServer()
			tt.setup(dev)
			handler := tt.handler(s)
			result, err := handler(context.Background(), makeReq(map[string]interface{}{
				"channel": float64(1),
				"mode":    float64(0),
				"count":   float64(1),
				"address": float64(0),
				"cs":      float64(0),
			}))
			if err != nil {
				t.Fatalf("handler returned Go error: %v", err)
			}
			if !result.IsError {
				t.Error("expected tool error (IsError=true)")
			}
			text := result.Content[0].(mcp.TextContent).Text
			if text == "" {
				t.Error("expected error message in result")
			}
		})
	}
}

// Ensure unused imports are referenced.
var _ = fmt.Sprintf
