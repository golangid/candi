package domain

import "github.com/dgrijalva/jwt-go"

const (

	// HS256 const
	HS256 = "HS256"

	// RS256 const
	RS256 = "RS256"
)

// Claim for token claim data
type Claim struct {
	jwt.StandardClaims
	DeviceID string `json:"did"`
	User     struct {
		ID       string `json:"id"`
		Username string `json:"username"`
	} `json:"user"`
	Alg string `json:"-"`
}
