package controllers

import (
	"github.com/mllb/sampletodo/models"
)

//go:generate moq -out moq_test.go . TodoService AuthService
type TodoService interface {
	GetTodoList() ([]models.Todo, error)
	CreateTodo(title string, body string) error
	UpdateTodo(id uint, title string, body string) error
	// DeleteTodo(id string) error
	GetTodo(id uint) (models.Todo, error)
}

type AuthService interface {
	SignUp(email string, password string) error
	SignIn(email string, password string) (string, error)
	SignOut(token string) error
}
