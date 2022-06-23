package validator

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Storage abstraction
type Storage interface {
	Get(schemaID string) (string, error)
	Store(schemaID string, schema string) error
}

type (
	localStorage struct {
		kv map[string]string
	}

	inMemStorage struct {
		storage map[string]string
	}

	sqlStorage struct {
		db *sql.DB
	}
)

// NewLocalStorage read from file
func NewLocalStorage(rootDir string) Storage {
	ls := &localStorage{
		kv: make(map[string]string),
	}
	filepath.Walk(rootDir, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		fileName := info.Name()
		if strings.HasSuffix(fileName, ".json") {
			s, err := os.ReadFile(p)
			if err != nil {
				return fmt.Errorf("%s: %v", fileName, err)
			}

			var data map[string]interface{}
			if err := json.Unmarshal(s, &data); err != nil {
				return fmt.Errorf("%s: %v", fileName, err)
			}
			id, ok := data["$id"].(string)
			if !ok {
				id = strings.Trim(strings.TrimSuffix(strings.TrimPrefix(p, rootDir), ".json"), "/") // take filename without extension
			}
			ls.Store(id, p)
		}
		return nil
	})

	return ls
}

func (l *localStorage) Get(schemaID string) (string, error) {
	path, ok := l.kv[schemaID]
	if !ok {
		return "", fmt.Errorf("schema '%s' not found", schemaID)
	}
	schema, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("schema '%s' not found", schemaID)
	}
	return string(schema), nil
}

func (l *localStorage) Store(schemaID string, schema string) error {
	l.kv[schemaID] = schema
	return nil
}

// NewInMemStorage constructor
func NewInMemStorage(rootDir string) Storage {
	inMem := &inMemStorage{
		storage: make(map[string]string),
	}

	filepath.Walk(rootDir, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		fileName := info.Name()
		if strings.HasSuffix(fileName, ".json") {
			s, err := os.ReadFile(p)
			if err != nil {
				return fmt.Errorf("%s: %v", fileName, err)
			}

			var data map[string]interface{}
			if err := json.Unmarshal(s, &data); err != nil {
				return fmt.Errorf("%s: %v", fileName, err)
			}
			id, ok := data["$id"].(string)
			if !ok {
				id = strings.Trim(strings.TrimSuffix(strings.TrimPrefix(p, rootDir), ".json"), "/") // take filename without extension
			}
			inMem.Store(id, string(s))
		}
		return nil
	})

	return inMem
}

func (i *inMemStorage) Get(schemaID string) (string, error) {
	schema, ok := i.storage[schemaID]
	if !ok {
		return "", fmt.Errorf("schema '%s' not found", schemaID)
	}
	return schema, nil
}

func (i *inMemStorage) Store(schemaID string, schema string) error {
	i.storage[schemaID] = schema
	return nil
}
