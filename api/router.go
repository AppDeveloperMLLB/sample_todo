package api

import (
	"crypto/sha256"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mllb/sampletodo/controllers"
	"github.com/mllb/sampletodo/repositories"
	"github.com/mllb/sampletodo/services"
	"gorm.io/gorm"
)

func NewRouter(db *gorm.DB) *echo.Echo {
	todoRepo := repositories.NewTodoRepository(db)
	todoService := services.NewTodoService(*todoRepo)
	todoController := controllers.NewTodoController(*todoService)

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	// Get todo list
	e.GET("/todo", todoController.GetTodoList)

	// Create todo
	e.POST("/todo", todoController.CreateTodo)

	// Update todo
	e.PUT("/todo/:id", todoController.UpdateTodo)

	// Get todo by id
	e.GET("/todo/:id", todoController.GetTodo)

	// e.POST("/login", func(c echo.Context) error {
	// 	req := new(LoginRequest)
	// 	if err := c.Bind(&req); err != nil {
	// 		fmt.Println(err)
	// 		fmt.Printf("%+v\n", req)
	// 		return c.String(http.StatusBadRequest, "Bad Request")
	// 	}
	// 	v := validator.New()
	// 	if err := v.Struct(req); err != nil {
	// 		fmt.Println(err)
	// 		fmt.Printf("%+v\n", req)
	// 		return c.String(http.StatusBadRequest, "Bad Request")
	// 	}
	// 	p := []byte(req.Password)
	// 	sha256 := sha256.Sum256(p)
	// 	fp := fetchPassword()
	// 	if sha256 != fp {
	// 		return c.String(http.StatusUnauthorized, "Unauthorized")
	// 	}

	// 	return c.String(http.StatusOK, "Login!!")
	// })

	// e.POST("/key1", func(c echo.Context) error {
	// 	req := new(Key1Request)
	// 	if err := c.Bind(&req); err != nil {
	// 		fmt.Println(err)
	// 		fmt.Printf("%+v\n", req)
	// 		return c.String(http.StatusBadRequest, "Bad Request")
	// 	}

	// 	var ctx = context.Background()
	// 	println("Request")
	// 	println(req.Value)
	// 	rdb.Set(ctx, "key1", req.Value, 0)
	// 	return c.String(http.StatusOK, "Success")
	// })
	// e.GET("/key1", func(c echo.Context) error {
	// 	var ctx = context.Background()
	// 	ret, err := rdb.Get(ctx, "key1").Result()
	// 	if err != nil {
	// 		println("Error: ", err)
	// 		return c.String(http.StatusInternalServerError, "Internal Server Error")
	// 	}
	// 	println(ret)
	// 	return c.String(http.StatusOK, ret)
	// })
	return e
}

func fetchPassword() [32]byte {
	p := []byte("password")
	sha256 := sha256.Sum256(p)
	return sha256
}
