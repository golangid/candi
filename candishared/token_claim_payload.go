package candishared

import "github.com/golang-jwt/jwt/v5"

// TokenClaim for token claim data
type TokenClaim struct {
	jwt.RegisteredClaims
	Role       string `json:"role"`
	Additional any    `json:"additional"`
}
