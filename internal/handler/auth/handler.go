package auth

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/auth"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
	authusecase "github.com/masaya-nishimura-09/movie-log-api/internal/usecase/auth"
)

type LoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshReq struct {
	RefreshToken string `json:"refresh_token"`
}

type AuthHandler struct {
	authUsecase *authusecase.AuthUsecase
}

func NewAuthHandler(usecase *authusecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{authUsecase: usecase}
}

func (ah *AuthHandler) Login(c *gin.Context) {
	ctx := c.Request.Context()
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_INPUT",
			"message": err.Error(),
		})
		return
	}

	email, err := user.NewEmail(req.Email)
	password, err := user.NewPassword(req.Password)

	if errors.Is(err, exception.ErrValidation) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_INPUT",
			"message": err.Error(),
		})
		return
	}

	accessToken, refreshToken, err := ah.authUsecase.Login(ctx, email, password)
	if errors.Is(err, exception.ErrInvalidCredentials) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "INVALID_CREDENTIALS",
			"message": "invalid email or password",
		})
		return
	}
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_SERVER_ERROR",
			"message": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  string(accessToken.Value),
		"refresh_token": string(refreshToken.Value),
	})
}

func (ah *AuthHandler) Refresh(c *gin.Context) {
	ctx := c.Request.Context()
	var req RefreshReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_INPUT",
			"message": err.Error(),
		})
		return
	}

	accessToken, refreshToken, err := ah.authUsecase.Refresh(
		ctx,
		auth.RefreshTokenValue(req.RefreshToken),
	)
	if errors.Is(err, exception.ErrInvalidToken) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "INVALID_TOKEN",
			"message": "invalid or expired refresh token",
		})
		return
	}
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_SERVER_ERROR",
			"message": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  string(accessToken.Value),
		"refresh_token": string(refreshToken.Value),
	})
}
