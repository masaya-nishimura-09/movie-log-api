package auth

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/auth"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
	"github.com/masaya-nishimura-09/movie-log-api/internal/handler/response"
	authusecase "github.com/masaya-nishimura-09/movie-log-api/internal/usecase/auth"
)

type LoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r LoginReq) toDomain() (
	user.Email,
	user.Password,
	error,
) {
	email, err := user.NewEmail(r.Email)
	if err != nil {
		return "", "", err
	}
	password, err := user.NewPassword(r.Password)
	if err != nil {
		return "", "", err
	}
	return email, password, nil
}

type LogoutReq struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshReq struct {
	RefreshToken string `json:"refresh_token"`
}

type AuthHandler struct {
	authUsecase authusecase.Usecase
}

func NewAuthHandler(usecase authusecase.Usecase) *AuthHandler {
	return &AuthHandler{authUsecase: usecase}
}

func (ah *AuthHandler) Login(c *gin.Context) {
	ctx := c.Request.Context()
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidInput(c, err)
		return
	}

	email, password, err := req.toDomain()
	if errors.Is(err, exception.ErrInvalid) {
		response.InvalidInput(c, err)
		return
	}
	if err != nil {
		log.Println(err)
		response.InternalServerError(c)
		return
	}

	accessToken, refreshToken, err := ah.authUsecase.Login(ctx, email, password)
	if errors.Is(err, exception.ErrInvalid) {
		response.InvalidCredentials(c)
		return
	}
	if err != nil {
		log.Println(err)
		response.InternalServerError(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  string(accessToken.Value),
		"refresh_token": string(refreshToken.Value),
	})
}

func (ah *AuthHandler) Logout(c *gin.Context) {
	ctx := c.Request.Context()
	var req LogoutReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidInput(c, err)
		return
	}

	err := ah.authUsecase.Logout(ctx, auth.RefreshTokenValue(req.RefreshToken))
	if errors.Is(err, exception.ErrInvalid) {
		c.Status(http.StatusNoContent)
		return
	}
	if err != nil {
		log.Println(err)
		response.InternalServerError(c)
		return
	}

	c.Status(http.StatusNoContent)
}

func (ah *AuthHandler) Refresh(c *gin.Context) {
	ctx := c.Request.Context()
	var req RefreshReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidInput(c, err)
		return
	}

	accessToken, refreshToken, err := ah.authUsecase.Refresh(
		ctx,
		auth.RefreshTokenValue(req.RefreshToken),
	)
	if errors.Is(err, exception.ErrInvalid) {
		response.InvalidRefreshToken(c)
		return
	}
	if err != nil {
		log.Println(err)
		response.InternalServerError(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  string(accessToken.Value),
		"refresh_token": string(refreshToken.Value),
	})
}
