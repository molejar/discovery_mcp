package dwf

import (
	"fmt"
)

// deviceNames maps human-readable names to DWF SDK device filter IDs.
var deviceNames = map[string]DevHandle{
	"Analog Discovery":          DevHandle(cDevidDiscovery),
	"Analog Discovery 2":        DevHandle(cDevidDiscovery2),
	"Analog Discovery Studio":   DevHandle(cDevidDiscovery2),
	"Digital Discovery":         DevHandle(cDevidDDiscovery),
	"Analog Discovery Pro 3X50": DevHandle(cDevidADP3X50),
	"Analog Discovery Pro 5250": DevHandle(cDevidADP5250),
}

// deviceIDToName maps device IDs back to names.
var deviceIDToName = map[int]string{
	int(cDevidDiscovery):  "Analog Discovery",
	int(cDevidDiscovery2): "Analog Discovery 2",
	int(cDevidDDiscovery): "Digital Discovery",
	int(cDevidADP3X50):    "Analog Discovery Pro 3X50",
	int(cDevidADP5250):    "Analog Discovery Pro 5250",
}

// Device is the concrete implementation of DiscoveryDevice.
// It holds the native device handle and provides access to all instruments.
type Device struct {
	handle DevHandle
	info   *DeviceInfo

	scope    *scopeImpl
	wavegen  *wavegenImpl
	supply   *supplyImpl
	dmm      *dmmImpl
	logic    *logicImpl
	pattern  *patternImpl
	staticIO *staticIOImpl
	uart     *uartImpl
	spi      *spiImpl
	i2c      *i2cImpl
}

// NewDevice creates a new unconnected Device instance.
func NewDevice() *Device {
	d := &Device{}
	d.scope = &scopeImpl{dev: d}
	d.wavegen = &wavegenImpl{dev: d}
	d.supply = &supplyImpl{dev: d}
	d.dmm = &dmmImpl{dev: d}
	d.logic = &logicImpl{dev: d}
	d.pattern = &patternImpl{dev: d}
	d.staticIO = &staticIOImpl{dev: d}
	d.uart = &uartImpl{dev: d}
	d.spi = &spiImpl{dev: d}
	d.i2c = &i2cImpl{dev: d}
	return d
}

// EnumDevices discovers all connected Digilent devices and returns their info.
// This does not open any device — it only enumerates what is available.
func (d *Device) EnumDevices() ([]EnumDevice, error) {
	count, err := dwfEnum(cEnumfilterAll)
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, nil
	}

	devices := make([]EnumDevice, count)
	for i := 0; i < count; i++ {
		ci := cInt(i)
		ed := EnumDevice{Index: i}
		if name, err := dwfEnumDeviceName(ci); err == nil {
			ed.DeviceName = name
		}
		if uname, err := dwfEnumUserName(ci); err == nil {
			ed.UserName = uname
		}
		if sn, err := dwfEnumSN(ci); err == nil {
			ed.SerialNumber = sn
		}
		if opened, err := dwfEnumDeviceIsOpened(ci); err == nil {
			ed.IsOpened = opened
		}
		devices[i] = ed
	}
	return devices, nil
}

// EnumConfigs returns the available hardware configurations for the device at
// the given index. Call this after dwfEnum but before Open to inspect what
// resource trade-offs each configuration offers.
func (d *Device) EnumConfigs(deviceIndex int) ([]DeviceConfig, error) {
	count, err := dwfEnumConfig(cInt(deviceIndex))
	if err != nil {
		return nil, err
	}

	configs := make([]DeviceConfig, count)
	for i := 0; i < count; i++ {
		ci := cInt(i)
		cfg := DeviceConfig{}
		if v, err := dwfEnumConfigInfo(ci, cDECIAnalogInChannelCount); err == nil {
			cfg.AnalogInChannels = v
		}
		if v, err := dwfEnumConfigInfo(ci, cDECIAnalogOutChannelCount); err == nil {
			cfg.AnalogOutChannels = v
		}
		if v, err := dwfEnumConfigInfo(ci, cDECIAnalogIOChannelCount); err == nil {
			cfg.AnalogIOChannels = v
		}
		if v, err := dwfEnumConfigInfo(ci, cDECIDigitalInChannelCount); err == nil {
			cfg.DigitalInChannels = v
		}
		if v, err := dwfEnumConfigInfo(ci, cDECIDigitalOutChannelCount); err == nil {
			cfg.DigitalOutChannels = v
		}
		if v, err := dwfEnumConfigInfo(ci, cDECIDigitalIOChannelCount); err == nil {
			cfg.DigitalIOChannels = v
		}
		if v, err := dwfEnumConfigInfo(ci, cDECIAnalogInBufferSize); err == nil {
			cfg.AnalogInBufferSize = v
		}
		if v, err := dwfEnumConfigInfo(ci, cDECIAnalogOutBufferSize); err == nil {
			cfg.AnalogOutBufferSize = v
		}
		if v, err := dwfEnumConfigInfo(ci, cDECIDigitalInBufferSize); err == nil {
			cfg.DigitalInBufferSize = v
		}
		if v, err := dwfEnumConfigInfo(ci, cDECIDigitalOutBufferSize); err == nil {
			cfg.DigitalOutBufferSize = v
		}
		configs[i] = cfg
	}
	return configs, nil
}

