package candihelper

// URLQueryGetter abstraction
type URLQueryGetter interface {
	Get(key string) string
}

// MultiError abstract interface
type MultiError interface {
	Append(key string, err error) MultiError
	HasError() bool
	IsNil() bool
	Clear()
	ToMap() map[string]string
	Merge(MultiError) MultiError
	Error() string
}

// FilterStreamer abstract interface
type FilterStreamer interface {
	GetPage() int
	IncrPage()
	GetLimit() int
}
