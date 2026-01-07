package logging

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Level represents logging level
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ParseLevel parses a string into a Level
func ParseLevel(s string) Level {
	switch strings.ToLower(s) {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}

// ConfigSource indicates where a configuration value came from
type ConfigSource string

const (
	SourceDefault     ConfigSource = "default"
	SourceEnvironment ConfigSource = "environment"
	SourceFlag        ConfigSource = "flag"
)

// ConfigValue represents a configuration value with its source
type ConfigValue struct {
	Value  string
	Source ConfigSource
}

// Config holds logger configuration
type Config struct {
	LogDir          string
	AppName         string
	Level           Level
	AddAppSubfolder bool
}

// StartupInfo contains information logged at startup
type StartupInfo struct {
	Version     string
	GoVersion   string
	OS          string
	Arch        string
	NumCPU      int
	LogDir      ConfigValue
	LogLevel    ConfigValue
	InstanceURL ConfigValue
	PID         int
	StartTime   time.Time
}

// Logger provides structured logging
type Logger struct {
	config    Config
	file      *os.File
	mu        sync.Mutex
	startTime time.Time
}

// NewLogger creates a new logger
func NewLogger(config Config) (*Logger, error) {
	logger := &Logger{
		config:    config,
		startTime: time.Now(),
	}

	if config.LogDir != "" {
		logDir := config.LogDir
		if config.AddAppSubfolder {
			logDir = filepath.Join(logDir, config.AppName)
		}

		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		logFile := filepath.Join(logDir, fmt.Sprintf("%s.log", config.AppName))
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		logger.file = file
	}

	return logger, nil
}

// Close closes the logger
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// log writes a log message
func (l *Logger) log(level Level, format string, args ...interface{}) {
	if level < l.config.Level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	msg := fmt.Sprintf(format, args...)
	timestamp := time.Now().Format("2006-01-02T15:04:05.000Z07:00")
	logLine := fmt.Sprintf("[%s] [%s] %s\n", timestamp, level.String(), msg)

	if l.file != nil {
		_, _ = l.file.WriteString(logLine)
	}

	// Also write to stderr for debugging
	fmt.Fprint(os.Stderr, logLine)
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(LevelDebug, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LevelInfo, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(LevelWarn, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(LevelError, format, args...)
}

// ToolCall logs a tool call
func (l *Logger) ToolCall(name string, args map[string]interface{}, duration time.Duration, success bool) {
	status := "success"
	if !success {
		status = "failure"
	}
	l.Info("Tool call: %s (duration: %v, status: %s)", name, duration, status)
}

// LogStartup logs startup information
func (l *Logger) LogStartup(info StartupInfo) {
	l.Info("=== %s Starting ===", l.config.AppName)
	l.Info("Version: %s", info.Version)
	l.Info("Go Version: %s", info.GoVersion)
	l.Info("OS/Arch: %s/%s", info.OS, info.Arch)
	l.Info("NumCPU: %d", info.NumCPU)
	l.Info("PID: %d", info.PID)
	l.Info("Log Directory: %s (source: %s)", info.LogDir.Value, info.LogDir.Source)
	l.Info("Log Level: %s (source: %s)", info.LogLevel.Value, info.LogLevel.Source)
	if info.InstanceURL.Value != "" {
		l.Info("ServiceNow Instance: %s (source: %s)", info.InstanceURL.Value, info.InstanceURL.Source)
	}
}

// LogShutdown logs shutdown information
func (l *Logger) LogShutdown(reason string) {
	uptime := time.Since(l.startTime)
	l.Info("=== %s Shutting Down ===", l.config.AppName)
	l.Info("Reason: %s", reason)
	l.Info("Uptime: %v", uptime)
}

// DefaultLogDir returns the default log directory for the given app
func DefaultLogDir(appName string) string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("LOCALAPPDATA"), appName, "logs")
	case "darwin":
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Logs", appName)
	default:
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".local", "share", appName, "logs")
	}
}

// LoadEnvFile loads environment variables from a .env file if it exists
func LoadEnvFile() {
	envFile := ".env"
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		return
	}

	file, err := os.Open(envFile)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}

		// Only set if not already set
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
}

// Writer returns an io.Writer that logs at the given level
func (l *Logger) Writer(level Level) io.Writer {
	return &logWriter{logger: l, level: level}
}

type logWriter struct {
	logger *Logger
	level  Level
}

func (w *logWriter) Write(p []byte) (n int, err error) {
	w.logger.log(w.level, "%s", strings.TrimSpace(string(p)))
	return len(p), nil
}