// Open connects to a Digilent device.
func (d *Device) Open(device string, config int) (*DeviceInfo, error) {
	filter := cEnumfilterAll
	if devID, ok := deviceNames[device]; ok {
		filter = cInt(int(devID))
	}

	count, err := dwfEnum(filter)
	if err != nil {
		return nil, err
	}
	if count <= 0 {
		if device == "" {
			return nil, fmt.Errorf("no connected devices found")
		}
		return nil, fmt.Errorf("no %s connected", device)
	}

	// attempt to open the first available device
	var hdwf DevHandle
	var openErr error
	for i := 0; i < count; i++ {
		hdwf, openErr = dwfDeviceConfigOpen(cInt(i), cInt(config))
		if hdwf != 0 {
			break
		}
	}
	if hdwf == 0 {
		if openErr != nil {
			return nil, openErr
		}
		return nil, fmt.Errorf("failed to open device")
	}
	d.handle = hdwf

	// detect device type
	devName := ""
	serialNum := ""
	if count > 0 {
		devID, _, err := dwfEnumDeviceType(0)
		if err == nil {
			if name, ok := deviceIDToName[devID]; ok {
				devName = name
			}
		}
		if sn, err := dwfEnumSN(0); err == nil {
			serialNum = sn
		}
	}

	// get version
	version, _ := dwfGetVersion()

	// query device capabilities
	info := &DeviceInfo{
		Handle:       int(hdwf),
		Name:         devName,
		SerialNumber: serialNum,
		Version:      version,
	}

	if n, err := dwfAnalogInChannelCount(hdwf); err == nil {
		info.AnalogInChannels = n
	}
	if n, err := dwfAnalogOutCount(hdwf); err == nil {
		info.AnalogOutChannels = n
	}
	if n, err := dwfAnalogInBufferSizeInfo(hdwf); err == nil {
		info.MaxAnalogInBufferSize = n
	}
	if n, err := dwfAnalogInBitsInfo(hdwf); err == nil {
		info.MaxAnalogInResolution = n
	}
	if n, err := dwfDigitalInBitsInfo(hdwf); err == nil {
		info.DigitalInChannels = n
	}
	if n, err := dwfDigitalOutCount(hdwf); err == nil {
		info.DigitalOutChannels = n
	}

	d.info = info
	return info, nil
}

// Close disconnects from the device.
func (d *Device) Close() error {
	if d.handle != 0 {
		err := dwfDeviceClose(d.handle)
		d.handle = 0
		return err
	}
	return nil
}

// Temperature returns the device board temperature in °C.
func (d *Device) Temperature() (float64, error) {
	chCount, err := dwfAnalogIOChannelCount(d.handle)
	if err != nil {
		return 0, err
	}

	for ch := 0; ch < chCount; ch++ {
		_, label, err := dwfAnalogIOChannelName(d.handle, cInt(ch))
		if err != nil || label != "System" {
			continue
		}
		nodeCount, err := dwfAnalogIOChannelInfo(d.handle, cInt(ch))
		if err != nil {
			continue
		}
		for n := 0; n < nodeCount; n++ {
			name, _, err := dwfAnalogIOChannelNodeName(d.handle, cInt(ch), cInt(n))
			if err != nil || name != "Temp" {
				continue
			}
			if err := dwfAnalogIOStatus(d.handle); err != nil {
				return 0, err
			}
			return dwfAnalogIOChannelNodeStatus(d.handle, cInt(ch), cInt(n))
		}
	}
	return 0, nil
}

// Instrument accessors
func (d *Device) Scope() Oscilloscope       { return d.scope }
func (d *Device) Wavegen() WavegenDriver    { return d.wavegen }
func (d *Device) Supply() PowerSupply       { return d.supply }
func (d *Device) DMM() DigitalMultimeter    { return d.dmm }
func (d *Device) Logic() LogicAnalyzer      { return d.logic }
func (d *Device) Pattern() PatternGenerator { return d.pattern }
func (d *Device) Static() StaticIO          { return d.staticIO }
func (d *Device) UARTProtocol() UART        { return d.uart }
func (d *Device) SPIProtocol() SPI          { return d.spi }
func (d *Device) I2CProtocol() I2C          { return d.i2c }

// ==================== Oscilloscope ====================

type scopeImpl struct {
	dev        *Device
	bufferSize int
}

