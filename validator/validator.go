package validator

import "pkg.agungdwiprasetyo.com/candi/config/env"

// Validator instance
type Validator struct {
	*JSONSchemaValidator
	*StructValidator
}

// NewValidator instance
func NewValidator() *Validator {
	return &Validator{
		JSONSchemaValidator: NewJSONSchemaValidator(env.BaseEnv().JSONSchemaDir),
		StructValidator:     NewStructValidator(),
	}
}
