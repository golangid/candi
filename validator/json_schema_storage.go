package validator

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
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
	fileLocalStorage struct {
		kv map[string]string
	}

	inMemStorage struct {
		storage map[string]string
	}

	fsStorage struct {
		sourceMap map[string]string
		storage   fs.FS
	}

	sqlStorage struct {
		db *sql.DB
	}
)

// NewFileLocalStorage read from file
func NewFileLocalStorage(schemaLocationDir string) Storage {
	ls := &fileLocalStorage{
		kv: make(map[string]string),
	}
	filepath.Walk(schemaLocationDir, func(p string, info os.FileInfo, err error) error {
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
				id = strings.Trim(strings.TrimSuffix(strings.TrimPrefix(p, schemaLocationDir), ".json"), "/") // take filename without extension
			}
			ls.Store(id, p)
		}
		return nil
	})

	return ls
}

func (l *fileLocalStorage) Get(schemaID string) (string, error) {
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

func (l *fileLocalStorage) Store(schemaID string, schema string) error {
	l.kv[schemaID] = schema
	return nil
}

// NewInMemStorage constructor
func NewInMemStorage(schemaLocationDir string) Storage {
	inMem := &inMemStorage{
		storage: make(map[string]string),
	}

	filepath.Walk(schemaLocationDir, func(p string, info os.FileInfo, err error) error {
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
				id = strings.Trim(strings.TrimSuffix(strings.TrimPrefix(p, schemaLocationDir), ".json"), "/") // take filename without extension
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

// NewFileSystemStorage constructor
func NewFileSystemStorage(fileSystem fs.FS, rootPath string) Storage {
	storage := &fsStorage{
		sourceMap: make(map[string]string),
		storage:   fileSystem,
	}

	fs.WalkDir(fileSystem, rootPath, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			log.Panicf("json schema: %s", err.Error())
		}
		if info.IsDir() {
			return nil
		}

		fileName := info.Name()
		if strings.HasSuffix(fileName, ".json") {
			s, err := fs.ReadFile(fileSystem, path)
			if err != nil {
				return fmt.Errorf("%s: %v", fileName, err)
			}

			var data map[string]interface{}
			if err := json.Unmarshal(s, &data); err != nil {
				return fmt.Errorf("%s: %v", fileName, err)
			}

			id, ok := data["$id"].(string)
			if !ok {
				id = strings.Trim(strings.TrimSuffix(strings.TrimPrefix(path, rootPath), ".json"), "/")
			}
			storage.sourceMap[id] = path
		}
		return nil
	})

	fmt.Println(storage.sourceMap)
	return storage
}

func (i *fsStorage) Get(schemaID string) (string, error) {
	schemaPath, ok := i.sourceMap[schemaID]
	if !ok {
		return "", fmt.Errorf("schema '%s' not found", schemaID)
	}

	s, err := fs.ReadFile(i.storage, schemaPath)
	return string(s), err
}

func (i *fsStorage) Store(schemaID string, schema string) error {
	return nil
}