func (s *scopeImpl) Open(cfg ScopeConfig) error {
	h := s.dev.handle
	if err := dwfAnalogInChannelEnableSet(h, -1, true); err != nil {
		return err
	}
	if err := dwfAnalogInChannelOffsetSet(h, -1, cfg.OffsetVoltage); err != nil {
		return err
	}
	if err := dwfAnalogInChannelRangeSet(h, -1, cfg.AmplitudeRange); err != nil {
		return err
	}

	maxBuf := 0
	if s.dev.info != nil {
		maxBuf = s.dev.info.MaxAnalogInBufferSize
	}
	bufSize := cfg.BufferSize
	if bufSize == 0 || bufSize > maxBuf {
		bufSize = maxBuf
	}
	s.bufferSize = bufSize
	if err := dwfAnalogInBufferSizeSet(h, bufSize); err != nil {
		return err
	}
	if err := dwfAnalogInFrequencySet(h, cfg.SamplingFrequency); err != nil {
		return err
	}
	return dwfAnalogInChannelFilterSet(h, -1, cFilterDecimate)
}

func (s *scopeImpl) Measure(channel int) (float64, error) {
	h := s.dev.handle
	if err := dwfAnalogInConfigure(h, false, false); err != nil {
		return 0, err
	}
	if _, err := dwfAnalogInStatus(h, false); err != nil {
		return 0, err
	}
	return dwfAnalogInStatusSample(h, cInt(channel-1))
}

func (s *scopeImpl) SetTrigger(cfg TriggerConfig) error {
	h := s.dev.handle
	if cfg.Enable && cfg.Source != TrigSrcNone {
		if err := dwfAnalogInTriggerAutoTimeoutSet(h, cfg.Timeout); err != nil {
			return err
		}
		if err := dwfAnalogInTriggerSourceSet(h, cTrigSrc(cfg.Source)); err != nil {
			return err
		}
		ch := cfg.Channel
		if cfg.Source == TrigSrcDetectorAnalogIn {
			ch--
		}
		if err := dwfAnalogInTriggerChannelSet(h, cInt(ch)); err != nil {
			return err
		}
		if err := dwfAnalogInTriggerTypeSet(h, cTrigtypeEdge); err != nil {
			return err
		}
		if err := dwfAnalogInTriggerLevelSet(h, cfg.Level); err != nil {
			return err
		}
		if cfg.EdgeRising {
			return dwfAnalogInTriggerConditionSet(h, cDwfTriggerSlopeRise)
		}
		return dwfAnalogInTriggerConditionSet(h, cDwfTriggerSlopeFall)
	}
	return dwfAnalogInTriggerSourceSet(h, cTrigsrcNone)
}

func (s *scopeImpl) Record(channel int) ([]float64, error) {
	h := s.dev.handle
	if err := dwfAnalogInConfigure(h, false, true); err != nil {
		return nil, err
	}
	for {
		status, err := dwfAnalogInStatus(h, true)
		if err != nil {
			return nil, err
		}
		if status == cDwfStateDone {
			break
		}
	}
	return dwfAnalogInStatusData(h, cInt(channel-1), s.bufferSize)
}

func (s *scopeImpl) Close() error {
	return dwfAnalogInReset(s.dev.handle)
}

// ==================== Wavegen ====================

type wavegenImpl struct {
	dev *Device
}

func (w *wavegenImpl) Generate(cfg WavegenConfig) error {
	h := w.dev.handle
	ch := cInt(cfg.Channel - 1)
	node := cAnalogOutNodeCarrier

	if err := dwfAnalogOutNodeEnableSet(h, ch, node, true); err != nil {
		return err
	}
	if err := dwfAnalogOutNodeFunctionSet(h, ch, node, cFunc(cfg.Function)); err != nil {
		return err
	}
	if cfg.Function == FuncCustom && len(cfg.CustomData) > 0 {
		if err := dwfAnalogOutNodeDataSet(h, ch, node, cfg.CustomData); err != nil {
			return err
		}
	}
	if err := dwfAnalogOutNodeFrequencySet(h, ch, node, cfg.Frequency); err != nil {
		return err
	}
	if err := dwfAnalogOutNodeAmplitudeSet(h, ch, node, cfg.Amplitude); err != nil {
		return err
	}
	if err := dwfAnalogOutNodeOffsetSet(h, ch, node, cfg.Offset); err != nil {
		return err
	}
	if err := dwfAnalogOutNodeSymmetrySet(h, ch, node, cfg.Symmetry); err != nil {
		return err
	}
	if err := dwfAnalogOutRunSet(h, ch, cfg.RunTime); err != nil {
		return err
	}
	if err := dwfAnalogOutWaitSet(h, ch, cfg.Wait); err != nil {
		return err
	}
	if err := dwfAnalogOutRepeatSet(h, ch, cfg.Repeat); err != nil {
		return err
	}
	return dwfAnalogOutConfigure(h, ch, true)
}

func (w *wavegenImpl) Enable(channel int) error {
	return dwfAnalogOutConfigure(w.dev.handle, cInt(channel-1), true)
}

