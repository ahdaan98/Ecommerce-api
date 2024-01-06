package initializer

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectToDb() {
	var err error

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",os.Getenv("DB_HOST"),os.Getenv("DB_USERNAME"),os.Getenv("DB_PASSWORD"),os.Getenv("DB_NAME"),os.Getenv("DB_PORT"),os.Getenv("SSLMODE"))

	DB,err=gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err!=nil {
		panic("Failed to connect Database")
	}
}