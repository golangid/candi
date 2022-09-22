package logger

import "io"

// Option for init logger option
type (
	Option struct {
		MultiWriter []io.Writer
	}

	// OptionFunc func
	OptionFunc func(*Option)
)

// OptionAddWriter option func
func OptionAddWriter(w io.Writer) OptionFunc {
	return func(o *Option) {
		o.MultiWriter = append(o.MultiWriter, w)
	}
}

// OptionSetWriter option func, overide all log writer
func OptionSetWriter(w ...io.Writer) OptionFunc {
	return func(o *Option) {
		o.MultiWriter = w
	}
}
