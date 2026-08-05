package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/masaya-nishimura-09/movie-log-api/internal/config"
	"github.com/masaya-nishimura-09/movie-log-api/internal/handler"
	"github.com/masaya-nishimura-09/movie-log-api/internal/middleware"
	"github.com/masaya-nishimura-09/movie-log-api/internal/repository"
	"github.com/masaya-nishimura-09/movie-log-api/internal/usecase"
	"github.com/ulule/limiter/v3"
	ginlimiter "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println(".env file not found, using environment variables")
	}

	s := os.Getenv("JWT_SECRET")
	if s == "" {
		log.Fatalf("environment variable JWT_SECRET is required")
	}
	secret := []byte(s)
	tokenRepo := repository.NewTokenRepo(secret)

	router := gin.Default()

	db, err := config.NewDB()
	if err != nil {
		log.Fatalf("%v", err)
	}

	rate := limiter.Rate{
		Period: 1 * time.Minute,
		Limit:  5,
	}
	store := memory.NewStore()
	loginLimiter := ginlimiter.NewMiddleware(limiter.New(store, rate))

	userRepo := repository.NewUserRepo(db)
	userUsecase := usecase.NewUserUsecase(userRepo, tokenRepo)
	userHandler := handler.NewUserHandler(userUsecase)

	users := router.Group("/users")
	{
		users.POST("/login", loginLimiter, userHandler.Login)
		users.POST("/register", userHandler.CreateUser)
	}

	authUsers := router.Group("/users")
	authUsers.Use(middleware.JWTAuth(userRepo, tokenRepo))
	{
		authUsers.PUT("/", userHandler.UpdateUser)
		authUsers.PUT("/password", userHandler.UpdatePassword)
		authUsers.DELETE("/", userHandler.DeleteUser)
	}

	router.Run("0.0.0.0:8080")
}
