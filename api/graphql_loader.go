package api

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// LoadGraphQLSchema graphql from file
func LoadGraphQLSchema(serviceName string) string {

	var schema strings.Builder
	here := fmt.Sprintf("api/%s/graphql/", serviceName)

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
