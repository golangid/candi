package validator

import (
	"os"

	"pkg.agungdp.dev/candi/candihelper"
)

// Validator instance
type Validator struct {
	*JSONSchemaValidator
	*StructValidator
}

// NewValidator constructor, using jsonschema & struct validator (github.com/go-playground/validator),
// jsonschema source file load from JSON_SCHEMA_DIR environment
func NewValidator() *Validator {
	return &Validator{
		JSONSchemaValidator: NewJSONSchemaValidator(os.Getenv(candihelper.WORKDIR) + "api/jsonschema"),
		StructValidator:     NewStructValidator(),
	}
}
