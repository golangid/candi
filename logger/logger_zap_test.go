package logger_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/golangid/candi/logger"
	"go.uber.org/zap/zapcore"
)

// helper function to capture logs
func captureLogs() (*bytes.Buffer, *zapcore.Core) {
	buf := new(bytes.Buffer)
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zapcore.EncoderConfig{
			MessageKey: "message",
			LevelKey:   "level",
			TimeKey:    "time",
			CallerKey:  "caller",
		}),
		zapcore.AddSync(buf),
		zapcore.DebugLevel,
	)

	return buf, &core
}

func TestInitZap(t *testing.T) {
	logOutput, _ := captureLogs()
	logger.InitZap(logger.OptionAddWriter(io.MultiWriter(logOutput)))

	logger.LogI("test message")

	// Check if the log output contains the expected message
	if !bytes.Contains(logOutput.Bytes(), []byte("test message")) {
		t.Error("Expected log message not found")
	}
}

func TestLog(t *testing.T) {
	logOutput, _ := captureLogs()
	logger.InitZap(logger.OptionAddWriter(io.MultiWriter(logOutput)))

	// Log a message with context and scope
	logger.Log(zapcore.InfoLevel, "testing log", "test_context", "test_scope")

	// Verify log output contains expected fields
	if !bytes.Contains(logOutput.Bytes(), []byte(`"testing log"`)) {
		t.Error("Expected log message not found")
	}
	if !bytes.Contains(logOutput.Bytes(), []byte(`"context":"test_context"`)) {
		t.Error("Expected context not found")
	}
	if !bytes.Contains(logOutput.Bytes(), []byte(`"scope":"test_scope"`)) {
		t.Error("Expected scope not found")
	}
}

func TestLogE(t *testing.T) {
	logOutput, _ := captureLogs()
	logger.InitZap(logger.OptionAddWriter(io.MultiWriter(logOutput)))

	logger.LogE("test error message")

	if !bytes.Contains(logOutput.Bytes(), []byte("test error message")) {
		t.Error("Expected error message not found")
	}
}

func TestLogIfError(t *testing.T) {
	logOutput, _ := captureLogs()
	logger.InitZap(logger.OptionAddWriter(io.MultiWriter(logOutput)))

	err := io.EOF
	logger.LogIfError(err)

	if !bytes.Contains(logOutput.Bytes(), []byte("EOF")) {
		t.Error("Expected error message not found")
	}
}

func TestLogPanicIfError(t *testing.T) {
	// Capture panic output
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic but did not occur")
		}
	}()

	logger.InitZap()
	logger.LogPanicIfError(io.EOF)
}

func TestLogEf(t *testing.T) {
	logOutput, _ := captureLogs()
	logger.InitZap(logger.OptionAddWriter(io.MultiWriter(logOutput)))

	logger.LogEf("formatted error: %s", "something went wrong")

	if !bytes.Contains(logOutput.Bytes(), []byte("formatted error: something went wrong")) {
		t.Error("Expected formatted error message not found")
	}
}

func TestLogIf(t *testing.T) {
	logOutput, _ := captureLogs()
	logger.InitZap(logger.OptionAddWriter(io.MultiWriter(logOutput)))

	logger.LogIf("this is a formatted info: %s", "info message")

	if !bytes.Contains(logOutput.Bytes(), []byte("this is a formatted info: info message")) {
		t.Error("Expected formatted info message not found")
	}
}

func TestLogWithField(t *testing.T) {
	logOutput, _ := captureLogs()
	logger.InitZap(logger.OptionAddWriter(io.MultiWriter(logOutput)))

	fields := map[string]interface{}{
		"message": "test log with fields",
		"context": "test_context",
		"scope":   "test_scope",
	}

	logger.LogWithField(zapcore.InfoLevel, fields)

	if !bytes.Contains(logOutput.Bytes(), []byte("test log with fields")) {
		t.Error("Expected message not found in log output")
	}
	if !bytes.Contains(logOutput.Bytes(), []byte(`"context":"test_context"`)) {
		t.Error("Expected context field not found in log output")
	}
	if !bytes.Contains(logOutput.Bytes(), []byte(`"scope":"test_scope"`)) {
		t.Error("Expected scope field not found in log output")
	}
}
