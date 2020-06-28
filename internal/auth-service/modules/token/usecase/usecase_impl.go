package usecase

import (
	"context"
	"crypto/rsa"
	"time"

	"agungdwiprasetyo.com/backend-microservices/internal/auth-service/modules/token/domain"
	"agungdwiprasetyo.com/backend-microservices/pkg/helper"
	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
	"github.com/dgrijalva/jwt-go"
)

// tokenUsecaseImpl repo
type tokenUsecaseImpl struct {
	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
}

// NewTokenUsecase constructor
func NewTokenUsecase(publicKey *rsa.PublicKey, privateKey *rsa.PrivateKey) TokenUsecase {
	return &tokenUsecaseImpl{
		publicKey:  publicKey,
		privateKey: privateKey,
	}
}

// Generate token
func (uc *tokenUsecaseImpl) Generate(ctx context.Context, payload *shared.TokenClaim, expired time.Duration) <-chan shared.Result {
	output := make(chan shared.Result)

	go func() {
		defer close(output)

		now := time.Now()
		exp := now.Add(expired)

		var key interface{}
		var token = new(jwt.Token)
		if payload.Alg == domain.HS256 {
			token = jwt.New(jwt.SigningMethodHS256)
			key = []byte(helper.TokenKey)
		} else {
			token = jwt.New(jwt.SigningMethodRS256)
			key = uc.privateKey
		}
		claims := jwt.MapClaims{
			"iss":  "agungdwiprasetyo",
			"exp":  exp.Unix(),
			"iat":  now.Unix(),
			"did":  payload.DeviceID,
			"aud":  payload.Audience,
			"jti":  payload.Id,
			"user": payload.User,
		}
		token.Claims = claims

		tokenString, err := token.SignedString(key)
		if err != nil {
			output <- shared.Result{Error: err}
			return
		}

		output <- shared.Result{Data: tokenString}
	}()

	return output
}

// Refresh token
func (uc *tokenUsecaseImpl) Refresh(ctx context.Context, token string) <-chan shared.Result {
	output := make(chan shared.Result)

	go func() {
		defer close(output)
	}()

	return output
}

// Validate token
func (uc *tokenUsecaseImpl) Validate(ctx context.Context, tokenString string) <-chan shared.Result {
	output := make(chan shared.Result)

	go func() {
		defer close(output)

		tokenParse, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			checkAlg, _ := shared.GetValueFromContext(ctx, shared.ContextKey("tokenAlg")).(string)
			if checkAlg == domain.HS256 {
				return []byte(helper.TokenKey), nil
			}
			return uc.publicKey, nil
		})

		var errToken error
		switch ve := err.(type) {
		case *jwt.ValidationError:
			if ve.Errors == jwt.ValidationErrorExpired {
				errToken = helper.ErrTokenExpired
			} else {
				errToken = helper.ErrTokenFormat
			}
		}

		if errToken != nil {
			output <- shared.Result{Error: errToken}
			return
		}

		if !tokenParse.Valid {
			output <- shared.Result{Error: helper.ErrTokenFormat}
			return
		}

		mapClaims, _ := tokenParse.Claims.(jwt.MapClaims)

		var tokenClaim domain.Claim
		tokenClaim.DeviceID, _ = mapClaims["did"].(string)
		tokenClaim.Audience, _ = mapClaims["aud"].(string)
		tokenClaim.Id, _ = mapClaims["jti"].(string)
		userData, _ := mapClaims["user"].(map[string]interface{})
		tokenClaim.User.ID, _ = userData["id"].(string)
		tokenClaim.User.Username, _ = userData["username"].(string)

		output <- shared.Result{Data: &tokenClaim}
	}()

	return output
}

// Revoke token
func (uc *tokenUsecaseImpl) Revoke(ctx context.Context, token string) <-chan shared.Result {
	output := make(chan shared.Result)

	go func() {
		defer close(output)
	}()

	return output
}