func (w *wavegenImpl) Disable(channel int) error {
	return dwfAnalogOutConfigure(w.dev.handle, cInt(channel-1), false)
}

func (w *wavegenImpl) Close(channel int) error {
	return dwfAnalogOutReset(w.dev.handle, cInt(channel-1))
}

// ==================== Power Supply ====================

type supplyImpl struct {
	dev *Device
}

func (s *supplyImpl) findChannelNode(label, nodeName string) (int, int, bool) {
	h := s.dev.handle
	chCount, err := dwfAnalogIOChannelCount(h)
	if err != nil {
		return -1, -1, false
	}
	for ch := 0; ch < chCount; ch++ {
		_, lbl, err := dwfAnalogIOChannelName(h, cInt(ch))
		if err != nil || lbl != label {
			continue
		}
		nodeCount, err := dwfAnalogIOChannelInfo(h, cInt(ch))
		if err != nil {
			continue
		}
		for n := 0; n < nodeCount; n++ {
			name, _, err := dwfAnalogIOChannelNodeName(h, cInt(ch), cInt(n))
			if err != nil {
				continue
			}
			if name == nodeName {
				return ch, n, true
			}
		}
	}
	return -1, -1, false
}

func (s *supplyImpl) setNode(labels []string, nodeName string, value float64) {
	h := s.dev.handle
	for _, label := range labels {
		if ch, node, ok := s.findChannelNode(label, nodeName); ok {
			_ = dwfAnalogIOChannelNodeSet(h, cInt(ch), cInt(node), value)
			return
		}
	}
}

func (s *supplyImpl) Switch(cfg SuppliesConfig) error {
	// positive supply
	posLabels := []string{"V+", "p25V"}
	enableVal := 0.0
	if cfg.PositiveState {
		enableVal = 1.0
	}
	s.setNode(posLabels, "Enable", enableVal)
	s.setNode(posLabels, "Voltage", cfg.PositiveVoltage)
	s.setNode(posLabels, "Current", cfg.PositiveCurrent)

	// negative supply
	negLabels := []string{"V-", "n25V"}
	enableVal = 0.0
	if cfg.NegativeState {
		enableVal = 1.0
	}
	s.setNode(negLabels, "Enable", enableVal)
	s.setNode(negLabels, "Voltage", cfg.NegativeVoltage)
	s.setNode(negLabels, "Current", cfg.NegativeCurrent)

	// digital/6V supply
	digLabels := []string{"VDD", "p6V"}
	enableVal = 0.0
	if cfg.State {
		enableVal = 1.0
	}
	s.setNode(digLabels, "Enable", enableVal)
	s.setNode(digLabels, "Voltage", cfg.Voltage)
	s.setNode(digLabels, "Current", cfg.Current)

	// master enable
	return dwfAnalogIOEnableSet(s.dev.handle, cfg.MasterState)
}

func (s *supplyImpl) Close() error {
	return dwfAnalogIOReset(s.dev.handle)
}

// ==================== DMM ====================

type dmmImpl struct {
	dev     *Device
	channel int
	nodes   struct {
		enable int
		mode   int
		rangN  int
		meas   int
		input  int
	}
}

func (m *dmmImpl) Open() error {
	h := m.dev.handle
	m.channel = -1
	m.nodes.enable = -1
	m.nodes.mode = -1
	m.nodes.rangN = -1
	m.nodes.meas = -1
	m.nodes.input = -1

	chCount, err := dwfAnalogIOChannelCount(h)
	if err != nil {
		return err
	}
	for ch := 0; ch < chCount; ch++ {
		_, label, err := dwfAnalogIOChannelName(h, cInt(ch))
		if err != nil || label != "DMM" {
			continue
		}
		m.channel = ch
		break
	}
	if m.channel < 0 {
		return fmt.Errorf("DMM not available on this device")
	}

	nodeCount, err := dwfAnalogIOChannelInfo(h, cInt(m.channel))
	if err != nil {
		return err
	}
	for n := 0; n < nodeCount; n++ {
		name, _, err := dwfAnalogIOChannelNodeName(h, cInt(m.channel), cInt(n))
		if err != nil {
			continue
		}
		switch name {
		case "Enable":
			m.nodes.enable = n
		case "Mode":
			m.nodes.mode = n
		case "Range":
			m.nodes.rangN = n
		case "Meas":
			m.nodes.meas = n
		case "Input":
			m.nodes.input = n
		}
	}

	if m.nodes.enable >= 0 {
		return dwfAnalogIOChannelNodeSet(h, cInt(m.channel), cInt(m.nodes.enable), 1.0)
	}
	return nil
}

