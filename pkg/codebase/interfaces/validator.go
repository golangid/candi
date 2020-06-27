package interfaces

// Validator abstract interface
type Validator interface {
	ValidateDocument(reference string, document []byte) error
}
