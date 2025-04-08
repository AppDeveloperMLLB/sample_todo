package jwt

import (
	_ "embed"

	"github.com/golang-jwt/jwt/v5"
)

type JwtCustomClaims struct {
	UID   uint   `json:"uid"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}

//go:embed signing_key
var SigningKey []byte
