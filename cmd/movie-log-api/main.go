package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/masaya-nishimura-09/movie-log-api/internal/config"
	authhandler "github.com/masaya-nishimura-09/movie-log-api/internal/handler/auth"
	mediahandler "github.com/masaya-nishimura-09/movie-log-api/internal/handler/media"
	moviehandler "github.com/masaya-nishimura-09/movie-log-api/internal/handler/movie"
	recordhandler "github.com/masaya-nishimura-09/movie-log-api/internal/handler/record"
	userhandler "github.com/masaya-nishimura-09/movie-log-api/internal/handler/user"
	authinfra "github.com/masaya-nishimura-09/movie-log-api/internal/infrastructure/auth"
	mediainfra "github.com/masaya-nishimura-09/movie-log-api/internal/infrastructure/media"
	movieinfra "github.com/masaya-nishimura-09/movie-log-api/internal/infrastructure/movie"
	recordinfra "github.com/masaya-nishimura-09/movie-log-api/internal/infrastructure/record"
	userinfra "github.com/masaya-nishimura-09/movie-log-api/internal/infrastructure/user"
	"github.com/masaya-nishimura-09/movie-log-api/internal/middleware"
	authusecase "github.com/masaya-nishimura-09/movie-log-api/internal/usecase/auth"
	mediausecase "github.com/masaya-nishimura-09/movie-log-api/internal/usecase/media"
	movieusecase "github.com/masaya-nishimura-09/movie-log-api/internal/usecase/movie"
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

	s3Client, err := config.NewS3Client()
	if err != nil {
		log.Fatalf("%v", err)
	}

	s3Bucket, err := config.S3Bucket()
	if err != nil {
		log.Fatal(err)
	}

	s3PublicBaseURL, err := config.S3PublicBaseURL()
	if err != nil {
		log.Fatal(err)
	}

	tmdbAccessToken, err := config.TMDBAccessToken()
	if err != nil {
		log.Fatal(err)
	}

	tmdbEndpoint, err := config.TMDBEndpoint()
	if err != nil {
		log.Fatal(err)
	}

	tmdbPosterBaseURL, err := config.TMDBPosterBaseURL()
	if err != nil {
		log.Fatal(err)
	}

	tmdbClient := movieinfra.NewTMDBClient(tmdbEndpoint, tmdbAccessToken)

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
	movieService := movieinfra.NewMovieService(
		tmdbClient,
		tmdbPosterBaseURL,
	)
	mediaService := mediainfra.NewService(
		s3Client,
		s3Bucket,
		s3PublicBaseURL,
	)

	// usecase
	authUsecase := authusecase.NewAuthUsecase(
		userRepo,
		accessTokenService,
		refreshTokenRepo,
	)
	userUsecase := userusecase.NewUserUsecase(userRepo, refreshTokenRepo)
	recordUsecase := recordusecase.NewRecordUsecase(recordRepo, mediaService)
	movieUsecase := movieusecase.NewMovieUsecase(movieService)
	mediaUsecase := mediausecase.NewMediaUsecase(mediaService)

	// handler
	authHandler := authhandler.NewAuthHandler(authUsecase)
	userHandler := userhandler.NewUserHandler(userUsecase)
	recordHandler := recordhandler.NewRecordHandler(recordUsecase)
	movieHandler := moviehandler.NewMovieHandler(movieUsecase)
	mediaHandler := mediahandler.NewMediaHandler(mediaUsecase)

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
		users.POST("/register", userHandler.Create)
	}

	authUsers := router.Group("/users")
	authUsers.Use(middleware.JWTAuth(authUsecase, userUsecase))
	{
		authUsers.PUT("/", userHandler.Update)
		authUsers.DELETE("/", userHandler.Delete)
	}

	records := router.Group("/records")
	records.Use(middleware.JWTAuth(authUsecase, userUsecase))
	{
		records.POST("/", recordHandler.Create)
		records.GET("/", recordHandler.ListByUserID)
		records.GET("/:id", recordHandler.GetByID)
		records.PUT("/:id", recordHandler.Update)
		records.DELETE("/:id", recordHandler.Delete)
	}

	movies := router.Group("/movies")
	movies.Use(middleware.JWTAuth(authUsecase, userUsecase))
	{
		movies.GET("/:id", movieHandler.GetByID)
		movies.GET("/search", movieHandler.SearchByTitle)
	}

	media := router.Group("/media")
	media.Use(middleware.JWTAuth(authUsecase, userUsecase))
	{
		media.POST("/", mediaHandler.Upload)
	}

	router.Run("0.0.0.0:8080")
}
