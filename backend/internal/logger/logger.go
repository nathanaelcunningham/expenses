package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"
)

// Logger defines the interface for structured logging
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, err error, fields ...Field)
	Error(msg string, err error, fields ...Field)
	Fatal(msg string, err error, fields ...Field)
	With(fields ...Field) Logger
}

// Field represents a key-value pair for structured logging
type Field struct {
	Key   string
	Value any
}

// Str creates a string field
func Str(key, value string) Field {
	return Field{Key: key, Value: value}
}

// Int creates an integer field
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Int64 creates an int64 field
func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

// Bool creates a boolean field
func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

// Duration creates a duration field
func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value}
}

// Err creates an error field
func Err(err error) Field {
	if err == nil {
		return Field{Key: "error", Value: nil}
	}
	return Field{Key: "error", Value: err.Error()}
}

// logger implements the Logger interface using slog
type logger struct {
	slogger   *slog.Logger
	component string
}

// New creates a new logger instance
func New(output io.Writer) Logger {
	if output == nil {
		output = os.Stdout
	}

	// Create a custom handler for better formatting
	handler := &readableHandler{
		output: output,
	}

	return &logger{
		slogger: slog.New(handler),
	}
}

// NewWithComponent creates a logger with a component name
func NewWithComponent(output io.Writer, component string) Logger {
	l := New(output).(*logger)
	l.component = component
	return l
}

// Debug logs a debug message
func (l *logger) Debug(msg string, fields ...Field) {
	l.log(slog.LevelDebug, msg, nil, fields...)
}

// Info logs an info message
func (l *logger) Info(msg string, fields ...Field) {
	l.log(slog.LevelInfo, msg, nil, fields...)
}

// Warn logs a warning message
func (l *logger) Warn(msg string, err error, fields ...Field) {
	if err != nil {
		fields = append(fields, Err(err))
	}
	l.log(slog.LevelWarn, msg, err, fields...)
}

// Error logs an error message
func (l *logger) Error(msg string, err error, fields ...Field) {
	if err != nil {
		fields = append(fields, Err(err))
	}
	l.log(slog.LevelError, msg, err, fields...)
}

// Fatal logs a fatal message and exits
func (l *logger) Fatal(msg string, err error, fields ...Field) {
	if err != nil {
		fields = append(fields, Err(err))
	}
	l.log(slog.LevelError, msg, err, fields...)
	os.Exit(1)
}

// With creates a new logger with additional fields
func (l *logger) With(fields ...Field) Logger {
	newLogger := &logger{
		slogger:   l.slogger,
		component: l.component,
	}

	if len(fields) > 0 {
		args := make([]any, len(fields)*2)
		for i, f := range fields {
			args[i*2] = f.Key
			args[i*2+1] = f.Value
		}
		newLogger.slogger = l.slogger.With(args...)
	}

	return newLogger
}

// log is the internal logging method
func (l *logger) log(level slog.Level, msg string, err error, fields ...Field) {
	attrs := make([]any, 0, len(fields)*2)

	// Add component if set
	if l.component != "" {
		attrs = append(attrs, "component", l.component)
	}

	// Add fields
	for _, f := range fields {
		attrs = append(attrs, f.Key, f.Value)
	}

	l.slogger.Log(nil, level, msg, attrs...)
}

// readableHandler provides a custom, human-readable log format
type readableHandler struct {
	output io.Writer
}

// Handle formats and writes log records
func (h *readableHandler) Handle(ctx context.Context, r slog.Record) error {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// Color coding for different log levels
	var levelStr string
	switch r.Level {
	case slog.LevelDebug:
		levelStr = "\033[36mDEBUG\033[0m" // Cyan
	case slog.LevelInfo:
		levelStr = "\033[32mINFO \033[0m" // Green
	case slog.LevelWarn:
		levelStr = "\033[33mWARN \033[0m" // Yellow
	case slog.LevelError:
		levelStr = "\033[31mERROR\033[0m" // Red
	default:
		levelStr = fmt.Sprintf("%-5s", r.Level.String())
	}

	// Start building the log line
	output := fmt.Sprintf("%s [%s] ", timestamp, levelStr)

	// Extract component if available
	var component string
	var fields []string

	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "component" {
			component = a.Value.String()
		} else {
			fields = append(fields, fmt.Sprintf("%s=%v", a.Key, a.Value))
		}
		return true
	})

	// Add component
	if component != "" {
		output += fmt.Sprintf("\033[35m%s\033[0m: ", component) // Magenta for component
	}

	// Add message
	output += r.Message

	// Add fields
	if len(fields) > 0 {
		output += " "
		for i, field := range fields {
			if i > 0 {
				output += " "
			}
			output += fmt.Sprintf("\033[90m%s\033[0m", field) // Gray for fields
		}
	}

	output += "\n"

	_, err := h.output.Write([]byte(output))
	return err
}

// Enabled always returns true - we'll handle level filtering elsewhere if needed
func (h *readableHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

// WithAttrs creates a new handler with additional attributes
func (h *readableHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// For simplicity, we'll return the same handler
	// In a more complex implementation, we'd preserve these attrs
	return h
}

// WithGroup creates a new handler with a group name
func (h *readableHandler) WithGroup(name string) slog.Handler {
	// For simplicity, we'll return the same handler
	// In a more complex implementation, we'd handle grouping
	return h
}

