package validator

import (
	"os"

	"github.com/golangid/candi/candihelper"
)

// Validator instance
type Validator struct {
	*JSONSchemaValidator
	*StructValidator
}

// NewValidator constructor, using jsonschema & struct validator (github.com/go-playground/validator),
// jsonschema source file load from "api/jsonschema" directory
func NewValidator() *Validator {
	return &Validator{
		JSONSchemaValidator: NewJSONSchemaValidator(os.Getenv(candihelper.WORKDIR) + "api/jsonschema"),
		StructValidator:     NewStructValidator(),
	}
}
