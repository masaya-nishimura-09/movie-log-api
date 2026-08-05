package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/model"
	"github.com/masaya-nishimura-09/movie-log-api/internal/usecase"
)

type LoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateReq struct {
	Username string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateReq struct {
	Username string `json:"name"`
	Email    string `json:"email"`
}

type UpdatePasswordReq struct {
	Password string `json:"password"`
}

type UserHandler struct {
	userUsecase *usecase.UserUsecase
}

func NewUserHandler(usecase *usecase.UserUsecase) *UserHandler {
	return &UserHandler{userUsecase: usecase}
}

func getAuthUserID(c *gin.Context) (model.UserID, bool) {
	v, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "UNAUTHENTICATED",
			"message": "authentication required",
		})
		return 0, false
	}
	id, ok := v.(model.UserID)
	if !ok {
		log.Println("userID in context is not of type uint")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_SERVER_ERROR",
			"message": "internal server error",
		})
		return 0, false
	}
	return model.UserID(id), true
}

func (uh *UserHandler) Login(c *gin.Context) {
	ctx := c.Request.Context()
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_INPUT",
			"message": err.Error(),
		})
		return
	}

	email := model.Email(req.Email)
	password := model.Password(req.Password)

	token, err := uh.userUsecase.Login(ctx, email, password)
	if errors.Is(err, model.ErrInvalidCredentials) {
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
		"token": string(token),
	})
}

func (uh *UserHandler) CreateUser(c *gin.Context) {
	ctx := c.Request.Context()

	var req CreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_INPUT",
			"message": err.Error(),
		})
		return
	}

	username := model.Username(req.Username)
	email := model.Email(req.Email)
	password := model.Password(req.Password)

	token, err := uh.userUsecase.Register(ctx, username, email, password)
	if errors.Is(err, model.ErrValidation) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_INPUT",
			"message": err.Error(),
		})
		return
	}

	if errors.Is(err, model.ErrUserAlreadyExists) {
		c.JSON(http.StatusConflict, gin.H{
			"code":    "USER_ALREADY_EXISTS",
			"message": "user with this email already exists",
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

	c.JSON(http.StatusCreated, gin.H{
		"token": string(token),
	})
}

func (uh *UserHandler) UpdateUser(c *gin.Context) {
	ctx := c.Request.Context()

	authUserID, ok := getAuthUserID(c)
	if !ok {
		return
	}

	var req UpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_INPUT",
			"message": err.Error(),
		})
		return
	}

	username := model.Username(req.Username)
	email := model.Email(req.Email)

	err := uh.userUsecase.UpdateUser(ctx, authUserID, username, email)
	if errors.Is(err, model.ErrValidation) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_INPUT",
			"message": err.Error(),
		})
		return
	}
	if errors.Is(err, model.ErrUserNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "USER_NOT_FOUND",
			"message": "user not found",
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
	c.JSON(http.StatusOK, gin.H{})
}

func (uh *UserHandler) UpdatePassword(c *gin.Context) {
	ctx := c.Request.Context()

	authUserID, ok := getAuthUserID(c)
	if !ok {
		return
	}

	var req UpdatePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_INPUT",
			"message": err.Error(),
		})
		return
	}

	password := model.Password(req.Password)

	err := uh.userUsecase.UpdatePassword(ctx, authUserID, password)
	if errors.Is(err, model.ErrValidation) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_INPUT",
			"message": err.Error(),
		})
		return
	}
	if errors.Is(err, model.ErrUserNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "USER_NOT_FOUND",
			"message": "user not found",
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
	c.JSON(http.StatusOK, gin.H{})
}

func (uh *UserHandler) DeleteUser(c *gin.Context) {
	ctx := c.Request.Context()

	authUserID, ok := getAuthUserID(c)
	if !ok {
		return
	}

	err := uh.userUsecase.DeleteUser(ctx, authUserID)
	if errors.Is(err, model.ErrUserNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "USER_NOT_FOUND",
			"message": "user not found",
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
	c.JSON(http.StatusOK, gin.H{})
}
