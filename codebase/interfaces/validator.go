package interfaces

// Validator abstract interface
type Validator interface {
	// ValidateDocument method using jsonschema with input is json source
	ValidateDocument(reference string, document any) error

	// ValidateStruct method, rules from struct tag using github.com/go-playground/validator
	ValidateStruct(data any) error
}
