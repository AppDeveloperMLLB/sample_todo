package repositories

import (
	"github.com/mllb/sampletodo/models"
	"gorm.io/gorm"
)

type TodoRepository struct {
	db *gorm.DB
}

func NewTodoRepository(db *gorm.DB) *TodoRepository {
	return &TodoRepository{
		db: db,
	}
}

func (r *TodoRepository) GetTodoList() ([]models.Todo, error) {
	var todoList []models.Todo
	if err :=
		r.db.Find(&todoList).Error; err != nil {
		return nil, err
	}
	return todoList, nil
}

func (r *TodoRepository) CreateTodo(todo models.Todo) error {
	if err := r.db.Create(&todo).Error; err != nil {
		return err
	}
	return nil
}

func (r *TodoRepository) GetTodo(id uint) (models.Todo, error) {
	var todo models.Todo
	result := r.db.Where("id = ?", id).Take(&todo)
	if result.Error != nil {
		return models.Todo{}, result.Error
	}
	return todo, nil
}

func (r *TodoRepository) UpdateTodo(todo models.Todo) error {
	if err := r.db.Save(&todo).Error; err != nil {
		return err
	}
	return nil
}
