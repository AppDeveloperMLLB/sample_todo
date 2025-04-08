package api

import (
	"crypto/sha256"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mllb/sampletodo/controllers"
	appjwt "github.com/mllb/sampletodo/jwt"
	"github.com/mllb/sampletodo/repositories"
	"github.com/mllb/sampletodo/services"
	"gorm.io/gorm"
)

func NewRouter(db *gorm.DB) *echo.Echo {
	todoRepo := repositories.NewTodoRepository(db)
	todoService := services.NewTodoService(todoRepo)
	todoController := controllers.NewTodoController(todoService)
	authRepo := repositories.NewAuthRepository(db)
	authService := services.NewAuthService(authRepo)
	authController := controllers.NewAuthController(authService)

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
	// e.Static("/assets", "public/assets")

	// e.File("/", "public/index.html") // GET /
	// e.File("/signup", "public/signup.html") // GET /signup
	e.POST("/signup", authController.SignUp) // POST /signup
	// e.File("/login", "public/login.html") // GET /login
	e.POST("/signin", authController.SignIn)   // POST /login
	e.POST("/signout", authController.SignOut) // POST /signout
	// e.File("/todos", "public/todos.html") // GET /todos
	config := echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(appjwt.JwtCustomClaims)
		},
		SigningKey: []byte(appjwt.SigningKey),
		ErrorHandler: func(c echo.Context, err error) error {
			msg := "トークンが無効なのだ"
			switch err {
			case echojwt.ErrJWTMissing:
				msg = "トークンがないのだ"
			case echojwt.ErrJWTInvalid:
				msg = "トークンが無効なのだ"
			}

			return c.JSON(http.StatusUnauthorized, msg)
		},
	}
	api := e.Group("/api")
	api.Use(echojwt.WithConfig(config)) // /api 下はJWTの認証が必要
	// Get todo list
	api.GET("/todo", todoController.GetTodoList)
	// api.GET("/todos", handler.GetTodos) // GET /api/todos
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
