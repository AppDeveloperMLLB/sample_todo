package controllers

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
)

type SignInRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type SignUpRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type AuthController struct {
	service AuthService
}

func NewAuthController(s AuthService) *AuthController {
	return &AuthController{
		service: s,
	}
}

func (con *AuthController) SignUp(c echo.Context) error {
	req := new(SignUpRequest)
	if err := c.Bind(&req); err != nil {
		fmt.Println(err)
		fmt.Printf("%+v\n", req)
		return c.JSON(http.StatusBadRequest, map[string]string{"status": "Bad Request"})
	}
	v := validator.New()
	if err := v.Struct(req); err != nil {
		fmt.Println(err)
		fmt.Printf("%+v\n", req)
		return c.JSON(http.StatusBadRequest, map[string]string{"status": "Bad Request"})
	}
	err := con.service.SignUp(req.Email, req.Password)
	if err != nil {
		fmt.Println(err)
		fmt.Printf("%+v\n", req)
		return c.JSON(http.StatusConflict, map[string]string{"status": "Conflict"})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "OK"})
}

func (con *AuthController) SignIn(c echo.Context) error {
	req := new(SignInRequest)
	if err := c.Bind(&req); err != nil {
		fmt.Println(err)
		fmt.Printf("%+v\n", req)
		return c.JSON(http.StatusBadRequest, map[string]string{"status": "Bad Request"})
	}
	v := validator.New()
	if err := v.Struct(req); err != nil {
		fmt.Println(err)
		fmt.Printf("%+v\n", req)
		return c.JSON(http.StatusBadRequest, map[string]string{"status": "Bad Request"})
	}
	token, err := con.service.SignIn(req.Email, req.Password)
	if err != nil {
		fmt.Println(err)
		fmt.Printf("%+v\n", req)
		return c.JSON(http.StatusUnauthorized, map[string]string{"status": "Unauthorized"})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "OK", "token": token})
}

func (con *AuthController) SignOut(c echo.Context) error {
	return c.String(http.StatusOK, "SignOut")
}
