package validator

// OptionFunc type
type OptionFunc func(*Validator)

// SetJSONSchemaValidator option func
func SetJSONSchemaValidator(jsonSchema JSONSchemaValidator) OptionFunc {
	return func(v *Validator) {
		v.jsonSchema = jsonSchema
	}
}

// SetStructValidator option func
func SetStructValidator(structValidator StructValidator) OptionFunc {
	return func(v *Validator) {
		v.structValidator = structValidator
	}
}

// Validator instance
type Validator struct {
	jsonSchema      JSONSchemaValidator
	structValidator StructValidator
}

// NewValidator constructor, using jsonschema & struct validator (github.com/go-playground/validator),
// jsonschema source file load from "api/jsonschema" directory
func NewValidator(opts ...OptionFunc) *Validator {
	v := &Validator{
		jsonSchema:      NewJSONSchemaValidator(),
		structValidator: NewStructValidator(),
	}

	for _, opt := range opts {
		opt(v)
	}
	return v
}

// ValidateDocument method using jsonschema with input is json source
func (v *Validator) ValidateDocument(reference string, document interface{}) error {
	return v.jsonSchema.ValidateDocument(reference, document)
}

// ValidateStruct method, rules from struct tag using github.com/go-playground/validator
func (v *Validator) ValidateStruct(data interface{}) error {
	return v.structValidator.ValidateStruct(data)
}
