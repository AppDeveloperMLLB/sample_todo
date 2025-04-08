package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/mllb/sampletodo/models"
	"gorm.io/gorm"
)

type CreateTodoRequest struct {
	Title string `json:"title" validate:"required"`
	Body  string `json:"body" validate:"required"`
}

type GetTodoResponse struct {
	Todos []models.Todo `json:"result"`
}

type TodoController struct {
	service TodoService
}

func NewTodoController(s TodoService) *TodoController {
	return &TodoController{
		service: s,
	}
}

func (con *TodoController) GetTodoList(c echo.Context) error {
	todoList, err := con.service.GetTodoList()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	res := GetTodoResponse{Todos: todoList}
	return c.JSON(http.StatusOK, &res)
}

func (con *TodoController) CreateTodo(c echo.Context) error {
	req := new(CreateTodoRequest)
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
	err := con.service.CreateTodo(req.Title, req.Body)
	if err != nil {
		fmt.Println(err)
		fmt.Printf("%+v\n", req)
		return c.JSON(http.StatusInternalServerError, map[string]string{"status": "Internal Server Error"})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "Success"})
}

func (con *TodoController) GetTodo(c echo.Context) error {
	idStr := c.Param("id")
	if idStr == "" {
		return c.String(http.StatusBadRequest, "Bad Request")
	}
	id64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid ID format")
	}

	// uint型にキャストするのだ
	id := uint(id64)

	todo, err := con.service.GetTodo(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.String(http.StatusNotFound, "Not Found")
		}
		fmt.Println(err)
		return c.String(http.StatusInternalServerError, "Internal Server Error")
	}

	return c.JSON(http.StatusOK, todo)
}

func (con *TodoController) UpdateTodo(c echo.Context) error {
	idStr := c.Param("id")
	if idStr == "" {
		return c.String(http.StatusBadRequest, "Bad Request")
	}
	id64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid ID format")
	}

	// uint型にキャストするのだ
	id := uint(id64)

	req := new(CreateTodoRequest)
	if err := c.Bind(&req); err != nil {
		fmt.Println(err)
		fmt.Printf("%+v\n", req)
		return c.String(http.StatusBadRequest, "Bad Request")
	}
	v := validator.New()
	if err := v.Struct(req); err != nil {
		fmt.Println(err)
		fmt.Printf("%+v\n", req)
		return c.String(http.StatusBadRequest, "Bad Request")
	}

	err = con.service.UpdateTodo(id, req.Title, req.Body)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.String(http.StatusNotFound, "Not Found")
		}
		fmt.Println(err)
		return c.String(http.StatusInternalServerError, "Internal Server Error")
	}

	return c.String(http.StatusOK, "Success")
}
