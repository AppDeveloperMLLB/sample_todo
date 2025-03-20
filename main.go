package main

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/mllb/sampletodo/models"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Key1Request struct {
	Value string `json:"value"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type CreateTodoRequest struct {
	Title string `json:"title" validate:"required"`
	Body  string `json:"body" validate:"required"`
}

type GetTodoResponse struct {
	Todos []models.Todo `json:"result"`
}

func main() {
	loadEnv()
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:16379",
		Password: "",
		DB:       0,
		PoolSize: 100,
	})
	host := os.Getenv("POSTGRES_HOST")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_NAME")
	port := os.Getenv("POSTGRES_PORT")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Tokyo",
		host, user, password, dbName, port)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return
	}
	if err := migrateDatabase(db); err != nil {
		return
	}

	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	// Get todo list
	e.GET("/todo", func(c echo.Context) error {
		var todoList []models.Todo
		db.Find(&todoList)
		res := GetTodoResponse{Todos: todoList}
		return c.JSON(http.StatusOK, &res)
	})

	// Create todo
	e.POST("/todo", func(c echo.Context) error {
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
		new := models.Todo{Title: req.Title, Body: req.Body}
		result := db.Create(&new)
		if result.Error != nil {
			fmt.Println(err)
			fmt.Printf("%+v\n", req)
			return c.String(http.StatusBadRequest, "Bad Request")
		}
		return c.String(http.StatusOK, "Success")
	})
	// Update todo
	e.PUT("/todo/:id", func(c echo.Context) error {
		id := c.Param("id")
		if id == "" {
			return c.String(http.StatusBadRequest, "Bad Request")
		}

		var todo models.Todo
		result := db.Where("id = ?", id).Take(&todo)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return c.String(http.StatusNotFound, "Not Found")
			}
			fmt.Println(err)
			return c.String(http.StatusBadRequest, "AAA")
		}

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

		result = db.Save(models.Todo{ID: todo.ID, Title: req.Title, Body: req.Body, CreatedAt: todo.CreatedAt})
		if result.Error != nil {
			fmt.Println(err)
			return c.String(http.StatusInternalServerError, "Server Error")
		}

		return c.String(http.StatusOK, "Success")
	})

	// Get todo by id
	e.GET("/todo/:id", func(c echo.Context) error {
		id := c.Param("id")
		if id == "" {
			return c.String(http.StatusBadRequest, "Bad Request")
		}

		var todo models.Todo
		result := db.Where("id = ?", id).Take(&todo)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return c.String(http.StatusNotFound, "Not Found")
			}
			fmt.Println(err)
			return c.String(http.StatusBadRequest, "AAA")
		}

		return c.JSON(http.StatusOK, &todo)
	})

	e.POST("/login", func(c echo.Context) error {
		req := new(LoginRequest)
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
		p := []byte(req.Password)
		sha256 := sha256.Sum256(p)
		fp := fetchPassword()
		if sha256 != fp {
			return c.String(http.StatusUnauthorized, "Unauthorized")
		}

		return c.String(http.StatusOK, "Login!!")
	})

	e.POST("/key1", func(c echo.Context) error {
		req := new(Key1Request)
		if err := c.Bind(&req); err != nil {
			fmt.Println(err)
			fmt.Printf("%+v\n", req)
			return c.String(http.StatusBadRequest, "Bad Request")
		}

		var ctx = context.Background()
		println("Request")
		println(req.Value)
		rdb.Set(ctx, "key1", req.Value, 0)
		return c.String(http.StatusOK, "Success")
	})
	e.GET("/key1", func(c echo.Context) error {
		var ctx = context.Background()
		ret, err := rdb.Get(ctx, "key1").Result()
		if err != nil {
			println("Error: ", err)
			return c.String(http.StatusInternalServerError, "Internal Server Error")
		}
		println(ret)
		return c.String(http.StatusOK, ret)
	})
	e.Logger.Fatal(e.Start(":1323"))
}

func fetchPassword() [32]byte {
	p := []byte("password")
	sha256 := sha256.Sum256(p)
	return sha256
}

func migrateDatabase(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Todo{},
	)
}

func loadEnv() {
	// ここで.envファイル全体を読み込みます。
	// この読み込み処理がないと、個々の環境変数が取得出来ません。
	// 読み込めなかったら err にエラーが入ります。
	err := godotenv.Load(".env")

	// もし err がnilではないなら、"読み込み出来ませんでした"が出力されます。
	if err != nil {
		fmt.Printf("読み込み出来ませんでした: %v", err)
	}
}
