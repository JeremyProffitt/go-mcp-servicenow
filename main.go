package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/elastiflow/go-mcp-servicenow/pkg/logging"
	"github.com/elastiflow/go-mcp-servicenow/pkg/mcp"
	"github.com/elastiflow/go-mcp-servicenow/pkg/servicenow"
	"github.com/elastiflow/go-mcp-servicenow/pkg/tools"
)

var Version = "1.0.0"

const AppName = "go-mcp-servicenow"

func main() {
	// Load environment file early (before parsing flags)
	logging.LoadEnvFile()

	// Parse command line flags
	logDir := flag.String("log-dir", "", "Directory for log files")
	logLevel := flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	httpMode := flag.Bool("http", false, "Run in HTTP mode instead of stdio")
	port := flag.Int("port", 3000, "HTTP port (only used with -http)")
	host := flag.String("host", "127.0.0.1", "HTTP host (only used with -http)")
	readOnlyMode := flag.Bool("read-only", false, "Enable read-only mode (disables write operations)")
	showVersion := flag.Bool("version", false, "Show version and exit")
	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Printf("%s version %s\n", AppName, Version)
		os.Exit(0)
	}

	// Resolve configuration with source tracking
	actualLogDir, logDirSource := resolveLogDir(*logDir)
	actualLogLevel, logLevelSource := resolveLogLevel(*logLevel)
	actualReadOnly := resolveReadOnlyMode(*readOnlyMode)

	// Initialize logger
	logger, err := logging.NewLogger(logging.Config{
		LogDir:          actualLogDir,
		AppName:         AppName,
		Level:           logging.ParseLevel(actualLogLevel),
		AddAppSubfolder: os.Getenv("MCP_LOG_DIR") != "",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	// Log startup information
	logger.LogStartup(logging.StartupInfo{
		Version:   Version,
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		NumCPU:    runtime.NumCPU(),
		LogDir:    logging.ConfigValue{Value: actualLogDir, Source: logDirSource},
		LogLevel:  logging.ConfigValue{Value: actualLogLevel, Source: logLevelSource},
		PID:       os.Getpid(),
		StartTime: time.Now(),
	})

	// Load ServiceNow configuration
	snConfig, err := servicenow.LoadConfigFromEnv()
	if err != nil {
		logger.Error("Failed to load ServiceNow configuration: %v", err)
		os.Exit(1)
	}

	// Mask sensitive values for logging
	maskedInstance := snConfig.InstanceURL
	if len(maskedInstance) > 30 {
		maskedInstance = maskedInstance[:30] + "..."
	}
	logger.Info("ServiceNow instance: %s", maskedInstance)
	logger.Info("Authentication type: %s", snConfig.Auth.Type)

	// Create ServiceNow client
	client, err := servicenow.NewClient(snConfig)
	if err != nil {
		logger.Error("Failed to create ServiceNow client: %v", err)
		os.Exit(1)
	}

	// Create MCP server
	server := mcp.NewServer(AppName, Version)

	// Set up telemetry callbacks
	server.SetToolCallCallback(func(name string, args map[string]interface{}, duration time.Duration, success bool) {
		logger.ToolCall(name, args, duration, success)
	})
	server.SetErrorCallback(func(err error, context string) {
		logger.Error("Error in %s: %v", context, err)
	})

	// Register tools
	registry := tools.NewRegistry(client, logger, actualReadOnly)
	toolCount := registry.RegisterAll(server)
	logger.Info("Registered %d tools (read-only mode: %v)", toolCount, actualReadOnly)

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Run server
	go func() {
		var runErr error
		if *httpMode {
			addr := fmt.Sprintf("%s:%d", *host, *port)
			logger.Info("Starting HTTP server on %s", addr)
			runErr = server.RunHTTP(addr)
		} else {
			logger.Info("Starting stdio server")
			runErr = server.Run()
		}
		if runErr != nil {
			logger.Error("Server error: %v", runErr)
			sigChan <- syscall.SIGTERM
		}
	}()

	// Wait for shutdown signal
	sig := <-sigChan
	logger.LogShutdown(fmt.Sprintf("received signal: %v", sig))
}

func resolveLogDir(flagValue string) (string, logging.ConfigSource) {
	if flagValue != "" {
		return flagValue, logging.SourceFlag
	}
	if envValue := os.Getenv("MCP_LOG_DIR"); envValue != "" {
		return envValue, logging.SourceEnvironment
	}
	return logging.DefaultLogDir(AppName), logging.SourceDefault
}

func resolveLogLevel(flagValue string) (string, logging.ConfigSource) {
	if flagValue != "info" {
		return flagValue, logging.SourceFlag
	}
	if envValue := os.Getenv("MCP_LOG_LEVEL"); envValue != "" {
		return envValue, logging.SourceEnvironment
	}
	return "info", logging.SourceDefault
}

func resolveReadOnlyMode(flagValue bool) bool {
	if flagValue {
		return true
	}
	envValue := strings.ToLower(os.Getenv("READ_ONLY_MODE"))
	return envValue == "true" || envValue == "1"
}
