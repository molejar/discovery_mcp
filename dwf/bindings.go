package dwf

/*
#cgo LDFLAGS: -ldwf
#include <stdlib.h>
#include <digilent/waveforms/dwf.h>
*/
import "C"

import (
	"fmt"
	"unsafe"
)

// --- Error handling ---

// lastError returns the last error message from the DWF SDK.
func lastError() error {
	var buf [512]C.char
	C.FDwfGetLastErrorMsg(&buf[0])
	msg := C.GoString(&buf[0])
	if msg == "" {
		return fmt.Errorf("unknown DWF SDK error")
	}
	return fmt.Errorf("dwf: %s", msg)
}

// --- Device functions ---

func dwfGetVersion() (string, error) {
	var buf [32]C.char
	if C.FDwfGetVersion(&buf[0]) == 0 {
		return "", lastError()
	}
	return C.GoString(&buf[0]), nil
}

func dwfEnum(filter C.int) (int, error) {
	var count C.int
	if C.FDwfEnum(filter, &count) == 0 {
		return 0, lastError()
	}
	return int(count), nil
}

func dwfDeviceConfigOpen(index, config C.int) (C.HDWF, error) {
	var hdwf C.HDWF
	if C.FDwfDeviceConfigOpen(index, config, &hdwf) == 0 {
		return 0, lastError()
	}
	return hdwf, nil
}

func dwfEnumDeviceType(index C.int) (int, int, error) {
	var devID, devRev C.int
	if C.FDwfEnumDeviceType(index, &devID, &devRev) == 0 {
		return 0, 0, lastError()
	}
	return int(devID), int(devRev), nil
}

func dwfEnumSN(index C.int) (string, error) {
	var buf [32]C.char
	if C.FDwfEnumSN(index, &buf[0]) == 0 {
		return "", lastError()
	}
	return C.GoString(&buf[0]), nil
}

func dwfEnumDeviceName(index C.int) (string, error) {
	var buf [32]C.char
	if C.FDwfEnumDeviceName(index, &buf[0]) == 0 {
		return "", lastError()
	}
	return C.GoString(&buf[0]), nil
}

func dwfEnumUserName(index C.int) (string, error) {
	var buf [32]C.char
	if C.FDwfEnumUserName(index, &buf[0]) == 0 {
		return "", lastError()
	}
	return C.GoString(&buf[0]), nil
}

func dwfEnumDeviceIsOpened(index C.int) (bool, error) {
	var opened C.int
	if C.FDwfEnumDeviceIsOpened(index, &opened) == 0 {
		return false, lastError()
	}
	return opened != 0, nil
}

func dwfEnumConfig(index C.int) (int, error) {
	var count C.int
	if C.FDwfEnumConfig(index, &count) == 0 {
		return 0, lastError()
	}
	return int(count), nil
}

func dwfEnumConfigInfo(config C.int, info C.DwfEnumConfigInfo) (int, error) {
	var val C.int
	if C.FDwfEnumConfigInfo(config, info, &val) == 0 {
		return 0, lastError()
	}
	return int(val), nil
}

func dwfDeviceClose(hdwf C.HDWF) error {
	if C.FDwfDeviceClose(hdwf) == 0 {
		return lastError()
	}
	return nil
}

// --- Analog Input (Oscilloscope) ---

func dwfAnalogInChannelCount(hdwf C.HDWF) (int, error) {
	var count C.int
	if C.FDwfAnalogInChannelCount(hdwf, &count) == 0 {
		return 0, lastError()
	}
	return int(count), nil
}

func dwfAnalogInBufferSizeInfo(hdwf C.HDWF) (int, error) {
	var maxSize C.int
	if C.FDwfAnalogInBufferSizeInfo(hdwf, nil, &maxSize) == 0 {
		return 0, lastError()
	}
	return int(maxSize), nil
}

func dwfAnalogInBitsInfo(hdwf C.HDWF) (int, error) {
	var bits C.int
	if C.FDwfAnalogInBitsInfo(hdwf, &bits) == 0 {
		return 0, lastError()
	}
	return int(bits), nil
}

func dwfAnalogInChannelEnableSet(hdwf C.HDWF, channel C.int, enable bool) error {
	var e C.int
	if enable {
		e = 1
	}
	if C.FDwfAnalogInChannelEnableSet(hdwf, channel, e) == 0 {
		return lastError()
	}
	return nil
}

