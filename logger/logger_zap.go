package logger

import (
	"os"
	"runtime"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.SugaredLogger

// InitZap logger
func InitZap() {
	var (
		logg *zap.Logger
		err  error
	)

	cfg := zap.Config{
		Encoding:         "json",
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey: "message",

			LevelKey:    "level",
			EncodeLevel: zapcore.CapitalLevelEncoder,

			TimeKey:    "time",
			EncodeTime: zapcore.ISO8601TimeEncoder,

			CallerKey: "caller",
			EncodeCaller: func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
				_, caller.File, caller.Line, _ = runtime.Caller(7)
				enc.AppendString(caller.FullPath())
			},
		},
		Level:       zap.NewAtomicLevelAt(zapcore.DebugLevel),
		Development: !(strings.ToLower(os.Getenv("ENVIRONMENT")) == "production"),
	}

	logg, err = cfg.Build()
	if err != nil {
		panic(err)
	}
	defer logg.Sync()

	// define logger
	logger = logg.Sugar()
}

// Log func
func Log(level zapcore.Level, message string, context string, scope string) {
	entry := logger.With(
		zap.String("context", context),
		zap.String("scope", scope),
	)

	switch level {
	case zapcore.DebugLevel:
		entry.Debug(message)
	case zapcore.InfoLevel:
		entry.Info(message)
	case zapcore.WarnLevel:
		entry.Warn(message)
	case zapcore.ErrorLevel:
		entry.Error(message)
	case zapcore.FatalLevel:
		entry.Fatal(message)
	case zapcore.PanicLevel:
		entry.Panic(message)
	}
}

// LogE error
func LogE(message string) {
	logger.Error(message)
}

// LogEf error with format
func LogEf(format string, i ...interface{}) {
	logger.Errorf(format, i...)
}

// LogI info
func LogI(message string) {
	logger.Info(message)
}

// LogIf info with format
func LogIf(format string, i ...interface{}) {
	logger.Infof(format, i...)
}
