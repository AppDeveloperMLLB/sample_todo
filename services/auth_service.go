package services

import (
	"crypto/sha256"
	"errors"
	"fmt"

	jwtgo "github.com/golang-jwt/jwt/v5"
	"github.com/mllb/sampletodo/jwt"
	"github.com/mllb/sampletodo/repositories"
)

type AuthService struct {
	repo repositories.AuthRepository
}

func NewAuthService(repo repositories.AuthRepository) *AuthService {
	return &AuthService{
		repo: repo,
	}
}

func (s *AuthService) SignUp(email string, password string) error {
	// Check if the user already exists
	_, err := s.repo.FindUser(email)
	if err == nil {
		return errors.New("User already exits") // User already exists
	}

	p := []byte(password)
	hashedPassword := sha256.Sum256(p)
	hashedPasswordStr := fmt.Sprintf("%x", hashedPassword)
	err = s.repo.SignUp(email, hashedPasswordStr)
	if err != nil {
		return err
	}
	return nil
}

func (s *AuthService) SignIn(email string, password string) (string, error) {
	user, err := s.repo.FindUser(email)
	if err != nil {
		return "", err
	}

	p := []byte(password)
	hashedPassword := sha256.Sum256(p)
	hashedPasswordStr := fmt.Sprintf("%x", hashedPassword)
	if user.Password != hashedPasswordStr {
		return "", errors.New("Invalid password")
	}

	claims := &jwt.JwtCustomClaims{
		UID:   user.ID,
		Email: user.Email,
	}

	token := jwtgo.NewWithClaims(jwtgo.SigningMethodHS256, claims)
	t, err := token.SignedString(jwt.SigningKey)
	if err != nil {
		return "", err
	}

	return t, nil
}
func (s *AuthService) SignOut(token string) error {
	// err := s.repo.SignOut(token)
	// if err != nil {
	// 	return err
	// }
	return nil
}
