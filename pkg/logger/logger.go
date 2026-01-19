package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sr-tamim/guardian/pkg/models"
	"github.com/sr-tamim/guardian/pkg/utils"
)

// Logger wraps slog.Logger with Guardian-specific functionality
type Logger struct {
	*slog.Logger
	config *models.LoggingConfig
	level  slog.Level
}

type syncWriter struct {
	w *os.File
}

func (s *syncWriter) Write(p []byte) (int, error) {
	n, err := s.w.Write(p)
	if err == nil {
		_ = s.w.Sync()
	}
	return n, err
}

// LogContext holds contextual information for each log entry
type LogContext struct {
	Function string // Function name
	File     string // File name (without path)
	Module   string // Module/package name
	Action   string // Custom action description
}

// Global logger instance
var globalLogger *Logger

// InitializeLogger sets up the global logger based on configuration
func InitializeLogger(config *models.LoggingConfig) error {
	// Always create a new logger to allow reconfiguration
	logger, err := NewLogger(config)
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}

	globalLogger = logger
	return nil
}

// NewLogger creates a new structured logger with Guardian-specific configuration
func NewLogger(config *models.LoggingConfig) (*Logger, error) {
	if config == nil {
		config = &models.LoggingConfig{
			Level:      "info",
			Format:     "text",
			Output:     "stdout",
			EnableFile: false,
		}
	}

	// Parse log level
	level, err := parseLogLevel(config.Level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level '%s': %w", config.Level, err)
	}

	// Set up output writers
	var writers []io.Writer

	// Add file output first (so stdout/stderr errors don't block file logging)
	if config.EnableFile {
		paths := utils.NewPlatformPaths()
		defaultPath := paths.GetDefaultGuardianLogPath()
		primaryPath := config.FilePath
		if primaryPath == "" {
			primaryPath = defaultPath
		}

		openLogFile := func(path string) (*os.File, error) {
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return nil, err
			}
			file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				return nil, err
			}
			_, _ = fmt.Fprintf(file, "[%s] Guardian logging initialized\n", time.Now().Format(time.RFC3339))
			_ = file.Sync()
			return file, nil
		}

		file, err := openLogFile(primaryPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open log file (%s): %v\n", primaryPath, err)
			if primaryPath != defaultPath {
				file, err = openLogFile(defaultPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "failed to open fallback log file (%s): %v\n", defaultPath, err)
				} else {
					config.FilePath = defaultPath
					writers = append(writers, &syncWriter{w: file})
				}
			}
		} else {
			config.FilePath = primaryPath
			writers = append(writers, &syncWriter{w: file})
		}
	}

	// Add stdout/stderr if configured
	if config.Output == "stdout" {
		writers = append(writers, os.Stdout)
	} else if config.Output == "stderr" {
		writers = append(writers, os.Stderr)
	}

	// If no writers configured, default to stdout
	if len(writers) == 0 {
		writers = append(writers, os.Stdout)
	}

	// Create multi-writer
	var output io.Writer
	if len(writers) == 1 {
		output = writers[0]
	} else {
		output = io.MultiWriter(writers...)
	}

	// Configure handler options
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: true, // Always include source info for debugging
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Customize timestamp format
			if a.Key == slog.TimeKey {
				return slog.String("timestamp", a.Value.Time().Format(time.RFC3339))
			}
			// Simplify source attribute to just filename:line
			if a.Key == slog.SourceKey {
				source := a.Value.Any().(*slog.Source)
				return slog.String("source", fmt.Sprintf("%s:%d", filepath.Base(source.File), source.Line))
			}
			return a
		},
	}

	// Create appropriate handler based on format
	var handler slog.Handler
	if config.Format == "json" {
		handler = slog.NewJSONHandler(output, opts)
	} else {
		handler = slog.NewTextHandler(output, opts)
	}

	return &Logger{
		Logger: slog.New(handler),
		config: config,
		level:  level,
	}, nil
}

// GetLogger returns the global logger instance
func GetLogger() *Logger {
	if globalLogger == nil {
		// Fallback to default logger if not initialized
		config := &models.LoggingConfig{
			Level:  "info",
			Format: "text",
			Output: "stdout",
		}
		logger, _ := NewLogger(config)
		globalLogger = logger
	}
	return globalLogger
}

