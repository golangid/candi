package validator

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/golangid/candi/candihelper"
	"github.com/golangid/gojsonschema"
)

// JSONSchemaValidator validator
type JSONSchemaValidator struct {
	SchemaStorage        Storage
	NotShowErrorListType map[string]bool
}

// NewJSONSchemaValidator constructor
func NewJSONSchemaValidator(schemaRootPath string) *JSONSchemaValidator {
	v := &JSONSchemaValidator{
		SchemaStorage: NewInMemStorage(schemaRootPath),
	}
	v.NotShowErrorListType = map[string]bool{
		"condition_else": true, "condition_then": true,
	}
	return v
}

// ValidateDocument based on schema id
func (v *JSONSchemaValidator) ValidateDocument(schemaID string, documentSource interface{}) error {

	s, err := v.SchemaStorage.Get(schemaID)
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
			if v.NotShowErrorListType[desc.Type()] {
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
