package config

import (
	"fmt"
	"log"
	"os"

	"backend/internal/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() (*gorm.DB, error) {
	_ = godotenv.Load()

	host := getEnv("DB_HOST", "127.0.0.1")
	port := getEnv("DB_PORT", "3306")
	user := getEnv("DB_USER", "root")
	pass := getEnv("DB_PASS", "")
	name := "educentral_hub"

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user,
		pass,
		host,
		port,
		name,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Printf("database connection failed: %v", err)
		return nil, err
	}

	if err := db.AutoMigrate(
		&models.Role{},
		&models.User{},
		&models.Profile{},
		&models.Course{},
		&models.Enrollment{},
		&models.Lesson{},
		&models.Assignment{},
		&models.Submission{},
		&models.Attendance{},
	); err != nil {
		log.Printf("database migration failed: %v", err)
		return nil, err
	}

	DB = db
	return db, nil
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value != "" {
		return value
	}

	return fallback
}
