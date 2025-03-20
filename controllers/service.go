package controllers

import (
	"github.com/mllb/sampletodo/models"
)

//go:generate go run github.com/matryer/moq -out moq_test.go . TodoService
type TodoService interface {
	GetTodoList() ([]models.Todo, error)
	CreateTodo(title string, body string) error
	UpdateTodo(id uint, title string, body string) error
	// DeleteTodo(id string) error
	GetTodo(id uint) (models.Todo, error)
}
