package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/masaya-nishimura-09/movie-log-api/internal/config"
	authhandler "github.com/masaya-nishimura-09/movie-log-api/internal/handler/auth"
	mediahandler "github.com/masaya-nishimura-09/movie-log-api/internal/handler/media"
	moviehandler "github.com/masaya-nishimura-09/movie-log-api/internal/handler/movie"
	recordhandler "github.com/masaya-nishimura-09/movie-log-api/internal/handler/record"
	"github.com/masaya-nishimura-09/movie-log-api/internal/handler/response"
	schedulerhandler "github.com/masaya-nishimura-09/movie-log-api/internal/handler/scheduler"
	userhandler "github.com/masaya-nishimura-09/movie-log-api/internal/handler/user"
	authinfra "github.com/masaya-nishimura-09/movie-log-api/internal/infrastructure/auth"
	mediainfra "github.com/masaya-nishimura-09/movie-log-api/internal/infrastructure/media"
	movieinfra "github.com/masaya-nishimura-09/movie-log-api/internal/infrastructure/movie"
	recordinfra "github.com/masaya-nishimura-09/movie-log-api/internal/infrastructure/record"
	userinfra "github.com/masaya-nishimura-09/movie-log-api/internal/infrastructure/user"
	"github.com/masaya-nishimura-09/movie-log-api/internal/logging"
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

	slog.SetDefault(logging.New(os.Stdout, config.AppVersion()))

	ginMode, err := config.GinMode()
	if err != nil {
		log.Fatal(err)
	}
	gin.SetMode(ginMode)

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

	trustedProxies, err := config.TrustedProxies()
	if err != nil {
		log.Fatal(err)
	}

	cleanupEnable, err := config.CleanupEnable()
	if err != nil {
		log.Fatal(err)
	}

	cleanupInterval, err := config.CleanupInterval()
	if err != nil {
		log.Fatal(err)
	}

	tmdbClient := movieinfra.NewTMDBClient(tmdbEndpoint, tmdbAccessToken)

	clientKey := middleware.ClientKey(config.InternalAPISecret())
	userKey := func(c *gin.Context) string {
		return fmt.Sprint(c.MustGet("userID"))
	}
	limitReached := func(c *gin.Context) {
		logging.Warn(c, "rate_limited", nil)
		response.TooManyRequests(c)
	}

	loginRate := limiter.Rate{
		Period: 1 * time.Minute,
		Limit:  5,
	}
	loginStore := memory.NewStore()
	loginLimiter := ginlimiter.NewMiddleware(
		limiter.New(loginStore, loginRate),
		ginlimiter.WithKeyGetter(clientKey),
		ginlimiter.WithLimitReachedHandler(limitReached),
	)

	passwordCheckRate := limiter.Rate{
		Period: 1 * time.Minute,
		Limit:  5,
	}
	passwordCheckStore := memory.NewStore()
	passwordCheckLimiter := ginlimiter.NewMiddleware(
		limiter.New(passwordCheckStore, passwordCheckRate),
		ginlimiter.WithKeyGetter(userKey),
		ginlimiter.WithLimitReachedHandler(limitReached),
	)

	registerRate := limiter.Rate{
		Period: 1 * time.Hour,
		Limit:  10,
	}
	registerStore := memory.NewStore()
	registerLimiter := ginlimiter.NewMiddleware(
		limiter.New(registerStore, registerRate),
		ginlimiter.WithKeyGetter(clientKey),
		ginlimiter.WithLimitReachedHandler(limitReached),
	)

	refreshRate := limiter.Rate{
		Period: 1 * time.Minute,
		Limit:  30,
	}
	refreshStore := memory.NewStore()
	refreshLimiter := ginlimiter.NewMiddleware(
		limiter.New(refreshStore, refreshRate),
		ginlimiter.WithKeyGetter(clientKey),
		ginlimiter.WithLimitReachedHandler(limitReached),
	)

	mediaRate := limiter.Rate{
		Period: 1 * time.Hour,
		Limit:  30,
	}
	mediaStore := memory.NewStore()
	mediaLimiter := ginlimiter.NewMiddleware(
		limiter.New(mediaStore, mediaRate),
		ginlimiter.WithKeyGetter(userKey),
		ginlimiter.WithLimitReachedHandler(limitReached),
	)

	movieRate := limiter.Rate{
		Period: 1 * time.Minute,
		Limit:  60,
	}
	movieStore := memory.NewStore()
	movieLimiter := ginlimiter.NewMiddleware(
		limiter.New(movieStore, movieRate),
		ginlimiter.WithKeyGetter(userKey),
		ginlimiter.WithLimitReachedHandler(limitReached),
	)

	bodyLimit := middleware.BodyLimit(1024 * 1024)

	// infrastructure
	accessTokenService := authinfra.NewAccessTokenService(
		secret,
		accessTokenTTL,
	)
	refreshTokenRepo := authinfra.NewRefreshTokenRepository(db, refreshTokenTTL)
	userRepo := userinfra.NewRepository(db)
	recordRepo := recordinfra.NewRepository(db)
	articleRepo := recordinfra.NewArticleRepository(db)
	movieService := movieinfra.NewService(
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
	userUsecase := userusecase.NewUserUsecase(
		userRepo,
		refreshTokenRepo,
		mediaService,
	)
	recordUsecase := recordusecase.NewRecordUsecase(recordRepo, mediaService)
	articleUsecase := recordusecase.NewArticleUsecase(articleRepo, recordRepo)
	movieUsecase := movieusecase.NewMovieUsecase(movieService)
	mediaUsecase := mediausecase.NewMediaUsecase(mediaService)

	// handler
	authHandler := authhandler.NewAuthHandler(authUsecase)
	userHandler := userhandler.NewUserHandler(userUsecase)
	recordHandler := recordhandler.NewRecordHandler(recordUsecase)
	articleHandler := recordhandler.NewArticleHandler(articleUsecase)
	movieHandler := moviehandler.NewMovieHandler(movieUsecase)
	mediaHandler := mediahandler.NewMediaHandler(mediaUsecase)
	scheduler := schedulerhandler.NewScheduler(authUsecase, cleanupInterval)

	// routing
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.CustomRecoveryWithWriter(nil, func(c *gin.Context, err any) {
		logging.Error(c, "panic_recovered", fmt.Errorf("%v", err), "stack", string(debug.Stack()))
		response.InternalServerError(c)
		c.Abort()
	}))
	router.NoRoute(response.NotFound)
	if err := router.SetTrustedProxies(trustedProxies); err != nil {
		log.Fatal(err)
	}
	router.RemoteIPHeaders = []string{"X-Forwarded-For"}

	auth := router.Group("/auth")
	auth.Use(bodyLimit)
	{
		auth.POST("/login", loginLimiter, authHandler.Login)
		auth.POST("/logout", authHandler.Logout)
		auth.POST("/refresh", refreshLimiter, authHandler.Refresh)
	}

	users := router.Group("/users")
	users.Use(bodyLimit)
	{
		users.POST("/register", registerLimiter, userHandler.Create)
	}

	authUsers := router.Group("/users")
	authUsers.Use(bodyLimit)
	authUsers.Use(middleware.JWTAuth(authUsecase, userUsecase))
	{
		authUsers.GET("/", userHandler.GetByID)
		authUsers.PUT("/username", userHandler.UpdateUsername)
		authUsers.PUT("/email", passwordCheckLimiter, userHandler.UpdateEmail)
		authUsers.PUT("/password", passwordCheckLimiter, userHandler.UpdatePassword)
		authUsers.DELETE("/", userHandler.Delete)
	}

	records := router.Group("/records")
	records.Use(bodyLimit)
	records.Use(middleware.JWTAuth(authUsecase, userUsecase))
	{
		records.POST("/", recordHandler.Create)
		records.GET("/", recordHandler.ListByUserID)
		records.GET("/:id", recordHandler.GetByID)
		records.PUT("/:id", recordHandler.Update)
		records.DELETE("/:id", recordHandler.Delete)
	}

	articles := router.Group("/articles")
	articles.Use(bodyLimit)
	articles.Use(middleware.JWTAuth(authUsecase, userUsecase))
	{
		articles.POST("/", articleHandler.Create)
		articles.GET("/", articleHandler.ListByUserID)
		articles.GET("/:id", articleHandler.GetByID)
		articles.PUT("/:id", articleHandler.Update)
		articles.DELETE("/:id", articleHandler.Delete)
	}

	movies := router.Group("/movies")
	movies.Use(middleware.JWTAuth(authUsecase, userUsecase))
	{
		movies.GET("/:id", movieLimiter, movieHandler.GetByID)
		movies.GET("/search", movieLimiter, movieHandler.SearchByTitle)
	}

	media := router.Group("/media")
	media.Use(middleware.JWTAuth(authUsecase, userUsecase))
	{
		media.POST("/", mediaLimiter, mediaHandler.Upload)
	}

	if cleanupEnable {
		ctx := context.Background()
		go scheduler.Run(ctx)
	}

	server := &http.Server{
		Addr:              "0.0.0.0:8080",
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