func dwfAnalogInChannelOffsetSet(hdwf C.HDWF, channel C.int, offset float64) error {
	if C.FDwfAnalogInChannelOffsetSet(hdwf, channel, C.double(offset)) == 0 {
		return lastError()
	}
	return nil
}

func dwfAnalogInChannelRangeSet(hdwf C.HDWF, channel C.int, volts float64) error {
	if C.FDwfAnalogInChannelRangeSet(hdwf, channel, C.double(volts)) == 0 {
		return lastError()
	}
	return nil
}

func dwfAnalogInBufferSizeSet(hdwf C.HDWF, size int) error {
	if C.FDwfAnalogInBufferSizeSet(hdwf, C.int(size)) == 0 {
		return lastError()
	}
	return nil
}

func dwfAnalogInFrequencySet(hdwf C.HDWF, freq float64) error {
	if C.FDwfAnalogInFrequencySet(hdwf, C.double(freq)) == 0 {
		return lastError()
	}
	return nil
}

func dwfAnalogInChannelFilterSet(hdwf C.HDWF, channel, filter C.int) error {
	if C.FDwfAnalogInChannelFilterSet(hdwf, channel, filter) == 0 {
		return lastError()
	}
	return nil
}

func dwfAnalogInConfigure(hdwf C.HDWF, reconfigure, start bool) error {
	var r, s C.int
	if reconfigure {
		r = 1
	}
	if start {
		s = 1
	}
	if C.FDwfAnalogInConfigure(hdwf, r, s) == 0 {
		return lastError()
	}
	return nil
}

func dwfAnalogInStatus(hdwf C.HDWF, readData bool) (byte, error) {
	var rd C.int
	if readData {
		rd = 1
	}
	var status C.DwfState
	if C.FDwfAnalogInStatus(hdwf, rd, &status) == 0 {
		return 0, lastError()
	}
	return byte(status), nil
}

func dwfAnalogInStatusSample(hdwf C.HDWF, channel C.int) (float64, error) {
	var voltage C.double
	if C.FDwfAnalogInStatusSample(hdwf, channel, &voltage) == 0 {
		return 0, lastError()
	}
	return float64(voltage), nil
}

func dwfAnalogInStatusData(hdwf C.HDWF, channel C.int, bufSize int) ([]float64, error) {
	buf := make([]float64, bufSize)
	if C.FDwfAnalogInStatusData(hdwf, channel, (*C.double)(unsafe.Pointer(&buf[0])), C.int(bufSize)) == 0 {
		return nil, lastError()
	}
	return buf, nil
}

func dwfAnalogInReset(hdwf C.HDWF) error {
	if C.FDwfAnalogInReset(hdwf) == 0 {
		return lastError()
	}
	return nil
}

// --- Trigger (Oscilloscope) ---

func dwfAnalogInTriggerAutoTimeoutSet(hdwf C.HDWF, timeout float64) error {
	if C.FDwfAnalogInTriggerAutoTimeoutSet(hdwf, C.double(timeout)) == 0 {
		return lastError()
	}
	return nil
}

func dwfAnalogInTriggerSourceSet(hdwf C.HDWF, src C.TRIGSRC) error {
	if C.FDwfAnalogInTriggerSourceSet(hdwf, src) == 0 {
		return lastError()
	}
	return nil
}

func dwfAnalogInTriggerChannelSet(hdwf C.HDWF, channel C.int) error {
	if C.FDwfAnalogInTriggerChannelSet(hdwf, channel) == 0 {
		return lastError()
	}
	return nil
}

func dwfAnalogInTriggerTypeSet(hdwf C.HDWF, trigType C.int) error {
	if C.FDwfAnalogInTriggerTypeSet(hdwf, trigType) == 0 {
		return lastError()
	}
	return nil
}

func dwfAnalogInTriggerLevelSet(hdwf C.HDWF, level float64) error {
	if C.FDwfAnalogInTriggerLevelSet(hdwf, C.double(level)) == 0 {
		return lastError()
	}
	return nil
}

func dwfAnalogInTriggerConditionSet(hdwf C.HDWF, cond C.DwfTriggerSlope) error {
	if C.FDwfAnalogInTriggerConditionSet(hdwf, cond) == 0 {
		return lastError()
	}
	return nil
}

// --- Analog Output (Wavegen) ---

