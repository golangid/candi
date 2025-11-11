package validator

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/golangid/candi/candihelper"
	"github.com/golangid/candi/candishared"
	"github.com/golangid/gojsonschema"
)

// JSONSchemaValidatorOptionFunc type
type JSONSchemaValidatorOptionFunc func(*JSONSchemaValidator)

// SetSchemaStorageJSONSchemaValidatorOption option func
func SetSchemaStorageJSONSchemaValidatorOption(s Storage) JSONSchemaValidatorOptionFunc {
	return func(v *JSONSchemaValidator) {
		v.SchemaStorage = s
	}
}

// AddHideErrorListTypeJSONSchemaValidatorOption option func
func AddHideErrorListTypeJSONSchemaValidatorOption(descType ...string) JSONSchemaValidatorOptionFunc {
	return func(v *JSONSchemaValidator) {
		for _, e := range descType {
			v.notShowErrorListType[e] = struct{}{}
		}
	}
}

// JSONSchemaValidator validator
type JSONSchemaValidator struct {
	SchemaStorage        Storage
	notShowErrorListType map[string]struct{}
}

// NewJSONSchemaValidator constructor
func NewJSONSchemaValidator(opts ...JSONSchemaValidatorOptionFunc) *JSONSchemaValidator {
	v := &JSONSchemaValidator{
		SchemaStorage: NewInMemStorage(os.Getenv(candihelper.WORKDIR) + "api/jsonschema"),
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
func (v *JSONSchemaValidator) ValidateDocument(schemaSource string, documentSource any) error {
	s, err := v.SchemaStorage.Get(schemaSource)
	if err != nil {
		return err
	}
	schemaSource = strings.ReplaceAll(s, "{{WORKDIR}}", os.Getenv(candihelper.WORKDIR))

	schema, err := gojsonschema.NewSchema(gojsonschema.NewStringLoader(schemaSource))
	if err != nil {
		return err
	}

	document := gojsonschema.NewBytesLoader(candihelper.ToBytes(documentSource))
	result, err := schema.Validate(document)
	if err != nil {
		return err
	}

	if !result.Valid() {
		multiError := candishared.NewMultiError()
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
