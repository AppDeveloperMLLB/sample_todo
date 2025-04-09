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

func CreateClaims(uid uint, email string) *JwtCustomClaims {
	claims := &JwtCustomClaims{
		UID:   uid,
		Email: email,
	}
	return claims
}

func GenerateToken(uid uint, email string) (string, error) {
	claims := CreateClaims(uid, email)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString(SigningKey)
	if err != nil {
		return "", err
	}

	return t, nil
}
