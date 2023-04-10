package validator

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/golangid/candi/candihelper"
	"github.com/golangid/gojsonschema"
)

// JSONSchemaValidator abstraction
type JSONSchemaValidator interface {
	ValidateDocument(schemaID string, documentSource interface{}) error
}

// JSONSchemaValidatorOptionFunc type
type JSONSchemaValidatorOptionFunc func(*jsonSchemaValidator)

// SetSchemaStorageJSONSchemaValidatorOption option func
func SetSchemaStorageJSONSchemaValidatorOption(s Storage) JSONSchemaValidatorOptionFunc {
	return func(v *jsonSchemaValidator) {
		v.schemaStorage = s
	}
}

// AddHideErrorListTypeJSONSchemaValidatorOption option func
func AddHideErrorListTypeJSONSchemaValidatorOption(descType ...string) JSONSchemaValidatorOptionFunc {
	return func(v *jsonSchemaValidator) {
		for _, e := range descType {
			v.notShowErrorListType[e] = struct{}{}
		}
	}
}

// jsonSchemaValidator validator
type jsonSchemaValidator struct {
	schemaStorage        Storage
	notShowErrorListType map[string]struct{}
}

// NewJSONSchemaValidator constructor
func NewJSONSchemaValidator(opts ...JSONSchemaValidatorOptionFunc) JSONSchemaValidator {
	v := &jsonSchemaValidator{
		schemaStorage: NewInMemStorage(os.Getenv(candihelper.WORKDIR) + "api/jsonschema"),
		notShowErrorListType: map[string]struct{}{
			"condition_else": {}, "condition_then": {},
		},
	}

	// overide with custom option
	for _, opt := range opts {
		opt(v)
	}
	return v
}

// ValidateDocument based on schema id
func (v *jsonSchemaValidator) ValidateDocument(schemaID string, documentSource interface{}) error {

	s, err := v.schemaStorage.Get(schemaID)
	if err != nil {
		return err
	}

	schema, err := gojsonschema.NewSchema(gojsonschema.NewStringLoader(
		strings.ReplaceAll(s, "{{WORKDIR}}", os.Getenv(candihelper.WORKDIR)),
	))
	if err != nil {
		return err
	}

	document := gojsonschema.NewBytesLoader(candihelper.ToBytes(documentSource))
	result, err := schema.Validate(document)
	if err != nil {
		return err
	}

	if !result.Valid() {
		multiError := candihelper.NewMultiError()
		for _, desc := range result.Errors() {
			if _, ok := v.notShowErrorListType[desc.Type()]; ok {
				continue
			}
			var field = desc.Field()
			if desc.Type() == "required" || desc.Type() == "additional_property_not_allowed" {
				field = fmt.Sprintf("%s.%s", field, desc.Details()["property"])
				field = strings.TrimPrefix(field, "(root).")
			}
			multiError.Append(field, errors.New(desc.Description()))
		}
		if multiError.HasError() {
			return multiError
		}
	}

	return nil
}