func dwfAnalogOutCount(hdwf C.HDWF) (int, error) {
	var count C.int
	if C.FDwfAnalogOutCount(hdwf, &count) == 0 {
		return 0, lastError()
	}
	return int(count), nil
}

func dwfAnalogOutNodeEnableSet(hdwf C.HDWF, channel, node C.int, enable bool) error {
	var e C.int
	if enable {
		e = 1
	}
	if C.FDwfAnalogOutNodeEnableSet(hdwf, channel, node, e) == 0 {
		return lastError()
	}
	return nil
}

func dwfAnalogOutNodeFunctionSet(hdwf C.HDWF, channel, node C.int, function C.FUNC) error {
	if C.FDwfAnalogOutNodeFunctionSet(hdwf, channel, node, function) == 0 {
		return lastError()
	}
	return nil
}

func dwfAnalogOutNodeDataSet(hdwf C.HDWF, channel, node C.int, data []float64) error {
	if len(data) == 0 {
		return nil
	}
	if C.FDwfAnalogOutNodeDataSet(hdwf, channel, node, (*C.double)(unsafe.Pointer(&data[0])), C.int(len(data))) == 0 {
		return lastError()
	}
	return nil
}

func dwfAnalogOutNodeFrequencySet(hdwf C.HDWF, channel, node C.int, freq float64) error {
	if C.FDwfAnalogOutNodeFrequencySet(hdwf, channel, node, C.double(freq)) == 0 {
		return lastError()
	}
	return nil
}

func dwfAnalogOutNodeAmplitudeSet(hdwf C.HDWF, channel, node C.int, amplitude float64) error {
	if C.FDwfAnalogOutNodeAmplitudeSet(hdwf, channel, node, C.double(amplitude)) == 0 {
		return lastError()
	}
	return nil
}

func dwfAnalogOutNodeOffsetSet(hdwf C.HDWF, channel, node C.int, offset float64) error {
	if C.FDwfAnalogOutNodeOffsetSet(hdwf, channel, node, C.double(offset)) == 0 {
		return lastError()
	}
	return nil
}

func dwfAnalogOutNodeSymmetrySet(hdwf C.HDWF, channel, node C.int, symmetry float64) error {
	if C.FDwfAnalogOutNodeSymmetrySet(hdwf, channel, node, C.double(symmetry)) == 0 {
		return lastError()
	}
	return nil
}

func dwfAnalogOutRunSet(hdwf C.HDWF, channel C.int, runTime float64) error {
	if C.FDwfAnalogOutRunSet(hdwf, channel, C.double(runTime)) == 0 {
		return lastError()
	}
	return nil
}

func dwfAnalogOutWaitSet(hdwf C.HDWF, channel C.int, wait float64) error {
	if C.FDwfAnalogOutWaitSet(hdwf, channel, C.double(wait)) == 0 {
		return lastError()
	}
	return nil
}

func dwfAnalogOutRepeatSet(hdwf C.HDWF, channel C.int, repeat int) error {
	if C.FDwfAnalogOutRepeatSet(hdwf, channel, C.int(repeat)) == 0 {
		return lastError()
	}
	return nil
}

func dwfAnalogOutConfigure(hdwf C.HDWF, channel C.int, start bool) error {
	var s C.int
	if start {
		s = 1
	}
	if C.FDwfAnalogOutConfigure(hdwf, channel, s) == 0 {
		return lastError()
	}
	return nil
}

func dwfAnalogOutReset(hdwf C.HDWF, channel C.int) error {
	if C.FDwfAnalogOutReset(hdwf, channel) == 0 {
		return lastError()
	}
	return nil
}

// --- Analog IO (Supplies, DMM, Temperature) ---

func dwfAnalogIOChannelCount(hdwf C.HDWF) (int, error) {
	var count C.int
	if C.FDwfAnalogIOChannelCount(hdwf, &count) == 0 {
		return 0, lastError()
	}
	return int(count), nil
}

func dwfAnalogIOChannelName(hdwf C.HDWF, channel C.int) (string, string, error) {
	var name, label [256]C.char
	if C.FDwfAnalogIOChannelName(hdwf, channel, &name[0], &label[0]) == 0 {
		return "", "", lastError()
	}
	return C.GoString(&name[0]), C.GoString(&label[0]), nil
}