// WithContext creates a logger with additional context information
func (l *Logger) WithContext(ctx LogContext) *Logger {
	args := make([]any, 0, 8) // Each attribute becomes 2 args (key, value)

	if ctx.Function != "" {
		args = append(args, "function", ctx.Function)
	}
	if ctx.File != "" {
		args = append(args, "file", ctx.File)
	}
	if ctx.Module != "" {
		args = append(args, "module", ctx.Module)
	}
	if ctx.Action != "" {
		args = append(args, "action", ctx.Action)
	}

	return &Logger{
		Logger: l.Logger.With(args...),
		config: l.config,
		level:  l.level,
	}
}

// Convenience methods for different log levels with automatic context detection

// Debug logs debug-level messages with automatic context detection
func (l *Logger) Debug(msg string, args ...any) {
	if l.level <= slog.LevelDebug {
		l.Logger.Debug(msg, args...)
	}
}

// Info logs info-level messages with automatic context detection
func (l *Logger) Info(msg string, args ...any) {
	if l.level <= slog.LevelInfo {
		l.Logger.Info(msg, args...)
	}
}

// Warn logs warning-level messages with automatic context detection
func (l *Logger) Warn(msg string, args ...any) {
	if l.level <= slog.LevelWarn {
		l.Logger.Warn(msg, args...)
	}
}

// Error logs error-level messages with automatic context detection
func (l *Logger) Error(msg string, args ...any) {
	l.Logger.Error(msg, args...)
}

// Event-specific logging methods for Guardian

// LogEventLookup logs event lookup operations with count information
func (l *Logger) LogEventLookup(service string, logPath string, eventCount int, ips []string) {
	ctx := LogContext{
		Module: "event_lookup",
		Action: "scan_logs",
	}
	l.WithContext(ctx).Info("Event lookup completed",
		slog.String("service", service),
		slog.String("log_path", logPath),
		slog.Int("event_count", eventCount),
		slog.Int("unique_ips", len(ips)),
		slog.Any("found_ips", ips),
	)
}

// LogIPBlocked logs when an IP is blocked with firewall rule details
func (l *Logger) LogIPBlocked(ip string, reason string, ruleName string, duration time.Duration) {
	ctx := LogContext{
		Module: "firewall",
		Action: "block_ip",
	}
	l.WithContext(ctx).Info("IP address blocked",
		slog.String("ip", ip),
		slog.String("reason", reason),
		slog.String("rule_name", ruleName),
		slog.Duration("duration", duration),
	)
}

// LogIPUnblocked logs when an IP is unblocked
func (l *Logger) LogIPUnblocked(ip string, ruleName string, activeTime time.Duration) {
	ctx := LogContext{
		Module: "firewall",
		Action: "unblock_ip",
	}
	l.WithContext(ctx).Info("IP address unblocked",
		slog.String("ip", ip),
		slog.String("rule_name", ruleName),
		slog.Duration("active_time", activeTime),
	)
}

// LogAttackAttempt logs detected attack attempts
func (l *Logger) LogAttackAttempt(ip string, service string, username string, severity string) {
	ctx := LogContext{
		Module: "threat_detection",
		Action: "attack_detected",
	}
	l.WithContext(ctx).Warn("Attack attempt detected",
		slog.String("ip", ip),
		slog.String("service", service),
		slog.String("username", username),
		slog.String("severity", severity),
	)
}

// LogMonitoringStart logs when monitoring starts for a service
func (l *Logger) LogMonitoringStart(service string, logPath string, provider string) {
	ctx := LogContext{
		Module: "monitoring",
		Action: "start_monitoring",
	}
	l.WithContext(ctx).Info("Started monitoring service",
		slog.String("service", service),
		slog.String("log_path", logPath),
		slog.String("provider", provider),
	)
}

// LogCleanupOperation logs cleanup operations
func (l *Logger) LogCleanupOperation(rulesRemoved int, totalProcessed int) {
	ctx := LogContext{
		Module: "cleanup",
		Action: "expire_rules",
	}
	l.WithContext(ctx).Info("Cleanup operation completed",
		slog.Int("rules_removed", rulesRemoved),
		slog.Int("rules_processed", totalProcessed),
	)
}

