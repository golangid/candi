package validator

import "pkg.agungdwiprasetyo.com/gendon/config"

// Validator instance
type Validator struct {
	*JSONSchemaValidator
	*StructValidator
}

// NewValidator instance
func NewValidator() *Validator {
	return &Validator{
		JSONSchemaValidator: NewJSONSchemaValidator(config.BaseEnv().JSONSchemaDir),
		StructValidator:     NewStructValidator(),
	}
}
