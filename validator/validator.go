package validator

// OptionFunc type
type OptionFunc func(*Validator)

// SetJSONSchemaValidator option func
func SetJSONSchemaValidator(jsonSchema *JSONSchemaValidator) OptionFunc {
	return func(v *Validator) {
		v.JSONSchema = jsonSchema
	}
}

// SetStructValidator option func
func SetStructValidator(structValidator *StructValidator) OptionFunc {
	return func(v *Validator) {
		v.StructValidator = structValidator
	}
}

// Validator instance
type Validator struct {
	JSONSchema      *JSONSchemaValidator
	StructValidator *StructValidator
}

// NewValidator constructor, using jsonschema & struct validator (github.com/go-playground/validator),
// jsonschema source file load from "api/jsonschema" directory
func NewValidator(opts ...OptionFunc) *Validator {
	v := &Validator{
		JSONSchema:      NewJSONSchemaValidator(),
		StructValidator: NewStructValidator(),
	}

	for _, opt := range opts {
		opt(v)
	}
	return v
}

// ValidateDocument method using jsonschema with input is json source
func (v *Validator) ValidateDocument(reference string, document any) error {
	return v.JSONSchema.ValidateDocument(reference, document)
}

// ValidateStruct method, rules from struct tag using github.com/go-playground/validator
func (v *Validator) ValidateStruct(data any) error {
	return v.StructValidator.ValidateStruct(data)
}
