package validator

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"agungdwiprasetyo.com/backend-microservices/pkg/helper"
	"github.com/xeipuuv/gojsonschema"
)

var notShowErrorListType = map[string]bool{
	"condition_else": true, "condition_then": true,
}

var inMemStorage = map[string]*gojsonschema.Schema{}

// loadJSONSchemaLocalFiles all json schema from given path
func loadJSONSchemaLocalFiles(serviceName string) error {

	here := fmt.Sprintf("api/%s/jsonschema/", serviceName)
	return filepath.Walk(here, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		fileName := info.Name()
		if strings.HasSuffix(fileName, ".json") {
			s, err := ioutil.ReadFile(p)
			if err != nil {
				return fmt.Errorf("%s: %v", fileName, err)
			}

			var data map[string]interface{}
			if err := json.Unmarshal(s, &data); err != nil {
				return fmt.Errorf("%s: %v", fileName, err)
			}
			id, ok := data["$id"].(string)
			if !ok {
				id = strings.Trim(strings.TrimSuffix(strings.TrimPrefix(p, here), ".json"), "/") // take filename without extension
			}
			inMemStorage[id], err = gojsonschema.NewSchema(gojsonschema.NewBytesLoader(s))
			if err != nil {
				return fmt.Errorf("%s: %v", fileName, err)
			}
		}
		return nil
	})
}

// JSONSchemaValidator validator
type JSONSchemaValidator struct {
}

// NewJSONSchemaValidator constructor
func NewJSONSchemaValidator(serviceName string) *JSONSchemaValidator {
	if err := loadJSONSchemaLocalFiles(serviceName); err != nil {
		log.Println("warning: failed load json schema")
	}
	return &JSONSchemaValidator{}
}

func (v *JSONSchemaValidator) getSchema(schemaID string) (schema *gojsonschema.Schema, err error) {
	s, ok := inMemStorage[schemaID]
	if !ok {
		return nil, fmt.Errorf("schema '%s' not found", schemaID)
	}

	return s, nil
}

// ValidateDocument based on schema id
func (v *JSONSchemaValidator) ValidateDocument(schemaID string, documentSource []byte) error {

	multiError := helper.NewMultiError()

	schema, err := v.getSchema(schemaID)
	if err != nil {
		return err
	}

	document := gojsonschema.NewBytesLoader(documentSource)

	result, err := schema.Validate(document)
	if err != nil {
		multiError.Append("validateInput", errors.New("Gagal memuat input data"))
		return multiError
	}

	if !result.Valid() {
		for _, desc := range result.Errors() {
			if notShowErrorListType[desc.Type()] {
				continue
			}
			var field = desc.Field()
			if desc.Type() == "required" || desc.Type() == "additional_property_not_allowed" {
				field = fmt.Sprintf("%s.%s", field, desc.Details()["property"])
				field = strings.TrimPrefix(field, "(root).")
			}
			multiError.Append(field, errors.New(desc.Description()))
		}
	}

	if multiError.HasError() {
		return multiError
	}

	return nil
}
