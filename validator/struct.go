package validator

import (
	"errors"
	"strings"

	validatorengine "github.com/go-playground/validator/v10"
	"github.com/golangid/candi/candishared"
)

// StructValidatorOptionFunc type
type StructValidatorOptionFunc func(*StructValidator)

// SetCoreStructValidatorOption option func
func SetCoreStructValidatorOption(additionalConfigFunc ...func(*validatorengine.Validate)) StructValidatorOptionFunc {
	return func(v *StructValidator) {
		ve := validatorengine.New()
		for _, additionalFunc := range additionalConfigFunc {
			additionalFunc(ve)
		}
		v.Validator = ve
	}
}

// StructValidator struct
type StructValidator struct {
	Validator *validatorengine.Validate
}

// NewStructValidator using go library
// https://github.com/go-playground/validator (all struct tags will be here)
// https://godoc.org/github.com/go-playground/validator (documentation using it)
// NewStructValidator function
func NewStructValidator(opts ...StructValidatorOptionFunc) *StructValidator {
	// set struct validator
	sv := &StructValidator{}
	for _, opt := range opts {
		opt(sv)
	}

	if sv.Validator == nil {
		sv.Validator = validatorengine.New()
	}

	return sv
}

// ValidateStruct function
func (v *StructValidator) ValidateStruct(data any) error {
	if err := v.Validator.Struct(data); err != nil {
		switch errs := err.(type) {
		case validatorengine.ValidationErrors:
			multiError := candishared.NewMultiError()
			for _, e := range errs {
				message := err.Error()
				multiError.Append(strings.ToLower(e.Field()), errors.New(message))
			}
			if multiError.HasError() {
				return multiError
			}
		default:
			return err
		}
	}

	return nil
}
