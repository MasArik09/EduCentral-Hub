package main

import (
	"log"
	"net/http"

	"backend/config"
	"backend/internal/handler"
	"backend/internal/middleware"
	"backend/internal/repository"
	"backend/internal/usecase"

	"github.com/gin-gonic/gin"
)

func main() {
	db, err := config.InitDB()
	if err != nil {
		log.Fatalf("database init failed: %v", err)
	}

	router := gin.Default()

	userRepo := repository.NewUserRepository(db)
	authUsecase := usecase.NewAuthUsecase(userRepo)
	authHandler := handler.NewAuthHandler(authUsecase)
	courseRepo := repository.NewCourseRepository(db)
	courseUsecase := usecase.NewCourseUsecase(courseRepo)
	courseHandler := handler.NewCourseHandler(courseUsecase)

	router.GET("/health", func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"database": "down", "error": "db handle unavailable"})
			return
		}

		if err := sqlDB.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"database": "down"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"database": "up"})
	})

	authGroup := router.Group("/api/auth")
	authGroup.POST("/register", authHandler.Register)
	authGroup.POST("/login", authHandler.Login)

	courseGroup := router.Group("/api/courses")
	courseGroup.GET("/:id", courseHandler.GetCourseDetails)
	courseGroup.POST("", middleware.JWTAuthMiddleware(), courseHandler.CreateCourse)
	courseGroup.POST("/:id/enroll", middleware.JWTAuthMiddleware(), courseHandler.EnrollCourse)

	if err := router.Run(); err != nil {
		log.Fatalf("server start failed: %v", err)
	}
}
