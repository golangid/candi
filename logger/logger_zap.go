package logger

import (
	"io"
	"os"
	"runtime"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// InitZap logger with default writer to stdout
func InitZap(opts ...OptionFunc) {
	opt := Option{
		MultiWriter: []io.Writer{os.Stdout},
	}

	for _, o := range opts {
		o(&opt)
	}

	encCfg := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		MessageKey: "message",

		LevelKey:    "level",
		EncodeLevel: zapcore.CapitalLevelEncoder,

		TimeKey:    "time",
		EncodeTime: zapcore.ISO8601TimeEncoder,

		CallerKey: "caller",
		EncodeCaller: func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
			caller.PC, caller.File, caller.Line, _ = runtime.Caller(7)
			enc.AppendString(caller.FullPath())
		},
	})

	var coreOpt []zapcore.Core
	for _, w := range opt.MultiWriter {
		coreOpt = append(coreOpt, zapcore.NewCore(encCfg, zapcore.AddSync(w), zapcore.DebugLevel))
	}
	core := zapcore.NewTee(coreOpt...)

	zapLog := zap.New(core, zap.AddCaller())
	zap.ReplaceGlobals(zapLog)
	zapLog.Sync()
}

// Log func
func Log(level zapcore.Level, message string, context string, scope string) {

	entry := zap.S().With(
		zap.String("context", context),
		zap.String("scope", scope),
	)

	setEntryType(level, entry, message)
}

// LogWithField func
func LogWithField(level zapcore.Level, fields map[string]interface{}) {

	var message interface{}
	var args []interface{}
	for k, v := range fields {
		if k == "message" {
			message = v
			continue
		}
		args = append(args, []interface{}{k, v}...)
	}
	entry := zap.S().With(args...)
	setEntryType(level, entry, message)
}

// LogE error
func LogE(message string) {
	zap.S().Error(message)
}

// LogIfError log if error not nil
func LogIfError(err error) {
	if err != nil {
		zap.S().Error(err.Error())
	}
}

// LogPanicIfError log if error not nil
func LogPanicIfError(err error) {
	if err != nil {
		zap.S().Panic(err.Error())
	}
}

// LogEf error with format
func LogEf(format string, i ...interface{}) {
	zap.S().Errorf(format, i...)
}

// LogI info
func LogI(message string) {
	zap.S().Info(message)
}

// LogIf info with format
func LogIf(format string, i ...interface{}) {
	zap.S().Infof(format, i...)
}

func setEntryType(level zapcore.Level, entry *zap.SugaredLogger, msg interface{}) {

	switch level {
	case zapcore.DebugLevel:
		entry.Debug(msg)
	case zapcore.InfoLevel:
		entry.Info(msg)
	case zapcore.WarnLevel:
		entry.Warn(msg)
	case zapcore.ErrorLevel:
		entry.Error(msg)
	case zapcore.FatalLevel:
		entry.Fatal(msg)
	case zapcore.PanicLevel:
		entry.Panic(msg)
	}
}
