package validator

import (
	"errors"
	"strings"

	validatorengine "github.com/go-playground/validator/v10"
	"github.com/golangid/candi/candihelper"
)

// StructValidator abstraction
type StructValidator interface {
	ValidateStruct(data interface{}) error
}

// StructValidatorOptionFunc type
type StructValidatorOptionFunc func(*structValidator)

// SetCoreStructValidatorOption option func
func SetCoreStructValidatorOption(additionalConfigFunc ...func(*validatorengine.Validate)) StructValidatorOptionFunc {
	return func(v *structValidator) {
		ve := validatorengine.New()
		for _, additionalFunc := range additionalConfigFunc {
			additionalFunc(ve)
		}
		v.validator = ve
	}
}

// structValidator struct
type structValidator struct {
	validator *validatorengine.Validate
}

// NewStructValidator using go library
// https://github.com/go-playground/validator (all struct tags will be here)
// https://godoc.org/github.com/go-playground/validator (documentation using it)
// NewStructValidator function
func NewStructValidator(opts ...StructValidatorOptionFunc) StructValidator {
	// set struct validator
	sv := &structValidator{}
	for _, opt := range opts {
		opt(sv)
	}

	if sv.validator == nil {
		sv.validator = validatorengine.New()
	}

	return sv
}

// ValidateStruct function
func (v *structValidator) ValidateStruct(data interface{}) error {
	if err := v.validator.Struct(data); err != nil {
		switch errs := err.(type) {
		case validatorengine.ValidationErrors:
			multiError := candihelper.NewMultiError()
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
