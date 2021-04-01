package candihelper

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTryCatch_Do(t *testing.T) {
	t.Run("Test Catch Panic", func(t *testing.T) {
		TryCatch{
			Try: func() {
				panic("test")
			},
			Catch: func(e error) {
				assert.NotNil(t, e)
				assert.Equal(t, e, errors.New("test"))
			},
		}.Do()
	})
	t.Run("Test Catch Panic Nil Pointer", func(t *testing.T) {
		TryCatch{
			Try: func() {
				var a *struct {
					s string
				}
				fmt.Println(a.s)
			},
			Catch: func(e error) {
				assert.NotNil(t, e)
				assert.Contains(t, e.Error(), "invalid memory address or nil pointer dereference")
			},
		}.Do()
	})
	t.Run("Test Catch Panic index out of range", func(t *testing.T) {
		TryCatch{
			Try: func() {
				var a []string
				fmt.Println(a[10])
			},
			Catch: func(e error) {
				assert.NotNil(t, e)
				assert.Contains(t, e.Error(), "index out of range")
			},
		}.Do()
	})
	t.Run("Test Catch Panic interface conversion", func(t *testing.T) {
		TryCatch{
			Try: func() {
				var a interface{}
				a = 10
				fmt.Println(a.(string))
			},
			Catch: func(e error) {
				assert.NotNil(t, e)
				assert.Contains(t, e.Error(), "interface conversion")
			},
		}.Do()
	})
}
