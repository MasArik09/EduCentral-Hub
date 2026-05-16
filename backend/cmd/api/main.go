package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"backend/config"
	"backend/internal/auth"
	"backend/internal/handler"
	"backend/internal/middleware"
	"backend/internal/repository"
	"backend/internal/usecase"

	"github.com/gin-gonic/gin"
)

func main() {
	db, err := config.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	if err := ensureUploadDirs(); err != nil {
		log.Fatalf("upload directory init failed: %v", err)
	}

	router := gin.Default()
	router.Use(middleware.CORSMiddleware())
	router.Static("/uploads", "./uploads")

	userRepo := repository.NewUserRepository(db)
	tokenService := auth.NewTokenService(os.Getenv("JWT_SECRET"), auth.NewInMemoryTokenStore())
	authUsecase := usecase.NewAuthUsecase(userRepo, tokenService)
	authHandler := handler.NewAuthHandler(authUsecase)
	userUsecase := usecase.NewUserUsecase(userRepo)
	userHandler := handler.NewUserHandler(userUsecase)
	courseRepo := repository.NewCourseRepository(db)
	courseUsecase := usecase.NewCourseUsecase(courseRepo)
	courseHandler := handler.NewCourseHandler(courseUsecase)
	lessonRepo := repository.NewLessonRepository(db)
	lessonUsecase := usecase.NewLessonUsecase(lessonRepo)
	lessonHandler := handler.NewLessonHandler(lessonUsecase)
	attendanceRepo := repository.NewAttendanceRepository(db)
	attendanceUsecase := usecase.NewAttendanceUsecase(attendanceRepo)
	attendanceHandler := handler.NewAttendanceHandler(attendanceUsecase)
	assignmentRepo := repository.NewAssignmentRepository(db)
	assignmentUsecase := usecase.NewAssignmentUsecase(assignmentRepo)
	assignmentHandler := handler.NewAssignmentHandler(assignmentUsecase)

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
	authGroup.POST("/refresh", authHandler.Refresh)
	authGroup.POST("/logout", middleware.JWTAuthMiddleware(tokenService), authHandler.Logout)

	userGroup := router.Group("/api/users")
	userGroup.POST("/profile-picture", middleware.JWTAuthMiddleware(tokenService), userHandler.UploadProfilePicture)

	courseGroup := router.Group("/api/courses")
	courseGroup.GET("/:id", courseHandler.GetCourseDetails)
	courseGroup.POST("", middleware.JWTAuthMiddleware(tokenService), courseHandler.CreateCourse)
	courseGroup.POST("/:id/enroll", middleware.JWTAuthMiddleware(tokenService), courseHandler.EnrollCourse)
	courseGroup.POST("/:id/lessons", middleware.JWTAuthMiddleware(tokenService), lessonHandler.CreateLesson)
	courseGroup.GET("/:id/lessons", middleware.JWTAuthMiddleware(tokenService), lessonHandler.ListLessonsByCourse)
	courseGroup.POST(":id/assignments", middleware.JWTAuthMiddleware(tokenService), assignmentHandler.CreateAssignment)
	courseGroup.POST("/:id/attendance", middleware.JWTAuthMiddleware(tokenService), attendanceHandler.MarkAttendance)
	courseGroup.GET("/:id/my-attendance", middleware.JWTAuthMiddleware(tokenService), attendanceHandler.ListMyAttendance)

	lessonGroup := router.Group("/api/lessons")
	lessonGroup.GET("/:lesson_id", middleware.JWTAuthMiddleware(tokenService), lessonHandler.GetLessonDetail)

	assignmentGroup := router.Group("/api/assignments")
	assignmentGroup.POST("/:id/submit", middleware.JWTAuthMiddleware(tokenService), assignmentHandler.SubmitAssignment)

	if err := router.Run(); err != nil {
		log.Fatalf("server start failed: %v", err)
	}
}

func ensureUploadDirs() error {
	profilesPath := filepath.Join("uploads", "profiles")
	assignmentsPath := filepath.Join("uploads", "assignments")
	if err := os.MkdirAll(profilesPath, 0o755); err != nil {
		return err
	}
	return os.MkdirAll(assignmentsPath, 0o755)
}
