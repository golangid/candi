package main

const (
	templateSharedMiddlewareImpl = `// {{.Header}}

package shared

// WARNING, this file only for example

import (
	"context"
	"errors"
	"fmt"

	"{{.LibraryName}}/candishared"
)

// DefaultTokenValidator for token validator example
type DefaultTokenValidator struct {
}

// ValidateToken implement TokenValidator
func (v DefaultTokenValidator) ValidateToken(ctx context.Context, token string) (*candishared.TokenClaim, error) {
	var tokenClaim candishared.TokenClaim
	tokenClaim.Subject = "USER_ID"
	return &tokenClaim, nil
}

// DefaultACLPermissionChecker for acl permission checker example
type DefaultACLPermissionChecker struct {
}

// CheckPermission implement interfaces.ACLPermissionChecker
func (a DefaultACLPermissionChecker) CheckPermission(ctx context.Context, userID string, permissionCode string) (role string, err error) {
	if permissionCode != "resource.public" {
		return role, errors.New("Forbidden")
	}
	fmt.Printf("users with id '%s' can access resource with permission code '%s' (return role for this user is 'superadmin')\n", userID, permissionCode)
	return "superadmin", nil
}
`
)
