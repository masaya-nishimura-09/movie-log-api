package user

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	userdomain "github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
	"github.com/masaya-nishimura-09/movie-log-api/internal/handler/response"
	userusecase "github.com/masaya-nishimura-09/movie-log-api/internal/usecase/user"
)

type UserReq struct {
	Username string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r UserReq) toDomain() (
	userdomain.Username,
	userdomain.Email,
	userdomain.Password,
	error,
) {
	username, err := userdomain.NewUsername(r.Username)
	if err != nil {
		return "", "", "", err
	}
	email, err := userdomain.NewEmail(r.Email)
	if err != nil {
		return "", "", "", err
	}
	password, err := userdomain.NewPassword(r.Password)
	if err != nil {
		return "", "", "", err
	}
	return username, email, password, nil
}

type UserHandler struct {
	userUsecase userusecase.Usecase
}

func NewUserHandler(usecase userusecase.Usecase) *UserHandler {
	return &UserHandler{userUsecase: usecase}
}

func getUserID(c *gin.Context) (userdomain.ID, bool) {
	v, exists := c.Get("userID")
	id, ok := v.(userdomain.ID)
	if !exists || !ok {
		log.Println("userID in context is missing or not of type user.ID")
		response.InternalServerError(c)
		return 0, false
	}
	return id, true
}

func (uh *UserHandler) CreateUser(c *gin.Context) {
	ctx := c.Request.Context()

	var req UserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.MalformedBody(c)
		return
	}

	username, email, password, err := req.toDomain()
	if errors.Is(err, exception.ErrInvalid) {
		response.InvalidInput(c, err)
		return
	}
	if err != nil {
		log.Println(err)
		response.InternalServerError(c)
		return
	}

	createdUser, err := uh.userUsecase.Register(ctx, username, email, password)
	if errors.Is(err, exception.ErrAlreadyExists) {
		response.UserAlreadyExists(c)
		return
	}
	if err != nil {
		log.Println(err)
		response.InternalServerError(c)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user_id":  strconv.FormatUint(uint64(createdUser.ID), 10),
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

	var req UserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.MalformedBody(c)
		return
	}

	username, email, password, err := req.toDomain()
	if errors.Is(err, exception.ErrInvalid) {
		response.InvalidInput(c, err)
		return
	}
	if err != nil {
		log.Println(err)
		response.InternalServerError(c)
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
		response.UserNotFound(c)
		return
	}
	if err != nil {
		log.Println(err)
		response.InternalServerError(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":  strconv.FormatUint(uint64(updatedUser.ID), 10),
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
		response.UserNotFound(c)
		return
	}
	if err != nil {
		log.Println(err)
		response.InternalServerError(c)
		return
	}
	c.Status(http.StatusNoContent)
}