func dwfAnalogIOChannelInfo(hdwf C.HDWF, channel C.int) (int, error) {
	var nodeCount C.int
	if C.FDwfAnalogIOChannelInfo(hdwf, channel, &nodeCount) == 0 {
		return 0, lastError()
	}
	return int(nodeCount), nil
}

func dwfAnalogIOChannelNodeName(hdwf C.HDWF, channel, node C.int) (string, string, error) {
	var name, unit [256]C.char
	if C.FDwfAnalogIOChannelNodeName(hdwf, channel, node, &name[0], &unit[0]) == 0 {
		return "", "", lastError()
	}
	return C.GoString(&name[0]), C.GoString(&unit[0]), nil
}

func dwfAnalogIOChannelNodeSet(hdwf C.HDWF, channel, node C.int, value float64) error {
	if C.FDwfAnalogIOChannelNodeSet(hdwf, channel, node, C.double(value)) == 0 {
		return lastError()
	}
	return nil
}

func dwfAnalogIOChannelNodeGet(hdwf C.HDWF, channel, node C.int) (float64, error) {
	var value C.double
	if C.FDwfAnalogIOChannelNodeGet(hdwf, channel, node, &value) == 0 {
		return 0, lastError()
	}
	return float64(value), nil
}

func dwfAnalogIOChannelNodeStatus(hdwf C.HDWF, channel, node C.int) (float64, error) {
	var value C.double
	if C.FDwfAnalogIOChannelNodeStatus(hdwf, channel, node, &value) == 0 {
		return 0, lastError()
	}
	return float64(value), nil
}

func dwfAnalogIOStatus(hdwf C.HDWF) error {
	if C.FDwfAnalogIOStatus(hdwf) == 0 {
		return lastError()
	}
	return nil
}

func dwfAnalogIOEnableSet(hdwf C.HDWF, enable bool) error {
	var e C.int
	if enable {
		e = 1
	}
	if C.FDwfAnalogIOEnableSet(hdwf, e) == 0 {
		return lastError()
	}
	return nil
}

func dwfAnalogIOReset(hdwf C.HDWF) error {
	if C.FDwfAnalogIOReset(hdwf) == 0 {
		return lastError()
	}
	return nil
}

// --- Digital Input (Logic Analyzer) ---

func dwfDigitalInBitsInfo(hdwf C.HDWF) (int, error) {
	var bits C.int
	if C.FDwfDigitalInBitsInfo(hdwf, &bits) == 0 {
		return 0, lastError()
	}
	return int(bits), nil
}

func dwfDigitalInBufferSizeInfo(hdwf C.HDWF) (int, error) {
	var maxSize C.int
	if C.FDwfDigitalInBufferSizeInfo(hdwf, &maxSize) == 0 {
		return 0, lastError()
	}
	return int(maxSize), nil
}

func dwfDigitalInInternalClockInfo(hdwf C.HDWF) (float64, error) {
	var freq C.double
	if C.FDwfDigitalInInternalClockInfo(hdwf, &freq) == 0 {
		return 0, lastError()
	}
	return float64(freq), nil
}

