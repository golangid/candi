package logger

import (
	"fmt"
	"time"

	"agungdwiprasetyo.com/backend-microservices/pkg/helper"
)

// LogWithDefer return defer func for status
func LogWithDefer(str string) (deferFunc func()) {
	fmt.Printf("%s %s ", time.Now().Format(helper.TimeFormatLogger), str)
	return func() {
		if r := recover(); r != nil {
			fmt.Printf("\x1b[31;1mERROR: %v\x1b[0m\n", r)
			panic(r)
		}
		fmt.Println("\x1b[32;1mSUCCESS\x1b[0m")
	}
}
