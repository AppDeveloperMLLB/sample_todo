package services

import (
	"github.com/mllb/sampletodo/models"
	"github.com/mllb/sampletodo/repositories"
)

type TodoService struct {
	repo repositories.TodoRepository
}

func NewTodoService(repo repositories.TodoRepository) *TodoService {
	return &TodoService{
		repo: repo,
	}
}

func (s *TodoService) GetTodoList(uid uint) ([]models.Todo, error) {
	todoList, err := s.repo.GetTodoList()
	if err != nil {
		return nil, err
	}
	return todoList, nil
}

func (s *TodoService) CreateTodo(uid uint, title string, body string) error {
	err := s.repo.CreateTodo(models.Todo{Title: title, Body: body})
	if err != nil {
		return err
	}
	return nil
}

func (s *TodoService) GetTodo(uid uint, id uint) (models.Todo, error) {
	todo, err := s.repo.GetTodo(id)
	if err != nil {
		return models.Todo{}, err
	}
	return todo, nil
}

func (s *TodoService) UpdateTodo(uid uint, id uint, title string, body string) error {
	todo, err := s.repo.GetTodo(id)
	if err != nil {
		return err
	}

	err = s.repo.UpdateTodo(models.Todo{ID: id, Title: title, Body: body, CreatedAt: todo.CreatedAt})
	if err != nil {
		return err
	}
	return nil
}
