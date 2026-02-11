// Digilent Discovery MCP Server — an MCP server that exposes Digilent DWF SDK
// instruments as MCP tools for LLM-based agents.
//
// Usage:
//
//	go run .                          # stdio mode (default)
//	go run . --transport sse          # SSE mode on port 8080
//	go run . --transport http         # Streamable HTTP on port 8080
//	go run . --transport sse --host localhost --port 9090   # custom address
//	go run . --check                  # check device connectivity
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	mcpserver "github.com/mark3labs/mcp-go/server"

	"github.com/molejar/discovery-mcp/dwf"
	"github.com/molejar/discovery-mcp/server"
)

func main() {
	transport := flag.String("transport", "stdio", "Transport mode: stdio, sse, or http")
	port := flag.String("port", "8080", "Listen port for sse/http transport")
	host := flag.String("host", "0.0.0.0", "Listen host/address for sse/http transport")
	check := flag.Bool("check", false, "Check device connectivity and print device info, then exit")
	flag.Parse()

	if *check {
		checkDevice()
		return
	}

	s := server.New()

	switch *transport {
	case "stdio":
		log.Println("Digilent Discovery MCP Server starting (stdio mode)...")
		if err := mcpserver.ServeStdio(s.MCPServer()); err != nil {
			log.Fatalf("Server error: %v", err)
		}

	case "sse":
		sseServer := mcpserver.NewSSEServer(s.MCPServer(),
			mcpserver.WithBaseURL(fmt.Sprintf("http://%s:%s", *host, *port)),
		)
		log.Printf("Digilent Discovery MCP Server starting (SSE mode) on %s ...", *port)
		log.Printf("  SSE endpoint:     http://%s:%s/sse", *host, *port)
		log.Printf("  Message endpoint: http://%s:%s/message", *host, *port)

		// graceful shutdown
		go func() {
			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
			<-sigCh
			log.Println("Shutting down SSE server...")
			if err := sseServer.Shutdown(context.Background()); err != nil {
				log.Printf("Shutdown error: %v", err)
			}
		}()

		address := fmt.Sprintf("%s:%s", *host, *port)

		if err := sseServer.Start(address); err != nil {
			log.Fatalf("SSE server error: %v", err)
		}

	case "http":
		httpServer := mcpserver.NewStreamableHTTPServer(s.MCPServer())
		log.Printf("Digilent Discovery MCP Server starting (Streamable HTTP mode) on %s ...", *port)
		log.Printf("  Endpoint: http://%s:%s/mcp", *host, *port)

		// graceful shutdown
		go func() {
			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
			<-sigCh
			log.Println("Shutting down HTTP server...")
			if err := httpServer.Shutdown(context.Background()); err != nil {
				log.Printf("Shutdown error: %v", err)
			}
		}()

		address := fmt.Sprintf("%s:%s", *host, *port)

		if err := httpServer.Start(address); err != nil {
			log.Fatalf("HTTP server error: %v", err)
		}

	default:
		log.Fatalf("Unknown transport: %s (use stdio, sse, or http)", *transport)
	}
}

func checkDevice() {
	// Enumerate connected devices
	dev := dwf.NewDevice()
	devices, err := dev.EnumDevices()
	if err != nil {
		log.Fatalf("Device enumeration failed: %v", err)
	}
	if len(devices) == 0 {
		fmt.Println("No connected devices found")
		return
	}

	// Print list of enumerated devices
	fmt.Printf("Enumerated Devices: %d\n", len(devices))
	for _, d := range devices {
		status := "available"
		if d.IsOpened {
			status = "in use"
		}
		name := d.DeviceName
		if d.UserName != "" && d.UserName != d.DeviceName {
			name += " (" + d.UserName + ")"
		}
		fmt.Printf("  %d) %-35s  %-20s  %s\n", d.Index, name, d.SerialNumber, status)
	}

	index := 0
	// if more than one device is connected, ask the user to select one
	if len(devices) > 1 {
		fmt.Println()
		fmt.Print("Select a device to open: ")
		fmt.Scanln(&index)
		if index < 0 || index >= len(devices) {
			fmt.Printf("Invalid device index, use 0 to %d\n", len(devices)-1)
			return
		}
	} else {
		fmt.Println()
		fmt.Println("Only one device connected, opening it...")
	}

	// Open the selected device
	info, err := dev.Open("", index)
	if err != nil {
		log.Fatalf("Device check failed: %v", err)
	}
	defer dev.Close()

	// Print device info
	fmt.Println()
	fmt.Println("[ Device Info ]")
	fmt.Printf("  Name:                %s\n", info.Name)
	fmt.Printf("  Serial Number:       %s\n", info.SerialNumber)
	fmt.Printf("  SDK Version:         %s\n", info.Version)
	fmt.Printf("  Analog In Channels:  %d\n", info.AnalogInChannels)
	fmt.Printf("  Analog Out Channels: %d\n", info.AnalogOutChannels)
	fmt.Printf("  Digital In Channels: %d\n", info.DigitalInChannels)
	fmt.Printf("  Digital Out Channels:%d\n", info.DigitalOutChannels)
	fmt.Printf("  Max Buffer Size:     %d\n", info.MaxAnalogInBufferSize)
	fmt.Printf("  ADC Resolution:      %d bits\n", info.MaxAnalogInResolution)

	// Print board temperature if available
	temp, err := dev.Temperature()
	if err == nil && temp > 0 {
		fmt.Printf("  Board Temperature:   %.1f °C\n", temp)
	}

	// Show available device configurations
	configs, err := dev.EnumConfigs(0)
	if err == nil && len(configs) > 0 {
		fmt.Println()
		fmt.Printf("[ Available Configurations: %d ]\n", len(configs))
		fmt.Println("  " + strings.Repeat("-", 88))
		fmt.Printf("  %-6s  %-6s %-6s %-6s %-6s %-6s %-6s %-8s %-8s %-8s %-8s\n",
			"Config", "AI-Ch", "AO-Ch", "IO-Ch", "DI-Ch", "DO-Ch", "DIO",
			"AI-Buf", "AO-Buf", "DI-Buf", "DO-Buf")
		fmt.Println("  " + strings.Repeat("-", 88))
		for i, cfg := range configs {
			fmt.Printf("    %-6d  %-6d %-6d %-6d %-5d %-5d %-6d %-8d %-8d %-8d %-8d\n",
				i, cfg.AnalogInChannels, cfg.AnalogOutChannels, cfg.AnalogIOChannels,
				cfg.DigitalInChannels, cfg.DigitalOutChannels, cfg.DigitalIOChannels,
				cfg.AnalogInBufferSize, cfg.AnalogOutBufferSize,
				cfg.DigitalInBufferSize, cfg.DigitalOutBufferSize)
		}
		fmt.Println("  " + strings.Repeat("-", 88))
	}

	fmt.Println()
	fmt.Println("Device is connected and operational.")
}
