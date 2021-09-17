package logger

import (
	"fmt"
	"time"

	"github.com/golangid/candi/candihelper"
	"github.com/golangid/candi/config/env"
)

func init() {
	InitZap()
}

// LogWithDefer return defer func for status
func LogWithDefer(str string) (deferFunc func()) {
	fmt.Printf("%s %s ", time.Now().Format(candihelper.TimeFormatLogger), str)
	return func() {
		if r := recover(); r != nil {
			fmt.Printf("\x1b[31;1mERROR: %v\x1b[0m\n", r)
			panic(r)
		}
		fmt.Println("\x1b[32;1mSUCCESS\x1b[0m")
	}
}

// LogYellow log with yellow color
func LogYellow(str string) {
	if env.BaseEnv().DebugMode {
		fmt.Printf("\x1b[33;5m%s\x1b[0m\n", str)
	}
}

// LogRed log with red color
func LogRed(str string) {
	if env.BaseEnv().DebugMode {
		fmt.Printf("\x1b[31;5m%s\x1b[0m\n", str)
	}
}

// LogGreen log with green color
func LogGreen(str string) {
	if env.BaseEnv().DebugMode {
		fmt.Printf("\x1b[32;5m%s\x1b[0m\n", str)
	}
}