func dwfDigitalInDividerSet(hdwf C.HDWF, divider int) error {
	if C.FDwfDigitalInDividerSet(hdwf, C.uint(divider)) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalInSampleFormatSet(hdwf C.HDWF, bits int) error {
	if C.FDwfDigitalInSampleFormatSet(hdwf, C.int(bits)) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalInBufferSizeSet(hdwf C.HDWF, size int) error {
	if C.FDwfDigitalInBufferSizeSet(hdwf, C.int(size)) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalInConfigure(hdwf C.HDWF, reconfigure, start bool) error {
	var r, s C.int
	if reconfigure {
		r = 1
	}
	if start {
		s = 1
	}
	if C.FDwfDigitalInConfigure(hdwf, r, s) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalInStatus(hdwf C.HDWF, readData bool) (byte, error) {
	var rd C.int
	if readData {
		rd = 1
	}
	var status C.DwfState
	if C.FDwfDigitalInStatus(hdwf, rd, &status) == 0 {
		return 0, lastError()
	}
	return byte(status), nil
}

func dwfDigitalInStatusData(hdwf C.HDWF, buf []uint16) error {
	if C.FDwfDigitalInStatusData(hdwf, unsafe.Pointer(&buf[0]), C.int(2*len(buf))) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalInReset(hdwf C.HDWF) error {
	if C.FDwfDigitalInReset(hdwf) == 0 {
		return lastError()
	}
	return nil
}

// --- Logic Trigger ---

func dwfDigitalInTriggerSourceSet(hdwf C.HDWF, src C.TRIGSRC) error {
	if C.FDwfDigitalInTriggerSourceSet(hdwf, src) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalInTriggerPositionSet(hdwf C.HDWF, position int) error {
	if C.FDwfDigitalInTriggerPositionSet(hdwf, C.uint(position)) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalInTriggerPrefillSet(hdwf C.HDWF, prefill int) error {
	if C.FDwfDigitalInTriggerPrefillSet(hdwf, C.uint(prefill)) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalInTriggerSet(hdwf C.HDWF, levelLow, levelHigh, edgeRise, edgeFall C.uint) error {
	if C.FDwfDigitalInTriggerSet(hdwf, levelLow, levelHigh, edgeRise, edgeFall) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalInTriggerResetSet(hdwf C.HDWF, levelLow, levelHigh, edgeRise, edgeFall C.uint) error {
	if C.FDwfDigitalInTriggerResetSet(hdwf, levelLow, levelHigh, edgeRise, edgeFall) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalInTriggerAutoTimeoutSet(hdwf C.HDWF, timeout float64) error {
	if C.FDwfDigitalInTriggerAutoTimeoutSet(hdwf, C.double(timeout)) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalInTriggerLengthSet(hdwf C.HDWF, min, max float64, sync C.int) error {
	if C.FDwfDigitalInTriggerLengthSet(hdwf, C.double(min), C.double(max), sync) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalInTriggerCountSet(hdwf C.HDWF, count, restart C.int) error {
	if C.FDwfDigitalInTriggerCountSet(hdwf, count, restart) == 0 {
		return lastError()
	}
	return nil
}

// --- Digital Output (Pattern Generator) ---

func dwfDigitalOutCount(hdwf C.HDWF) (int, error) {
	var count C.int
	if C.FDwfDigitalOutCount(hdwf, &count) == 0 {
		return 0, lastError()
	}
	return int(count), nil
}

func dwfDigitalOutInternalClockInfo(hdwf C.HDWF) (float64, error) {
	var freq C.double
	if C.FDwfDigitalOutInternalClockInfo(hdwf, &freq) == 0 {
		return 0, lastError()
	}
	return float64(freq), nil
}

func dwfDigitalOutEnableSet(hdwf C.HDWF, channel C.int, enable bool) error {
	var e C.int
	if enable {
		e = 1
	}
	if C.FDwfDigitalOutEnableSet(hdwf, channel, e) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalOutTypeSet(hdwf C.HDWF, channel C.int, outType C.DwfDigitalOutType) error {
	if C.FDwfDigitalOutTypeSet(hdwf, channel, outType) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalOutDividerSet(hdwf C.HDWF, channel C.int, divider int) error {
	if C.FDwfDigitalOutDividerSet(hdwf, channel, C.uint(divider)) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalOutIdleSet(hdwf C.HDWF, channel C.int, idle C.DwfDigitalOutIdle) error {
	if C.FDwfDigitalOutIdleSet(hdwf, channel, idle) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalOutRunSet(hdwf C.HDWF, runTime float64) error {
	if C.FDwfDigitalOutRunSet(hdwf, C.double(runTime)) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalOutWaitSet(hdwf C.HDWF, wait float64) error {
	if C.FDwfDigitalOutWaitSet(hdwf, C.double(wait)) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalOutRepeatSet(hdwf C.HDWF, repeat int) error {
	if C.FDwfDigitalOutRepeatSet(hdwf, C.uint(repeat)) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalOutCounterSet(hdwf C.HDWF, channel C.int, low, high int) error {
	if C.FDwfDigitalOutCounterSet(hdwf, channel, C.uint(low), C.uint(high)) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalOutDataSet(hdwf C.HDWF, channel C.int, data []uint16) error {
	if len(data) == 0 {
		return nil
	}
	if C.FDwfDigitalOutDataSet(hdwf, channel, unsafe.Pointer(&data[0]), C.uint(len(data))) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalOutConfigure(hdwf C.HDWF, start bool) error {
	var s C.int
	if start {
		s = 1
	}
	if C.FDwfDigitalOutConfigure(hdwf, s) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalOutReset(hdwf C.HDWF) error {
	if C.FDwfDigitalOutReset(hdwf) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalOutRepeatTriggerSet(hdwf C.HDWF, enable bool) error {
	var e C.int
	if enable {
		e = 1
	}
	if C.FDwfDigitalOutRepeatTriggerSet(hdwf, e) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalOutTriggerSourceSet(hdwf C.HDWF, src C.TRIGSRC) error {
	if C.FDwfDigitalOutTriggerSourceSet(hdwf, src) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalOutTriggerSlopeSet(hdwf C.HDWF, slope C.DwfTriggerSlope) error {
	if C.FDwfDigitalOutTriggerSlopeSet(hdwf, slope) == 0 {
		return lastError()
	}
	return nil
}

// --- Digital IO (Static I/O) ---

func dwfDigitalIOOutputEnableGet(hdwf C.HDWF) (uint32, error) {
	var mask C.uint
	if C.FDwfDigitalIOOutputEnableGet(hdwf, &mask) == 0 {
		return 0, lastError()
	}
	return uint32(mask), nil
}

func dwfDigitalIOOutputEnableSet(hdwf C.HDWF, mask uint32) error {
	if C.FDwfDigitalIOOutputEnableSet(hdwf, C.uint(mask)) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalIOOutputGet(hdwf C.HDWF) (uint32, error) {
	var mask C.uint
	if C.FDwfDigitalIOOutputGet(hdwf, &mask) == 0 {
		return 0, lastError()
	}
	return uint32(mask), nil
}

func dwfDigitalIOOutputSet(hdwf C.HDWF, mask uint32) error {
	if C.FDwfDigitalIOOutputSet(hdwf, C.uint(mask)) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalIOStatus(hdwf C.HDWF) error {
	if C.FDwfDigitalIOStatus(hdwf) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalIOInputStatus(hdwf C.HDWF) (uint32, error) {
	var data C.uint
	if C.FDwfDigitalIOInputStatus(hdwf, &data) == 0 {
		return 0, lastError()
	}
	return uint32(data), nil
}

func dwfDigitalIOReset(hdwf C.HDWF) error {
	if C.FDwfDigitalIOReset(hdwf) == 0 {
		return lastError()
	}
	return nil
}

// --- UART ---

func dwfDigitalUartRateSet(hdwf C.HDWF, rate float64) error {
	if C.FDwfDigitalUartRateSet(hdwf, C.double(rate)) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalUartTxSet(hdwf C.HDWF, channel C.int) error {
	if C.FDwfDigitalUartTxSet(hdwf, channel) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalUartRxSet(hdwf C.HDWF, channel C.int) error {
	if C.FDwfDigitalUartRxSet(hdwf, channel) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalUartBitsSet(hdwf C.HDWF, bits C.int) error {
	if C.FDwfDigitalUartBitsSet(hdwf, bits) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalUartParitySet(hdwf C.HDWF, parity C.int) error {
	if C.FDwfDigitalUartParitySet(hdwf, parity) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalUartStopSet(hdwf C.HDWF, stop float64) error {
	if C.FDwfDigitalUartStopSet(hdwf, C.double(stop)) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalUartTx(hdwf C.HDWF, data []byte) error {
	if len(data) == 0 {
		if C.FDwfDigitalUartTx(hdwf, nil, C.int(0)) == 0 {
			return lastError()
		}
		return nil
	}
	if C.FDwfDigitalUartTx(hdwf, (*C.char)(unsafe.Pointer(&data[0])), C.int(len(data))) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalUartRx(hdwf C.HDWF, bufSize int) ([]byte, int, error) {
	buf := make([]C.char, bufSize)
	var count, parity C.int
	if C.FDwfDigitalUartRx(hdwf, &buf[0], C.int(bufSize), &count, &parity) == 0 {
		return nil, int(parity), lastError()
	}
	result := make([]byte, int(count))
	for i := 0; i < int(count); i++ {
		result[i] = byte(buf[i])
	}
	return result, int(parity), nil
}

func dwfDigitalUartReset(hdwf C.HDWF) error {
	if C.FDwfDigitalUartReset(hdwf) == 0 {
		return lastError()
	}
	return nil
}

// --- SPI ---

func dwfDigitalSpiFrequencySet(hdwf C.HDWF, freq float64) error {
	if C.FDwfDigitalSpiFrequencySet(hdwf, C.double(freq)) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalSpiClockSet(hdwf C.HDWF, channel C.int) error {
	if C.FDwfDigitalSpiClockSet(hdwf, channel) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalSpiDataSet(hdwf C.HDWF, idx, channel C.int) error {
	if C.FDwfDigitalSpiDataSet(hdwf, idx, channel) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalSpiIdleSet(hdwf C.HDWF, idx C.int, idle C.DwfDigitalOutIdle) error {
	if C.FDwfDigitalSpiIdleSet(hdwf, idx, idle) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalSpiModeSet(hdwf C.HDWF, mode C.int) error {
	if C.FDwfDigitalSpiModeSet(hdwf, mode) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalSpiOrderSet(hdwf C.HDWF, order C.int) error {
	if C.FDwfDigitalSpiOrderSet(hdwf, order) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalSpiSelect(hdwf C.HDWF, cs, level C.int) error {
	if C.FDwfDigitalSpiSelect(hdwf, cs, level) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalSpiWriteOne(hdwf C.HDWF, csMode C.int, bits C.int, data C.uint) error {
	if C.FDwfDigitalSpiWriteOne(hdwf, csMode, bits, data) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalSpiRead(hdwf C.HDWF, csMode, bits C.int, buf []byte) error {
	if C.FDwfDigitalSpiRead(hdwf, csMode, bits, (*C.uchar)(unsafe.Pointer(&buf[0])), C.int(len(buf))) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalSpiWrite(hdwf C.HDWF, csMode, bits C.int, data []byte) error {
	if C.FDwfDigitalSpiWrite(hdwf, csMode, bits, (*C.uchar)(unsafe.Pointer(&data[0])), C.int(len(data))) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalSpiWriteRead(hdwf C.HDWF, csMode, bits C.int, txData []byte, rxBuf []byte) error {
	if C.FDwfDigitalSpiWriteRead(hdwf, csMode, bits,
		(*C.uchar)(unsafe.Pointer(&txData[0])), C.int(len(txData)),
		(*C.uchar)(unsafe.Pointer(&rxBuf[0])), C.int(len(rxBuf))) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalSpiReset(hdwf C.HDWF) error {
	if C.FDwfDigitalSpiReset(hdwf) == 0 {
		return lastError()
	}
	return nil
}

// --- I2C ---

func dwfDigitalI2cReset(hdwf C.HDWF) error {
	if C.FDwfDigitalI2cReset(hdwf) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalI2cStretchSet(hdwf C.HDWF, enable bool) error {
	var e C.int
	if enable {
		e = 1
	}
	if C.FDwfDigitalI2cStretchSet(hdwf, e) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalI2cRateSet(hdwf C.HDWF, rate float64) error {
	if C.FDwfDigitalI2cRateSet(hdwf, C.double(rate)) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalI2cSclSet(hdwf C.HDWF, channel C.int) error {
	if C.FDwfDigitalI2cSclSet(hdwf, channel) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalI2cSdaSet(hdwf C.HDWF, channel C.int) error {
	if C.FDwfDigitalI2cSdaSet(hdwf, channel) == 0 {
		return lastError()
	}
	return nil
}

func dwfDigitalI2cClear(hdwf C.HDWF) (int, error) {
	var nak C.int
	if C.FDwfDigitalI2cClear(hdwf, &nak) == 0 {
		return 0, lastError()
	}
	return int(nak), nil
}

func dwfDigitalI2cRead(hdwf C.HDWF, address C.int, buf []byte) (int, error) {
	var nak C.int
	if C.FDwfDigitalI2cRead(hdwf, C.uchar(address), (*C.uchar)(unsafe.Pointer(&buf[0])), C.int(len(buf)), &nak) == 0 {
		return int(nak), lastError()
	}
	return int(nak), nil
}

func dwfDigitalI2cWrite(hdwf C.HDWF, address C.int, data []byte) (int, error) {
	var nak C.int
	var dataPtr *C.uchar
	dataLen := C.int(len(data))
	if len(data) > 0 {
		dataPtr = (*C.uchar)(unsafe.Pointer(&data[0]))
	}
	if C.FDwfDigitalI2cWrite(hdwf, C.uchar(address), dataPtr, dataLen, &nak) == 0 {
		return int(nak), lastError()
	}
	return int(nak), nil
}

func dwfDigitalI2cWriteRead(hdwf C.HDWF, address C.int, txData, rxBuf []byte) (int, error) {
	var nak C.int
	if C.FDwfDigitalI2cWriteRead(hdwf, C.uchar(address),
		(*C.uchar)(unsafe.Pointer(&txData[0])), C.int(len(txData)),
		(*C.uchar)(unsafe.Pointer(&rxBuf[0])), C.int(len(rxBuf)),
		&nak) == 0 {
		return int(nak), lastError()
	}
	return int(nak), nil
}

// ============================================================
// Go-level type aliases and constant wrappers
// These allow device.go to work without importing "C" directly.
// ============================================================

// DevHandle is the Go-level alias for the native device handle.
type DevHandle = C.HDWF

// Go-level constants wrapping C SDK constants
var (
	cEnumfilterAll        = C.int(C.enumfilterAll)
	cDevidDiscovery       = C.int(C.devidDiscovery)
	cDevidDiscovery2      = C.int(C.devidDiscovery2)
	cDevidDDiscovery      = C.int(C.devidDDiscovery)
	cDevidADP3X50         = C.int(C.devidADP3X50)
	cDevidADP5250         = C.int(C.devidADP5250)
	cFilterDecimate       = C.int(C.filterDecimate)
	cTrigsrcNone          = C.TRIGSRC(C.trigsrcNone)
	cTrigsrcDetectorDigIn = C.TRIGSRC(C.trigsrcDetectorDigitalIn)
	cTrigtypeEdge         = C.int(C.trigtypeEdge)
	cDwfTriggerSlopeRise  = C.DwfTriggerSlope(C.DwfTriggerSlopeRise)
	cDwfTriggerSlopeFall  = C.DwfTriggerSlope(C.DwfTriggerSlopeFall)
	cDwfStateDone         = byte(C.DwfStateDone)
	cAnalogOutNodeCarrier = C.int(C.AnalogOutNodeCarrier)
	cDwfDigitalOutIdleZet = C.DwfDigitalOutIdle(C.DwfDigitalOutIdleZet)

	// DwfEnumConfigInfo constants
	cDECIAnalogInChannelCount   = C.DwfEnumConfigInfo(C.DECIAnalogInChannelCount)
	cDECIAnalogOutChannelCount  = C.DwfEnumConfigInfo(C.DECIAnalogOutChannelCount)
	cDECIAnalogIOChannelCount   = C.DwfEnumConfigInfo(C.DECIAnalogIOChannelCount)
	cDECIDigitalInChannelCount  = C.DwfEnumConfigInfo(C.DECIDigitalInChannelCount)
	cDECIDigitalOutChannelCount = C.DwfEnumConfigInfo(C.DECIDigitalOutChannelCount)
	cDECIDigitalIOChannelCount  = C.DwfEnumConfigInfo(C.DECIDigitalIOChannelCount)
	cDECIAnalogInBufferSize     = C.DwfEnumConfigInfo(C.DECIAnalogInBufferSize)
	cDECIAnalogOutBufferSize    = C.DwfEnumConfigInfo(C.DECIAnalogOutBufferSize)
	cDECIDigitalInBufferSize    = C.DwfEnumConfigInfo(C.DECIDigitalInBufferSize)
	cDECIDigitalOutBufferSize   = C.DwfEnumConfigInfo(C.DECIDigitalOutBufferSize)
)

// cInt converts Go int to C.int for use in device.go
func cInt(v int) C.int { return C.int(v) }

// cUint converts Go uint32 to C.uint
func cUint(v uint32) C.uint { return C.uint(v) }

// cFunc converts Go WavegenFunc to C.FUNC
func cFunc(v WavegenFunc) C.FUNC { return C.FUNC(v) }

// cTrigSrc converts Go TriggerSource to C.TRIGSRC
func cTrigSrc(v TriggerSource) C.TRIGSRC { return C.TRIGSRC(v) }

// cDigitalOutType converts Go DigitalOutType to C.DwfDigitalOutType
func cDigitalOutType(v DigitalOutType) C.DwfDigitalOutType { return C.DwfDigitalOutType(v) }

// cDigitalOutIdle converts Go DigitalOutIdle to C.DwfDigitalOutIdle
func cDigitalOutIdle(v DigitalOutIdle) C.DwfDigitalOutIdle { return C.DwfDigitalOutIdle(v) }
