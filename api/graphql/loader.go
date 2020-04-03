package graphql

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// LoadSchema graphql from file
func LoadSchema(serviceName string) string {

	var schema strings.Builder
	here := fmt.Sprintf("%s/api/graphql/%s/", os.Getenv("APP_PATH"), serviceName)

	filepath.Walk(here, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			panic(err)
		}
		if info.IsDir() {
			return nil
		}

		fileName := info.Name()
		if strings.HasSuffix(fileName, ".graphql") {
			s, err := ioutil.ReadFile(p)
			if err != nil {
				panic(err)
			}
			schema.Write(s)
		}
		return nil
	})

	return schema.String()
}
