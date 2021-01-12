package main

const (
	templateSharedTokenValidator = `// {{.Header}}

package shared

import (
	"context"

	"{{.PackageName}}/candishared"
)

// DefaultTokenValidator for token validator
type DefaultTokenValidator struct {
}

// ValidateToken implement TokenValidator
func (v *DefaultTokenValidator) ValidateToken(ctx context.Context, token string) (*candishared.TokenClaim, error) {
	return &candishared.TokenClaim{}, nil
}
`
)
