# Digilent Discovery MCP Server

An [MCP](https://modelcontextprotocol.io) server that exposes [Digilent DWF SDK](https://digilent.com/reference/software/waveforms/waveforms-sdk/start) instruments as tools for LLM-based agents.

Connect an Analog Discovery 2, Digital Discovery, or other Digilent device and let AI agents directly control oscilloscopes, waveform generators, logic analyzers, power supplies, DMMs, and protocol interfaces (UART, SPI, I2C).

## Prerequisites

- **Go 1.25+**
- **Digilent DWF SDK** — install the [Adept Runtime](https://digilent.com/reference/software/adept/start) and [WaveForms](https://digilent.com/reference/software/waveforms/waveforms-3/start) so that `libdwf` is available on your system
- A supported Digilent device connected via USB

### Supported Devices

| Device | Filter ID |
|---|---|
| Analog Discovery | `devidDiscovery` |
| Analog Discovery 2 / Studio | `devidDiscovery2` |
| Digital Discovery | `devidDDiscovery` |
| Analog Discovery Pro 3X50 | `devidADP3X50` |
| Analog Discovery Pro 5250 | `devidADP5250` |

## Build

```bash
go build -o discovery-mcp .
```

## Usage

### Transport Modes

```bash
# stdio (default) — for MCP clients like Claude Desktop
./discovery-mcp

# SSE — Server-Sent Events on HTTP
./discovery-mcp --transport sse --host localhost --port 8080

# Streamable HTTP
./discovery-mcp --transport http --host localhost --port 8080
```

### Device Check

Verify device connectivity and inspect available configurations:

```bash
./discovery-mcp --check
```

Example output:

```
Enumerated Devices: 1
  0) Analog Discovery 2 (Discovery2)      SN:210321A197E9       available

Only one device connected, opening it...

Device Info:
  Name:                Analog Discovery 2
  Serial Number:       SN:210321A1B2C3
  SDK Version:         3.20.1
  Analog In Channels:  2
  Analog Out Channels: 2
  Digital In Channels: 16
  Digital Out Channels:16
  Max Buffer Size:     8192
  ADC Resolution:      14 bits
  Board Temperature:   38.2 °C

Available Configurations (8):
  --------------------------------------------------------------------------------------
  Config  AI-Ch  AO-Ch  IO-Ch  DI-Ch  DO-Ch  DIO    AI-Buf   AO-Buf   DI-Buf   DO-Buf  
  --------------------------------------------------------------------------------------
    0       2      2      2      16    16    16     8192     4096     4096     1024    
    1       2      2      2      16    0     16     16384    1024     1024     0    
  ...
```

### CLI Flags

| Flag | Default | Description |
|---|---|---|
| `--transport` | `stdio` | Transport mode: `stdio`, `sse`, or `http` |
| `--host` | `0.0.0.0` | Listen address for SSE/HTTP modes |
| `--port` | `8080` | Listen port for SSE/HTTP modes |
| `--check` | `false` | Print device info and exit |

### MCP Client Configuration

Add this to your MCP client config (e.g. Claude Desktop `claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "discovery": {
      "command": "/path/to/discovery-mcp"
    }
  }
}
```

For SSE mode, point your client to `http://localhost:8080/sse`.

## MCP Tools Reference

### Device

#### `discovery_device_open`

Open a connection to a Digilent DWF device. Must be called before using any instrument.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `device` | string | No | Device name filter. Empty string connects to the first available device. Examples: `"Analog Discovery 2"`, `"Digital Discovery"` |
| `config` | number | No | Device configuration index. `0` = default. Use `--check` to see available configurations and their resource allocations |

**Returns:** Device info including name, serial number, channel counts, buffer sizes, and ADC resolution.

#### `discovery_device_close`

Close the connection to the device and free all resources. No parameters.

#### `discovery_device_temperature`

Read the on-board temperature sensor. No parameters.

**Returns:** Temperature in °C. Not all devices have a temperature sensor.

---

### Oscilloscope

#### `discovery_scope_open`

Initialize the oscilloscope instrument with acquisition parameters.

| Parameter | Type | Required | Default | Description |
|---|---|---|---|---|
| `sampling_frequency` | number | No | 20 MHz | Sampling rate in Hz |
| `buffer_size` | number | No | max | Number of samples per acquisition. `0` = device maximum |
| `offset_voltage` | number | No | 0 | DC offset in Volts |
| `amplitude_range` | number | No | 5 | Input range in Volts (e.g. `5` for ±5 V) |

#### `discovery_scope_measure`

Take a single instantaneous voltage reading.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `channel` | number | **Yes** | Oscilloscope channel (1-based) |

**Returns:** Voltage in Volts.

#### `discovery_scope_trigger`

Configure the oscilloscope trigger for edge-triggered acquisition.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `enable` | boolean | No | Enable or disable the trigger |
| `source` | number | No | Trigger source: `0`=none, `2`=analog in detector, `3`=digital in detector, `11–14`=external |
| `channel` | number | No | Trigger channel (1-based for analog) |
| `timeout` | number | No | Auto-trigger timeout in seconds. `0` disables auto-trigger |
| `edge_rising` | boolean | No | `true` = rising edge, `false` = falling edge |
| `level` | number | No | Trigger level in Volts |

#### `discovery_scope_record`

Capture a full buffer of analog samples. Configure the oscilloscope and optionally set a trigger before calling this.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `channel` | number | **Yes** | Oscilloscope channel (1-based) |

**Returns:** JSON with sample count, min/max values, and the full data array.

#### `discovery_scope_close`

Reset the oscilloscope instrument. No parameters.

---

### Wavegen

#### `discovery_wavegen_generate`

Generate an analog waveform on a wavegen channel.

| Parameter | Type | Required | Default | Description |
|---|---|---|---|---|
| `channel` | number | **Yes** | — | Output channel (1 or 2) |
| `function` | number | **Yes** | — | Waveform type: `0`=DC, `1`=sine, `2`=square, `3`=triangle, `4`=ramp up, `5`=ramp down, `6`=noise, `7`=pulse, `8`=trapezium, `9`=sine power, `30`=custom |
| `frequency` | number | No | 0 | Frequency in Hz |
| `amplitude` | number | No | 0 | Peak amplitude in Volts |
| `offset` | number | No | 0 | DC offset in Volts |
| `symmetry` | number | No | 0 | Symmetry in % (0–100) |
| `wait` | number | No | 0 | Wait time before start in seconds |
| `run_time` | number | No | 0 | Duration in seconds. `0` = continuous |
| `repeat` | number | No | 0 | Repeat count. `0` = infinite |

#### `discovery_wavegen_enable` / `discovery_wavegen_disable`

Enable or disable output on a wavegen channel.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `channel` | number | **Yes** | Channel (1-based) |

#### `discovery_wavegen_close`

Reset a wavegen channel.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `channel` | number | **Yes** | Channel (1-based) |

---

### Power Supplies

#### `discovery_supplies_switch`

Configure and enable/disable the onboard power supply rails (V+, V−, VDD/6V).

| Parameter | Type | Required | Description |
|---|---|---|---|
| `master_state` | boolean | No | Master enable for all supplies |
| `positive_state` | boolean | No | Enable the positive (V+) rail |
| `positive_voltage` | number | No | Positive voltage in V |
| `positive_current` | number | No | Positive current limit in A |
| `negative_state` | boolean | No | Enable the negative (V−) rail |
| `negative_voltage` | number | No | Negative voltage in V |
| `negative_current` | number | No | Negative current limit in A |
| `state` | boolean | No | Enable the digital/6V (VDD) rail |
| `voltage` | number | No | Digital/6V rail voltage in V |
| `current` | number | No | Digital/6V rail current limit in A |

#### `discovery_supplies_close`

Reset all power supplies. No parameters.

---

### Digital Multimeter

> **Note:** DMM is only available on certain devices (e.g. Analog Discovery Pro).

#### `discovery_dmm_open`

Initialize the DMM instrument. No parameters.

#### `discovery_dmm_measure`

Perform a DMM measurement.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `mode` | number | **Yes** | Measurement mode: `0`=AC voltage, `1`=DC voltage, `2`=AC current, `3`=DC current, `4`=resistance, `5`=continuity, `6`=diode, `7`=temperature, `8`=AC low current, `9`=DC low current, `10`=AC high current, `11`=DC high current |
| `range` | number | No | Measurement range. `0` = auto-range |
| `high_impedance` | boolean | No | Use 10 GΩ input impedance (vs 10 MΩ) for DC voltage |

**Returns:** Measured value with appropriate unit.

#### `discovery_dmm_close`

Reset the DMM. No parameters.

---

### Logic Analyzer

#### `discovery_logic_open`

Initialize the logic analyzer.

| Parameter | Type | Required | Default | Description |
|---|---|---|---|---|
| `sampling_frequency` | number | No | 100 MHz | Sampling rate in Hz |
| `buffer_size` | number | No | max | Buffer size. `0` = device maximum |

#### `discovery_logic_trigger`

Configure the logic analyzer trigger.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `enable` | boolean | No | Enable or disable the trigger |
| `channel` | number | No | DIO line number to trigger on |
| `position` | number | No | Prefill size (samples before trigger) |
| `timeout` | number | No | Auto-trigger timeout in seconds |
| `rising_edge` | boolean | No | `true` = rising edge, `false` = falling edge |
| `length_min` | number | No | Minimum trigger sequence duration in seconds |
| `length_max` | number | No | Maximum trigger sequence duration in seconds |
| `count` | number | No | Trigger event count |

#### `discovery_logic_record`

Capture digital samples from a DIO channel.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `channel` | number | **Yes** | DIO line number |

**Returns:** JSON with sample count and the data array.

#### `discovery_logic_close`

Reset the logic analyzer. No parameters.

---

### Pattern Generator

#### `discovery_pattern_generate`

Generate a digital pattern on a DIO line.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `channel` | number | **Yes** | DIO line number |
| `function` | number | **Yes** | Output type: `0`=pulse, `1`=custom, `2`=random |
| `frequency` | number | **Yes** | Frequency in Hz |
| `duty_cycle` | number | No | Duty cycle % (for pulse mode) |
| `wait` | number | No | Wait time before start in seconds |
| `repeat` | number | No | Repeat count. `0` = infinite |
| `run_time` | number | No | Duration in seconds. `0`=infinite, `-1`=auto |

#### `discovery_pattern_enable` / `discovery_pattern_disable`

Enable or disable a digital output.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `channel` | number | **Yes** | DIO line number |

#### `discovery_pattern_close`

Reset the pattern generator. No parameters.

---

### Static I/O

#### `discovery_static_set_mode`

Configure a DIO line as input or output.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `channel` | number | **Yes** | DIO channel number |
| `output` | boolean | **Yes** | `true` = output, `false` = input |

#### `discovery_static_get_state`

Read the current state of a DIO line.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `channel` | number | **Yes** | DIO channel number |

**Returns:** `true` (HIGH) or `false` (LOW).

#### `discovery_static_set_state`

Set a DIO line high or low.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `channel` | number | **Yes** | DIO channel number |
| `value` | boolean | **Yes** | `true` = HIGH, `false` = LOW |

#### `discovery_static_close`

Reset static I/O. No parameters.

---

### UART

#### `discovery_uart_open`

Initialize UART communication on DIO lines.

| Parameter | Type | Required | Default | Description |
|---|---|---|---|---|
| `rx` | number | **Yes** | — | DIO line for receiving |
| `tx` | number | **Yes** | — | DIO line for transmitting |
| `baud_rate` | number | No | 9600 | Baud rate in bits/s |
| `parity` | number | No | 0 | `0`=none, `1`=odd, `2`=even |
| `data_bits` | number | No | 8 | Number of data bits |
| `stop_bits` | number | No | 1 | Number of stop bits |

#### `discovery_uart_read`

Read available data from the UART RX buffer. No parameters.

**Returns:** Received bytes as a string.

#### `discovery_uart_write`

Send data through UART TX.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `data` | string | **Yes** | Text data to send |

#### `discovery_uart_close`

Reset the UART interface. No parameters.

---

### SPI

#### `discovery_spi_open`

Initialize SPI communication.

| Parameter | Type | Required | Default | Description |
|---|---|---|---|---|
| `cs` | number | **Yes** | — | DIO line for chip select |
| `sck` | number | **Yes** | — | DIO line for serial clock |
| `miso` | number | No | -1 | DIO line for MISO. `-1` to skip |
| `mosi` | number | No | -1 | DIO line for MOSI. `-1` to skip |
| `clock_frequency` | number | No | 1 MHz | Clock frequency in Hz |
| `mode` | number | No | 0 | SPI mode (0–3) |
| `msb_first` | boolean | No | true | `true` = MSB first, `false` = LSB first |

#### `discovery_spi_read`

Read bytes from SPI.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `count` | number | **Yes** | Number of bytes to read |
| `cs` | number | **Yes** | Chip select DIO line |

#### `discovery_spi_write`

Write data through SPI.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `data` | string | **Yes** | Hex string to send (e.g. `"FF01A2"`) |
| `cs` | number | **Yes** | Chip select DIO line |

#### `discovery_spi_close`

Reset the SPI interface. No parameters.

---

### I2C

#### `discovery_i2c_open`

Initialize I2C communication.

| Parameter | Type | Required | Default | Description |
|---|---|---|---|---|
| `sda` | number | **Yes** | — | DIO line for data |
| `scl` | number | **Yes** | — | DIO line for clock |
| `clock_rate` | number | No | 100 kHz | Clock rate in Hz |
| `stretching` | boolean | No | false | Enable clock stretching |

#### `discovery_i2c_read`

Read bytes from an I2C device.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `count` | number | **Yes** | Number of bytes to read |
| `address` | number | **Yes** | 7-bit I2C device address |

#### `discovery_i2c_write`

Write data to an I2C device.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `data` | string | **Yes** | Hex string to send (e.g. `"FF01A2"`) |
| `address` | number | **Yes** | 7-bit I2C device address |

#### `discovery_i2c_close`

Reset the I2C interface. No parameters.

---

## Examples

### Measure a DC Voltage

Open the device and take a quick voltage reading from oscilloscope channel 1:

```
1. discovery_device_open  → { "device": "", "config": 0 }
2. discovery_scope_open   → { "sampling_frequency": 20000000, "amplitude_range": 5 }
3. discovery_scope_measure → { "channel": 1 }
   ← "3.312500 V on channel 1"
4. discovery_scope_close
5. discovery_device_close
```

### Record a Triggered Waveform

Capture a 1 kHz signal with rising-edge trigger at 1.5 V:

```
1. discovery_device_open   → { "config": 0 }
2. discovery_scope_open    → { "sampling_frequency": 1000000, "buffer_size": 8192 }
3. discovery_scope_trigger → { "enable": true, "source": 2, "channel": 1,
                               "edge_rising": true, "level": 1.5, "timeout": 5 }
4. discovery_scope_record  → { "channel": 1 }
   ← { "samples": 8192, "min": -1.65, "max": 1.65, "data": [...] }
5. discovery_scope_close
6. discovery_device_close
```

### Generate a 1 kHz Sine Wave

Output a 1 kHz sine wave at 2 V amplitude on wavegen channel 1:

```
1. discovery_device_open       → { "config": 0 }
2. discovery_wavegen_generate  → { "channel": 1, "function": 1,
                                   "frequency": 1000, "amplitude": 2 }
   ← "Waveform started on channel 1"
   ... (signal is now outputting) ...
3. discovery_wavegen_disable   → { "channel": 1 }
4. discovery_wavegen_close     → { "channel": 1 }
5. discovery_device_close
```

### Power Supply Control

Enable +5 V supply with 100 mA limit:

```
1. discovery_device_open     → { "config": 0 }
2. discovery_supplies_switch → { "master_state": true,
                                 "positive_state": true,
                                 "positive_voltage": 5.0,
                                 "positive_current": 0.1 }
   ← "Power supplies configured"
   ... (supply is now active) ...
3. discovery_supplies_close
4. discovery_device_close
```

### Read a Sensor over I2C

Read 2 bytes from a sensor at address `0x48`:

```
1. discovery_device_open → { "config": 0 }
2. discovery_i2c_open    → { "sda": 0, "scl": 1, "clock_rate": 100000 }
3. discovery_i2c_read    → { "count": 2, "address": 72 }
   ← { "data": "1A3F" }
4. discovery_i2c_close
5. discovery_device_close
```

### Blink an LED with Static I/O

Toggle DIO line 0 as a digital output:

```
1. discovery_device_open      → { "config": 0 }
2. discovery_static_set_mode  → { "channel": 0, "output": true }
3. discovery_static_set_state → { "channel": 0, "value": true }
   ← "DIO 0 set HIGH"
4. discovery_static_set_state → { "channel": 0, "value": false }
   ← "DIO 0 set LOW"
5. discovery_static_close
6. discovery_device_close
```

### UART Communication

Send and receive data at 115200 baud:

```
1. discovery_device_open → { "config": 0 }
2. discovery_uart_open   → { "rx": 1, "tx": 0, "baud_rate": 115200 }
3. discovery_uart_write  → { "data": "AT\r\n" }
   ← "Sent 4 bytes"
4. discovery_uart_read
   ← "OK\r\n"
5. discovery_uart_close
6. discovery_device_close
```

## Project Structure

```
├── main.go              # CLI entry point and --check diagnostics
├── server/
│   ├── server.go        # MCP server setup and tool registration
│   ├── handlers.go      # MCP tool handler implementations
│   └── handlers_test.go # Unit tests with mock device
└── dwf/
    ├── interfaces.go    # Go interfaces (Oscilloscope, WavegenDriver, etc.)
    ├── types.go         # Configuration structs and enums
    ├── bindings.go      # CGo bindings to libdwf
    └── device.go        # Concrete device implementation
```

## Testing

```bash
go test ./...
```

Tests use mock device implementations and do not require hardware.

## License

MIT
