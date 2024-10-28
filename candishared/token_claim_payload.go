package candishared

import "github.com/golang-jwt/jwt"

// TokenClaim for token claim data
type TokenClaim struct {
	jwt.StandardClaims
	Role       string      `json:"role"`
	Additional any `json:"additional"`
}
