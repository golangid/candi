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

	"github.com/xeipuuv/gojsonschema"
	"pkg.agungdp.dev/candi/candihelper"
)

var notShowErrorListType = map[string]bool{
	"condition_else": true, "condition_then": true,
}

var (
	inMemStorage = map[string]*gojsonschema.Schema{}
	inMemJSON    = map[string]interface{}{}
)

// loadJSONSchemaLocalFiles all json schema from given path
func loadJSONSchemaLocalFiles(path string) error {

	return filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
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
				id = strings.Trim(strings.TrimSuffix(strings.TrimPrefix(p, path), ".json"), "/") // take filename without extension
			}
			inMemStorage[id], err = gojsonschema.NewSchema(gojsonschema.NewBytesLoader(s))
			if err != nil {
				return fmt.Errorf("%s: %v", fileName, err)
			}
			inMemJSON[id] = data
		}
		return nil
	})
}

// JSONSchemaValidator validator
type JSONSchemaValidator struct {
}

// NewJSONSchemaValidator constructor
func NewJSONSchemaValidator(schemaRootPath string) *JSONSchemaValidator {
	if err := loadJSONSchemaLocalFiles(schemaRootPath); err != nil {
		log.Println(candihelper.StringYellow("Validator: warning, failed load json schema in path " + schemaRootPath))
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
func (v *JSONSchemaValidator) ValidateDocument(schemaID string, documentSource interface{}) error {

	schema, err := v.getSchema(schemaID)
	if err != nil {
		return err
	}
	jsonObj, _ := inMemJSON[schemaID]

	document := gojsonschema.NewBytesLoader(candihelper.ToBytes(documentSource))
	result, err := schema.Validate(document)
	if err != nil {
		return err
	}

	multiError := candihelper.NewMultiError()
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
			msg, found := getMessage(jsonObj, field, "message")
			if found {
				multiError.Append(field, errors.New(msg))
			} else {
				multiError.Append(field, errors.New(desc.Description()))
			}
		}
	}

	if multiError.HasError() {
		return multiError
	}

	return nil
}

func getMessage(obj interface{}, key, messageKey string) (string, bool) {
	switch t := obj.(type) {
	case map[string]interface{}:
		if v, ok := t[key]; ok {
			if msg, ok := v.(map[string]interface{})[messageKey]; ok {
				return fmt.Sprintf("%v", msg), ok
			}
			if msg, ok := getMessage(v, key, messageKey); ok {
				return msg, ok
			}
		}
		for _, v := range t {
			if msg, ok := getMessage(v, key, messageKey); ok {
				return msg, ok
			}
		}
	}
	return "", false
}