func (m *dmmImpl) Measure(mode DMMMode, range_ float64, highImpedance bool) (float64, error) {
	h := m.dev.handle
	if m.nodes.input >= 0 {
		inputVal := 0.0
		if highImpedance {
			inputVal = 1.0
		}
		if err := dwfAnalogIOChannelNodeSet(h, cInt(m.channel), cInt(m.nodes.input), inputVal); err != nil {
			return 0, err
		}
	}
	if m.nodes.mode >= 0 {
		if err := dwfAnalogIOChannelNodeSet(h, cInt(m.channel), cInt(m.nodes.mode), float64(mode)); err != nil {
			return 0, err
		}
	}
	if m.nodes.rangN >= 0 {
		if err := dwfAnalogIOChannelNodeSet(h, cInt(m.channel), cInt(m.nodes.rangN), range_); err != nil {
			return 0, err
		}
	}
	if err := dwfAnalogIOStatus(h); err != nil {
		return -1, err
	}
	if m.nodes.meas >= 0 {
		return dwfAnalogIOChannelNodeStatus(h, cInt(m.channel), cInt(m.nodes.meas))
	}
	return -1, fmt.Errorf("DMM measurement node not found")
}

func (m *dmmImpl) Close() error {
	h := m.dev.handle
	if m.nodes.enable >= 0 {
		_ = dwfAnalogIOChannelNodeSet(h, cInt(m.channel), cInt(m.nodes.enable), 0)
	}
	return dwfAnalogIOReset(h)
}

// ==================== Logic Analyzer ====================

type logicImpl struct {
	dev        *Device
	bufferSize int
}

func (l *logicImpl) Open(cfg LogicConfig) error {
	h := l.dev.handle
	maxBuf, _ := dwfDigitalInBufferSizeInfo(h)
	l.bufferSize = cfg.BufferSize
	if l.bufferSize == 0 || l.bufferSize > maxBuf {
		l.bufferSize = maxBuf
	}

	internalFreq, err := dwfDigitalInInternalClockInfo(h)
	if err != nil {
		return err
	}
	divider := int(internalFreq / cfg.SamplingFrequency)
	if err := dwfDigitalInDividerSet(h, divider); err != nil {
		return err
	}
	if err := dwfDigitalInSampleFormatSet(h, 16); err != nil {
		return err
	}
	return dwfDigitalInBufferSizeSet(h, l.bufferSize)
}

func (l *logicImpl) SetTrigger(cfg LogicTriggerConfig) error {
	h := l.dev.handle
	if cfg.Enable {
		if err := dwfDigitalInTriggerSourceSet(h, cTrigsrcDetectorDigIn); err != nil {
			return err
		}
	} else {
		return dwfDigitalInTriggerSourceSet(h, cTrigsrcNone)
	}

	pos := cfg.Position
	if pos < 0 {
		pos = 0
	}
	if pos > l.bufferSize {
		pos = l.bufferSize
	}
	if err := dwfDigitalInTriggerPositionSet(h, l.bufferSize-pos); err != nil {
		return err
	}
	if err := dwfDigitalInTriggerPrefillSet(h, pos); err != nil {
		return err
	}

	chBit := cUint(1 << cfg.Channel)
	if cfg.RisingEdge {
		if err := dwfDigitalInTriggerSet(h, 0, chBit, 0, 0); err != nil {
			return err
		}
		if err := dwfDigitalInTriggerResetSet(h, 0, 0, chBit, 0); err != nil {
			return err
		}
	} else {
		if err := dwfDigitalInTriggerSet(h, chBit, 0, 0, 0); err != nil {
			return err
		}
		if err := dwfDigitalInTriggerResetSet(h, 0, 0, 0, chBit); err != nil {
			return err
		}
	}

	if err := dwfDigitalInTriggerAutoTimeoutSet(h, cfg.Timeout); err != nil {
		return err
	}
	if err := dwfDigitalInTriggerLengthSet(h, cfg.LengthMin, cfg.LengthMax, 0); err != nil {
		return err
	}
	return dwfDigitalInTriggerCountSet(h, cInt(cfg.Count), 0)
}

func (l *logicImpl) Record(channel int) ([]uint16, error) {
	h := l.dev.handle
	if err := dwfDigitalInConfigure(h, false, true); err != nil {
		return nil, err
	}
	for {
		status, err := dwfDigitalInStatus(h, true)
		if err != nil {
			return nil, err
		}
		if status == cDwfStateDone {
			break
		}
	}
	buffer := make([]uint16, l.bufferSize)
	if err := dwfDigitalInStatusData(h, buffer); err != nil {
		return nil, err
	}
	for i := range buffer {
		buffer[i] = (buffer[i] & (1 << channel)) >> channel
	}
	return buffer, nil
}

func (l *logicImpl) Close() error {
	return dwfDigitalInReset(l.dev.handle)
}

// ==================== Pattern Generator ====================

type patternImpl struct {
	dev *Device
}

