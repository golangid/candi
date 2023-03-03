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
		fmt.Printf("%s\n", YellowColor(str))
	}
}

// LogRed log with red color
func LogRed(str string) {
	if env.BaseEnv().DebugMode {
		fmt.Printf("%s\n", RedColor(str))
	}
}

// LogGreen log with green color
func LogGreen(str string) {
	if env.BaseEnv().DebugMode {
		fmt.Printf("%s\n", GreenColor(str))
	}
}

// RedColor func
func RedColor(str interface{}) string {
	return fmt.Sprintf("\x1b[31;5m%v\x1b[0m", str)
}

// GreenColor func
func GreenColor(str interface{}) string {
	return fmt.Sprintf("\x1b[32;5m%v\x1b[0m", str)
}

// YellowColor func
func YellowColor(str interface{}) string {
	return fmt.Sprintf("\x1b[33;5m%v\x1b[0m", str)
}

// CyanColor func
func CyanColor(str interface{}) string {
	return fmt.Sprintf("\x1b[36;5m%v\x1b[0m", str)
}
