package logger

import (
	"bytes"
	"encoding/json"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		env     string
		wantErr bool
	}{
		{
			name:    "creates production logger",
			env:     "production",
			wantErr: false,
		},
		{
			name:    "creates development logger",
			env:     "development",
			wantErr: false,
		},
		{
			name:    "defaults to production for unknown env",
			env:     "unknown",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := New(tt.env, "test-service")
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if logger == nil && !tt.wantErr {
				t.Error("New() returned nil logger")
			}
			if logger != nil {
				logger.Sync()
			}
		})
	}
}

func TestLogger_StructuredFields(t *testing.T) {
	// Create a logger that writes to a buffer for testing
	buf := &bytes.Buffer{}
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = ""  // Remove timestamp for easier testing
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(buf),
		zapcore.InfoLevel,
	)
	zapLogger := zap.New(core)
	logger := &Logger{logger: zapLogger}

	// Log a message with structured fields
	logger.Info("test message",
		zap.String("key1", "value1"),
		zap.Int("key2", 42),
	)

	// Parse the JSON output
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	// Verify structured fields
	if logEntry["msg"] != "test message" {
		t.Errorf("msg = %v, want 'test message'", logEntry["msg"])
	}
	if logEntry["key1"] != "value1" {
		t.Errorf("key1 = %v, want 'value1'", logEntry["key1"])
	}
	if logEntry["key2"] != float64(42) {  // JSON numbers are float64
		t.Errorf("key2 = %v, want 42", logEntry["key2"])
	}
}

func TestLogger_WithContext(t *testing.T) {
	buf := &bytes.Buffer{}
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = ""
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(buf),
		zapcore.InfoLevel,
	)
	zapLogger := zap.New(core)
	logger := &Logger{logger: zapLogger}

	// Create a logger with context fields
	contextLogger := logger.With(
		zap.String("request_id", "req-123"),
		zap.String("user_id", "user-456"),
	)

	contextLogger.Info("test message")

	// Parse and verify
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	if logEntry["request_id"] != "req-123" {
		t.Errorf("request_id = %v, want 'req-123'", logEntry["request_id"])
	}
	if logEntry["user_id"] != "user-456" {
		t.Errorf("user_id = %v, want 'user-456'", logEntry["user_id"])
	}
}

func TestLogger_Levels(t *testing.T) {
	tests := []struct {
		name     string
		logFn    func(*Logger, string, ...zap.Field)
		level    zapcore.Level
		wantLog  bool
	}{
		{
			name:    "debug logs at debug level",
			logFn:   (*Logger).Debug,
			level:   zapcore.DebugLevel,
			wantLog: true,
		},
		{
			name:    "info logs at info level",
			logFn:   (*Logger).Info,
			level:   zapcore.InfoLevel,
			wantLog: true,
		},
		{
			name:    "warn logs at warn level",
			logFn:   (*Logger).Warn,
			level:   zapcore.WarnLevel,
			wantLog: true,
		},
		{
			name:    "error logs at error level",
			logFn:   (*Logger).Error,
			level:   zapcore.ErrorLevel,
			wantLog: true,
		},
		{
			name:    "debug does not log at info level",
			logFn:   (*Logger).Debug,
			level:   zapcore.InfoLevel,
			wantLog: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			encoderConfig := zap.NewProductionEncoderConfig()
			core := zapcore.NewCore(
				zapcore.NewJSONEncoder(encoderConfig),
				zapcore.AddSync(buf),
				tt.level,
			)
			zapLogger := zap.New(core)
			logger := &Logger{logger: zapLogger}

			tt.logFn(logger, "test message")

			logged := buf.Len() > 0
			if logged != tt.wantLog {
				t.Errorf("logged = %v, want %v", logged, tt.wantLog)
			}
		})
	}
}

func TestLogger_Fatal(t *testing.T) {
	// Note: We can't easily test Fatal as it calls os.Exit
	// This is a placeholder test to document the behavior
	t.Skip("Fatal calls os.Exit, cannot test directly")
}