func (p *patternImpl) Generate(cfg PatternConfig) error {
	h := p.dev.handle
	ch := cInt(cfg.Channel)
	if p.dev.info != nil && p.dev.info.Name == "Digital Discovery" {
		ch = cInt(cfg.Channel - 24)
	}

	internalFreq, err := dwfDigitalOutInternalClockInfo(h)
	if err != nil {
		return err
	}

	if err := dwfDigitalOutEnableSet(h, ch, true); err != nil {
		return err
	}
	if err := dwfDigitalOutTypeSet(h, ch, cDigitalOutType(cfg.Function)); err != nil {
		return err
	}

	divider := int(internalFreq / cfg.Frequency)
	if err := dwfDigitalOutDividerSet(h, ch, divider); err != nil {
		return err
	}
	if err := dwfDigitalOutIdleSet(h, ch, cDigitalOutIdle(cfg.IdleState)); err != nil {
		return err
	}

	runTime := cfg.RunTime
	if runTime < 0 && len(cfg.Data) > 0 {
		runTime = int(float64(len(cfg.Data)) / cfg.Frequency)
	}
	if err := dwfDigitalOutRunSet(h, float64(runTime)); err != nil {
		return err
	}
	if err := dwfDigitalOutWaitSet(h, cfg.Wait); err != nil {
		return err
	}
	if err := dwfDigitalOutRepeatSet(h, cfg.Repeat); err != nil {
		return err
	}

	if err := dwfDigitalOutRepeatTriggerSet(h, cfg.TriggerEnabled); err != nil {
		return err
	}
	if cfg.TriggerEnabled {
		if err := dwfDigitalOutTriggerSourceSet(h, cTrigSrc(cfg.TriggerSource)); err != nil {
			return err
		}
		if cfg.TriggerEdgeRising {
			if err := dwfDigitalOutTriggerSlopeSet(h, cDwfTriggerSlopeRise); err != nil {
				return err
			}
		} else {
			if err := dwfDigitalOutTriggerSlopeSet(h, cDwfTriggerSlopeFall); err != nil {
				return err
			}
		}
	}

	if cfg.Function == DigitalOutTypePulse {
		steps := int(internalFreq/cfg.Frequency) / divider
		high := int(float64(steps) * cfg.DutyCycle / 100)
		low := steps - high
		if err := dwfDigitalOutCounterSet(h, ch, low, high); err != nil {
			return err
		}
	} else if cfg.Function == DigitalOutTypeCustom && len(cfg.Data) > 0 {
		if err := dwfDigitalOutDataSet(h, ch, cfg.Data); err != nil {
			return err
		}
	}

	return dwfDigitalOutConfigure(h, true)
}

func (p *patternImpl) Enable(channel int) error {
	h := p.dev.handle
	ch := cInt(channel)
	if p.dev.info != nil && p.dev.info.Name == "Digital Discovery" {
		ch = cInt(channel - 24)
	}
	if err := dwfDigitalOutEnableSet(h, ch, true); err != nil {
		return err
	}
	return dwfDigitalOutConfigure(h, true)
}

func (p *patternImpl) Disable(channel int) error {
	h := p.dev.handle
	ch := cInt(channel)
	if p.dev.info != nil && p.dev.info.Name == "Digital Discovery" {
		ch = cInt(channel - 24)
	}
	if err := dwfDigitalOutEnableSet(h, ch, false); err != nil {
		return err
	}
	return dwfDigitalOutConfigure(h, true)
}

func (p *patternImpl) Close() error {
	return dwfDigitalOutReset(p.dev.handle)
}

// ==================== Static I/O ====================

type staticIOImpl struct {
	dev *Device
}

func (s *staticIOImpl) channelCount() int {
	if s.dev.info == nil {
		return 16
	}
	in := s.dev.info.DigitalInChannels
	out := s.dev.info.DigitalOutChannels
	if in < out {
		return in
	}
	return out
}

func (s *staticIOImpl) adjustChannel(channel int) int {
	if s.dev.info != nil && s.dev.info.Name == "Digital Discovery" {
		return channel - 24
	}
	return channel
}

func rotateLeft(number, position, size uint32) uint32 {
	return (number << position) | (number >> (size - position))
}

func (s *staticIOImpl) SetMode(channel int, output bool) error {
	h := s.dev.handle
	ch := s.adjustChannel(channel)
	count := uint32(s.channelCount())

	mask, err := dwfDigitalIOOutputEnableGet(h)
	if err != nil {
		return err
	}
	if output {
		mask |= rotateLeft(1, uint32(ch), count)
	} else {
		bits := uint32((1 << count) - 2)
		mask &= rotateLeft(bits, uint32(ch), count)
	}
	return dwfDigitalIOOutputEnableSet(h, mask)
}

func (s *staticIOImpl) GetState(channel int) (bool, error) {
	h := s.dev.handle
	ch := s.adjustChannel(channel)

	if err := dwfDigitalIOStatus(h); err != nil {
		return false, err
	}
	data, err := dwfDigitalIOInputStatus(h)
	if err != nil {
		return false, err
	}
	return data&(1<<ch) != 0, nil
}

