package candihelper

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// LoadAllFile from path
func LoadAllFile(path, formatFile string) []byte {
	buff := bytes.NewBuffer(make([]byte, 128))
	buff.Reset()
	filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			panic(err)
		}
		if info.IsDir() {
			return nil
		}

		fileName := info.Name()
		if strings.HasSuffix(fileName, formatFile) {
			s, err := os.ReadFile(p)
			if err != nil {
				panic(err)
			}
			buff.Write(s)
		}
		return nil
	})

	return buff.Bytes()
}

// LoadAllFileFromFS helper for load all file from file system
func LoadAllFileFromFS(fileSystem fs.FS, sourcePath, formatFile string) []byte {
	buff := bytes.NewBuffer(make([]byte, 128))
	buff.Reset()
	fs.WalkDir(fileSystem, sourcePath, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			panic(err)
		}
		if info.IsDir() {
			return nil
		}

		fileName := info.Name()
		if strings.HasSuffix(fileName, formatFile) {
			s, err := fs.ReadFile(fileSystem, path)
			if err != nil {
				panic(err)
			}
			buff.Write(s)
		}
		return nil
	})

	return buff.Bytes()
}
