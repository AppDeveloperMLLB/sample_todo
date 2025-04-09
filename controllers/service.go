package controllers

import (
	"github.com/mllb/sampletodo/models"
)

//go:generate moq -out moq_test.go . TodoService AuthService
type TodoService interface {
	GetTodoList(uid uint) ([]models.Todo, error)
	CreateTodo(uid uint, title string, body string) error
	UpdateTodo(uid uint, id uint, title string, body string) error
	// DeleteTodo(id string) error
	GetTodo(uid uint, id uint) (models.Todo, error)
}

type AuthService interface {
	SignUp(email string, password string) error
	SignIn(email string, password string) (string, error)
	SignOut(token string) error
}
