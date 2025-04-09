package controllers

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	appjwt "github.com/mllb/sampletodo/jwt"
)

func getUserID(c echo.Context) (uint, error) {
	token, ok := c.Get("user").(*jwt.Token)
	if !ok {
		fmt.Println("error getting token from context")
		return 0, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*appjwt.JwtCustomClaims)
	if !ok {
		fmt.Println("error getting claims from token")
		return 0, errors.New("invalid token")
	}

	return claims.UID, nil
}
