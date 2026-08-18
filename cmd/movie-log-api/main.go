package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/masaya-nishimura-09/movie-log-api/internal/config"
	authhandler "github.com/masaya-nishimura-09/movie-log-api/internal/handler/auth"
	recordhandler "github.com/masaya-nishimura-09/movie-log-api/internal/handler/record"
	userhandler "github.com/masaya-nishimura-09/movie-log-api/internal/handler/user"
	authinfra "github.com/masaya-nishimura-09/movie-log-api/internal/infrastructure/auth"
	recordinfra "github.com/masaya-nishimura-09/movie-log-api/internal/infrastructure/record"
	userinfra "github.com/masaya-nishimura-09/movie-log-api/internal/infrastructure/user"
	"github.com/masaya-nishimura-09/movie-log-api/internal/middleware"
	authusecase "github.com/masaya-nishimura-09/movie-log-api/internal/usecase/auth"
	recordusecase "github.com/masaya-nishimura-09/movie-log-api/internal/usecase/record"
	userusecase "github.com/masaya-nishimura-09/movie-log-api/internal/usecase/user"
	"github.com/ulule/limiter/v3"
	ginlimiter "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println(".env file not found, using environment variables")
	}

	db, err := config.NewDB()
	if err != nil {
		log.Fatalf("%v", err)
	}

	secret, err := config.Secret()
	if err != nil {
		log.Fatal(err)
	}

	accessTokenTTL, err := config.AccessTokenTTL()
	if err != nil {
		log.Fatal(err)
	}

	refreshTokenTTL, err := config.RefreshTokenTTL()
	if err != nil {
		log.Fatal(err)
	}

	rate := limiter.Rate{
		Period: 1 * time.Minute,
		Limit:  5,
	}
	store := memory.NewStore()
	loginLimiter := ginlimiter.NewMiddleware(limiter.New(store, rate))

	// infrastructure
	accessTokenService := authinfra.NewAccessTokenService(
		secret,
		accessTokenTTL,
	)
	refreshTokenRepo := authinfra.NewRefreshTokenRepo(db, refreshTokenTTL)
	userRepo := userinfra.NewUserRepo(db)
	recordRepo := recordinfra.NewRecordRepo(db)

	// usecase
	authUsecase := authusecase.NewAuthUsecase(
		userRepo,
		accessTokenService,
		refreshTokenRepo,
	)
	userUsecase := userusecase.NewUserUsecase(userRepo, refreshTokenRepo)
	recordUsecase := recordusecase.NewRecordUsecase(recordRepo)

	// handler
	authHandler := authhandler.NewAuthHandler(authUsecase)
	userHandler := userhandler.NewUserHandler(userUsecase)
	recordHandler := recordhandler.NewRecordHandler(recordUsecase)

	// routing
	router := gin.Default()

	auth := router.Group("/auth")
	{
		auth.POST("/login", loginLimiter, authHandler.Login)
		auth.POST("/logout", authHandler.Logout)
		auth.POST("/refresh", loginLimiter, authHandler.Refresh)
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

	records := router.Group("/records")
	records.Use(middleware.JWTAuth(authUsecase, userUsecase))
	{
		records.POST("/", recordHandler.CreateRecord)
		records.GET("/", recordHandler.ListRecords)
		records.GET("/:id", recordHandler.GetRecord)
		records.PUT("/:id", recordHandler.UpdateRecord)
		records.DELETE("/:id", recordHandler.DeleteRecord)
	}

	router.Run("0.0.0.0:8080")
}