func (s *staticIOImpl) SetState(channel int, value bool) error {
	h := s.dev.handle
	ch := s.adjustChannel(channel)
	count := uint32(s.channelCount())

	mask, err := dwfDigitalIOOutputGet(h)
	if err != nil {
		return err
	}
	if value {
		mask |= rotateLeft(1, uint32(ch), count)
	} else {
		bits := uint32((1 << count) - 2)
		mask &= rotateLeft(bits, uint32(ch), count)
	}
	return dwfDigitalIOOutputSet(h, mask)
}

func (s *staticIOImpl) SetCurrent(current float64) error {
	h := s.dev.handle
	chCount, err := dwfAnalogIOChannelCount(h)
	if err != nil {
		return err
	}
	for ch := 0; ch < chCount; ch++ {
		_, label, err := dwfAnalogIOChannelName(h, cInt(ch))
		if err != nil || label != "VDD" {
			continue
		}
		nodeCount, err := dwfAnalogIOChannelInfo(h, cInt(ch))
		if err != nil {
			continue
		}
		for n := 0; n < nodeCount; n++ {
			name, _, err := dwfAnalogIOChannelNodeName(h, cInt(ch), cInt(n))
			if err != nil || name != "Drive" {
				continue
			}
			return dwfAnalogIOChannelNodeSet(h, cInt(ch), cInt(n), current)
		}
	}
	return fmt.Errorf("drive current node not found")
}

func (s *staticIOImpl) SetPull(channel int, direction PullDirection) error {
	_ = channel
	_ = direction
	return fmt.Errorf("SetPull: not yet implemented for this device")
}

func (s *staticIOImpl) Close() error {
	return dwfDigitalIOReset(s.dev.handle)
}

// ==================== UART ====================

type uartImpl struct {
	dev *Device
}

func (u *uartImpl) Open(cfg UARTConfig) error {
	h := u.dev.handle
	if err := dwfDigitalUartRateSet(h, float64(cfg.BaudRate)); err != nil {
		return err
	}
	if err := dwfDigitalUartTxSet(h, cInt(cfg.TX)); err != nil {
		return err
	}
	if err := dwfDigitalUartRxSet(h, cInt(cfg.RX)); err != nil {
		return err
	}
	if err := dwfDigitalUartBitsSet(h, cInt(cfg.DataBits)); err != nil {
		return err
	}
	if err := dwfDigitalUartParitySet(h, cInt(cfg.Parity)); err != nil {
		return err
	}
	if err := dwfDigitalUartStopSet(h, float64(cfg.StopBits)); err != nil {
		return err
	}
	_ = dwfDigitalUartTx(h, nil)
	_, _, _ = dwfDigitalUartRx(h, 0)
	return nil
}

func (u *uartImpl) Read() ([]byte, error) {
	h := u.dev.handle
	maxBuf := 8192
	if u.dev.info != nil && u.dev.info.MaxAnalogInBufferSize > 0 {
		maxBuf = u.dev.info.MaxAnalogInBufferSize
	}

	data, parity, err := dwfDigitalUartRx(h, maxBuf)
	if err != nil {
		return nil, err
	}
	if parity < 0 {
		return data, fmt.Errorf("UART buffer overflow")
	}
	if parity > 0 {
		return data, fmt.Errorf("UART parity error at index %d", parity)
	}
	return data, nil
}

func (u *uartImpl) Write(data []byte) error {
	return dwfDigitalUartTx(u.dev.handle, data)
}

func (u *uartImpl) Close() error {
	return dwfDigitalUartReset(u.dev.handle)
}

// ==================== SPI ====================

type spiImpl struct {
	dev *Device
}

func (sp *spiImpl) Open(cfg SPIConfig) error {
	h := sp.dev.handle
	if err := dwfDigitalSpiFrequencySet(h, cfg.ClockFrequency); err != nil {
		return err
	}
	if err := dwfDigitalSpiClockSet(h, cInt(cfg.SCK)); err != nil {
		return err
	}
	if cfg.MOSI >= 0 {
		if err := dwfDigitalSpiDataSet(h, 0, cInt(cfg.MOSI)); err != nil {
			return err
		}
		if err := dwfDigitalSpiIdleSet(h, 0, cDwfDigitalOutIdleZet); err != nil {
			return err
		}
	}
	if cfg.MISO >= 0 {
		if err := dwfDigitalSpiDataSet(h, 1, cInt(cfg.MISO)); err != nil {
			return err
		}
		if err := dwfDigitalSpiIdleSet(h, 1, cDwfDigitalOutIdleZet); err != nil {
			return err
		}
	}
	if err := dwfDigitalSpiModeSet(h, cInt(cfg.Mode)); err != nil {
		return err
	}
	order := 0
	if cfg.MSBFirst {
		order = 1
	}
	if err := dwfDigitalSpiOrderSet(h, cInt(order)); err != nil {
		return err
	}
	if err := dwfDigitalSpiSelect(h, cInt(cfg.CS), 1); err != nil {
		return err
	}
	return dwfDigitalSpiWriteOne(h, 1, 0, 0)
}

