package main

import (
	"crypto/sha256"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/mllb/sampletodo/api"
	"github.com/mllb/sampletodo/models"
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

func main() {
	loadEnv()
	// rdb := redis.NewClient(&redis.Options{
	// 	Addr:     "localhost:16379",
	// 	Password: "",
	// 	DB:       0,
	// 	PoolSize: 100,
	// })
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

	e := api.NewRouter(db)
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
