package middleware

import (
	"encoding/base64"
	"errors"
	"strings"
)

// BasicAuth function basic auth
func (m *mw) BasicAuth(authorization string) error {
	authorizations := strings.Split(authorization, " ")
	if len(authorizations) != 2 {
		return errors.New("Unauthorized")
	}

	authType, val := authorizations[0], authorizations[1]
	if authType != "Basic" {
		return errors.New("Unauthorized")
	}

	isValid := func() bool {
		data, err := base64.StdEncoding.DecodeString(val)
		if err != nil {
			return false
		}

		decoded := strings.Split(string(data), ":")
		if len(decoded) < 2 {
			return false
		}
		username, password := decoded[0], decoded[1]

		if username != m.username || password != m.password {
			return false
		}

		return true
	}

	if !isValid() {
		return errors.New("Unauthorized")
	}

	return nil
}