func (sp *spiImpl) Read(count int, cs int) ([]byte, error) {
	h := sp.dev.handle
	if err := dwfDigitalSpiSelect(h, cInt(cs), 0); err != nil {
		return nil, err
	}
	buf := make([]byte, count)
	if err := dwfDigitalSpiRead(h, 1, 8, buf); err != nil {
		_ = dwfDigitalSpiSelect(h, cInt(cs), 1)
		return nil, err
	}
	if err := dwfDigitalSpiSelect(h, cInt(cs), 1); err != nil {
		return buf, err
	}
	return buf, nil
}

func (sp *spiImpl) Write(data []byte, cs int) error {
	h := sp.dev.handle
	if err := dwfDigitalSpiSelect(h, cInt(cs), 0); err != nil {
		return err
	}
	if err := dwfDigitalSpiWrite(h, 1, 8, data); err != nil {
		_ = dwfDigitalSpiSelect(h, cInt(cs), 1)
		return err
	}
	return dwfDigitalSpiSelect(h, cInt(cs), 1)
}

func (sp *spiImpl) Exchange(txData []byte, rxCount int, cs int) ([]byte, error) {
	h := sp.dev.handle
	if err := dwfDigitalSpiSelect(h, cInt(cs), 0); err != nil {
		return nil, err
	}
	rxBuf := make([]byte, rxCount)
	if err := dwfDigitalSpiWriteRead(h, 1, 8, txData, rxBuf); err != nil {
		_ = dwfDigitalSpiSelect(h, cInt(cs), 1)
		return nil, err
	}
	if err := dwfDigitalSpiSelect(h, cInt(cs), 1); err != nil {
		return rxBuf, err
	}
	return rxBuf, nil
}

func (sp *spiImpl) Close() error {
	return dwfDigitalSpiReset(sp.dev.handle)
}

// ==================== I2C ====================

type i2cImpl struct {
	dev *Device
}

func (ic *i2cImpl) Open(cfg I2CConfig) error {
	h := ic.dev.handle
	if err := dwfDigitalI2cReset(h); err != nil {
		return err
	}
	if err := dwfDigitalI2cStretchSet(h, cfg.Stretching); err != nil {
		return err
	}
	if err := dwfDigitalI2cRateSet(h, cfg.ClockRate); err != nil {
		return err
	}
	if err := dwfDigitalI2cSclSet(h, cInt(cfg.SCL)); err != nil {
		return err
	}
	if err := dwfDigitalI2cSdaSet(h, cInt(cfg.SDA)); err != nil {
		return err
	}

	nak, err := dwfDigitalI2cClear(h)
	if err != nil {
		return err
	}
	if nak == 0 {
		return fmt.Errorf("I2C bus lockup")
	}

	_, _ = dwfDigitalI2cWrite(h, 0, nil)
	return nil
}

func (ic *i2cImpl) Scan() ([]int, error) {
	h := ic.dev.handle
	var found []int
	for addr := 0x08; addr <= 0x77; addr++ {
		nak, err := dwfDigitalI2cWrite(h, cInt(addr<<1), nil)
		if err != nil {
			return nil, err
		}
		if nak == 0 {
			found = append(found, addr)
		}
	}
	return found, nil
}

func (ic *i2cImpl) Read(count int, address int) ([]byte, error) {
	h := ic.dev.handle
	buf := make([]byte, count)
	nak, err := dwfDigitalI2cRead(h, cInt(address<<1), buf)
	if err != nil {
		return nil, err
	}
	if nak != 0 {
		return buf, fmt.Errorf("I2C NAK at index %d", nak)
	}
	return buf, nil
}

func (ic *i2cImpl) Write(data []byte, address int) error {
	h := ic.dev.handle
	nak, err := dwfDigitalI2cWrite(h, cInt(address<<1), data)
	if err != nil {
		return err
	}
	if nak != 0 {
		return fmt.Errorf("I2C NAK at index %d", nak)
	}
	return nil
}

func (ic *i2cImpl) Exchange(txData []byte, rxCount int, address int) ([]byte, error) {
	h := ic.dev.handle
	rxBuf := make([]byte, rxCount)
	nak, err := dwfDigitalI2cWriteRead(h, cInt(address<<1), txData, rxBuf)
	if err != nil {
		return nil, err
	}
	if nak != 0 {
		return rxBuf, fmt.Errorf("I2C NAK at index %d", nak)
	}
	return rxBuf, nil
}

func (ic *i2cImpl) Close() error {
	return dwfDigitalI2cReset(ic.dev.handle)
}

// Compile-time interface checks
var _ DiscoveryDevice = (*Device)(nil)
