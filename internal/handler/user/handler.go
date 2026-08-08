package user

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	userdomain "github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
	userusecase "github.com/masaya-nishimura-09/movie-log-api/internal/usecase/user"
)

type CreateReq struct {
	Username string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateReq struct {
	Username string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserHandler struct {
	userUsecase *userusecase.UserUsecase
}

func NewUserHandler(usecase *userusecase.UserUsecase) *UserHandler {
	return &UserHandler{userUsecase: usecase}
}

func getUserID(c *gin.Context) (userdomain.ID, bool) {
	v, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "UNAUTHENTICATED",
			"message": "authentication required",
		})
		return 0, false
	}
	id, ok := v.(userdomain.ID)
	if !ok {
		log.Println("userID in context is not of type uint")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_SERVER_ERROR",
			"message": "internal server error",
		})
		return 0, false
	}
	return id, true
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

	username, err := userdomain.NewUsername(req.Username)
	if errors.Is(err, exception.ErrInvalid) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_INPUT",
			"message": err.Error(),
		})
		return
	}

	email, err := userdomain.NewEmail(req.Email)
	if errors.Is(err, exception.ErrInvalid) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_INPUT",
			"message": err.Error(),
		})
		return
	}

	password, err := userdomain.NewPassword(req.Password)
	if errors.Is(err, exception.ErrInvalid) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_INPUT",
			"message": err.Error(),
		})
		return
	}

	createdUser, err := uh.userUsecase.Register(ctx, username, email, password)

	if errors.Is(err, exception.ErrAlreadyExists) {
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
		"user_id":  string(createdUser.ID),
		"username": string(createdUser.Username),
		"email":    string(createdUser.Email),
	})
}

func (uh *UserHandler) UpdateUser(c *gin.Context) {
	ctx := c.Request.Context()

	authUserID, ok := getUserID(c)
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

	username, err := userdomain.NewUsername(req.Username)
	email, err := userdomain.NewEmail(req.Email)
	password, err := userdomain.NewPassword(req.Password)

	if errors.Is(err, exception.ErrInvalid) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_INPUT",
			"message": err.Error(),
		})
		return
	}

	updatedUser, err := uh.userUsecase.UpdateUser(
		ctx,
		authUserID,
		username,
		email,
		password,
	)

	if errors.Is(err, exception.ErrNotFound) {
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

	c.JSON(http.StatusOK, gin.H{
		"user_id":  string(updatedUser.ID),
		"username": string(updatedUser.Username),
		"email":    string(updatedUser.Email),
	})
}

func (uh *UserHandler) DeleteUser(c *gin.Context) {
	ctx := c.Request.Context()

	authUserID, ok := getUserID(c)
	if !ok {
		return
	}

	err := uh.userUsecase.DeleteUser(ctx, authUserID)
	if errors.Is(err, exception.ErrNotFound) {
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
	c.Status(http.StatusNoContent)
}
