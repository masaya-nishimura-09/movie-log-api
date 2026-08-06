package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/masaya-nishimura-09/movie-log-api/internal/config"
	"github.com/masaya-nishimura-09/movie-log-api/internal/middleware"
	authhandler "github.com/masaya-nishimura-09/movie-log-api/internal/handler/auth"
	userhandler "github.com/masaya-nishimura-09/movie-log-api/internal/handler/user"
	authusecase "github.com/masaya-nishimura-09/movie-log-api/internal/usecase/auth"
	userusecase "github.com/masaya-nishimura-09/movie-log-api/internal/usecase/user"
	authrepository "github.com/masaya-nishimura-09/movie-log-api/internal/repository/auth"
	userrepository "github.com/masaya-nishimura-09/movie-log-api/internal/repository/user"
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

	// repository
	refreshTokenRepo := authrepository.NewRefreshTokenRepo(secret)
	userRepo := userrepository.NewUserRepo(db)

	// usecase
	authUsecase := authusecase.NewAuthUsecase(userRepo, refreshTokenRepo, secret)
	userUsecase := userusecase.NewUserUsecase(userRepo)

	// handler
	authHandler := authhandler.NewAuthHandler(authUsecase)
	userHandler := userhandler.NewUserHandler(userUsecase)

	auth := router.Group("/auth")
	{
		auth.POST("/login", loginLimiter, authHandler.Login)
	}

	users := router.Group("/users")
	{
		users.POST("/register", userHandler.CreateUser)
	}

	authUsers := router.Group("/users")
	authUsers.Use(middleware.JWTAuth(authUsecase, userUsecase))
	{
		authUsers.PUT("/", userHandler.UpdateUser)
		authUsers.DELETE("/", userHandler.DeleteUser)
	}

	router.Run("0.0.0.0:8080")
}