// Global convenience functions that use the global logger

// Debug logs debug-level messages using the global logger
func Debug(msg string, args ...any) {
	if globalLogger != nil {
		globalLogger.Debug(msg, args...)
	}
}

// Info logs info-level messages using the global logger
func Info(msg string, args ...any) {
	if globalLogger != nil {
		globalLogger.Info(msg, args...)
	}
}

// Warn logs warning-level messages using the global logger
func Warn(msg string, args ...any) {
	if globalLogger != nil {
		globalLogger.Warn(msg, args...)
	}
}

// Error logs error-level messages using the global logger
func Error(msg string, args ...any) {
	if globalLogger != nil {
		globalLogger.Error(msg, args...)
	}
}

// Guardian-specific elegant logging functions that handle config checks internally

// LogIPBlocked logs IP blocking with automatic config checking
func LogIPBlocked(config *models.Config, ip, reason, ruleName string, duration time.Duration) {
	if globalLogger != nil && config != nil && config.Logging.LogFirewallActions {
		globalLogger.LogIPBlocked(ip, reason, ruleName, duration)
	}
}

// LogIPUnblocked logs IP unblocking with automatic config checking
func LogIPUnblocked(config *models.Config, ip, ruleName string, activeTime time.Duration) {
	if globalLogger != nil && config != nil && config.Logging.LogFirewallActions {
		globalLogger.LogIPUnblocked(ip, ruleName, activeTime)
	}
}

// LogAttackAttempt logs attack attempts with automatic config checking
func LogAttackAttempt(config *models.Config, ip, service, username, severity string) {
	if globalLogger != nil && config != nil && config.Logging.LogAttackAttempts {
		globalLogger.LogAttackAttempt(ip, service, username, severity)
	}
}

// LogMonitoringStart logs monitoring start events with automatic config checking
func LogMonitoringStart(config *models.Config, service, logPath, provider string) {
	if globalLogger != nil && config != nil && config.Logging.LogMonitoringEvents {
		globalLogger.LogMonitoringStart(service, logPath, provider)
	}
}

// LogEventLookup logs event lookup operations with automatic config checking
func LogEventLookup(config *models.Config, service, logPath string, eventCount int, ips []string) {
	if globalLogger != nil && config != nil && config.Logging.LogEventLookups {
		globalLogger.LogEventLookup(service, logPath, eventCount, ips)
	}
}

// LogCleanupOperation logs cleanup operations with automatic config checking
func LogCleanupOperation(config *models.Config, rulesRemoved, totalProcessed int) {
	if globalLogger != nil && config != nil && config.Logging.LogCleanupEvents {
		globalLogger.LogCleanupOperation(rulesRemoved, totalProcessed)
	}
}

// Console output functions that also log if structured logging is enabled

// InfoConsole prints to console and optionally logs to structured logger
func InfoConsole(msg string, args ...any) {
	// Always print to console for immediate feedback
	if len(args) > 0 {
		fmt.Printf(msg+"\n", args...)
	} else {
		fmt.Println(msg)
	}
	// Also log if structured logging is available
	if globalLogger != nil {
		globalLogger.Info(msg, args...)
	}
}

// ErrorConsole prints error to console and logs to structured logger
func ErrorConsole(msg string, args ...any) {
	// Always print to console for immediate feedback
	if len(args) > 0 {
		fmt.Printf("❌ "+msg+"\n", args...)
	} else {
		fmt.Println("❌ " + msg)
	}
	// Also log if structured logging is available
	if globalLogger != nil {
		globalLogger.Error(msg, args...)
	}
}

// Helper functions

// parseLogLevel converts string log level to slog.Level
func parseLogLevel(level string) (slog.Level, error) {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("unknown log level: %s", level)
	}
}

// getCallerContext automatically extracts context from the calling function
func getCallerContext() LogContext {
	return LogContext{
		Function: "unknown",
		File:     "unknown:0",
		Module:   "unknown",
	}
}
