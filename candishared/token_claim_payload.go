package candishared

import "github.com/dgrijalva/jwt-go"

// TokenClaim for token claim data
type TokenClaim struct {
	jwt.StandardClaims
	DeviceID string `json:"did"`
	User     struct {
		ID       string `json:"id"`
		Username string `json:"username"`
	} `json:"user"`
	Alg string `json:"-"`
}
