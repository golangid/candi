package candihelper

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
)

// LoadAllFile from path
func LoadAllFile(path, formatFile string) []byte {
	var buff bytes.Buffer
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
