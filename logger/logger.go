package logger

import (
	"fmt"
	"time"

	"pkg.agungdwiprasetyo.com/candi/candihelper"
)

var debugMode bool

func init() {
	InitZap()
}

// SetDebugMode set local debug mode
func SetDebugMode(mode bool) {
	debugMode = mode
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
	if debugMode {
		fmt.Printf("\x1b[33;2m%s\x1b[0m\n", str)
	}
}

// LogRed log with red color
func LogRed(str string) {
	if debugMode {
		fmt.Printf("\x1b[31;2m%s\x1b[0m\n", str)
	}
}

// LogGreen log with green color
func LogGreen(str string) {
	if debugMode {
		fmt.Printf("\x1b[32;2m%s\x1b[0m\n", str)
	}
}
