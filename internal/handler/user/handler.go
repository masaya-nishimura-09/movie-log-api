package user

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	maildomain "github.com/masaya-nishimura-09/movie-log-api/internal/domain/mail"
	userdomain "github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
	"github.com/masaya-nishimura-09/movie-log-api/internal/handler/response"
	"github.com/masaya-nishimura-09/movie-log-api/internal/logging"
	userusecase "github.com/masaya-nishimura-09/movie-log-api/internal/usecase/user"
)

type UserReq struct {
	Username string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateUsernameReq struct {
	Username string `json:"name"`
}

type UpdateEmailReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdatePasswordReq struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type VerifyEmailReq struct {
	Token string `json:"token"`
}

type ResendVerificationEmailReq struct {
	Email string `json:"email"`
}

type VerifyEmailChangeReq struct {
	Token string `json:"token"`
}

type RequestPasswordResetReq struct {
	Email string `json:"email"`
}

type ResetPasswordReq struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

func (r RequestPasswordResetReq) toDomain() (
	userdomain.Email,
	error,
) {
	email, err := userdomain.NewEmail(r.Email)
	if err != nil {
		return "", err
	}
	return email, nil
}

func (r ResetPasswordReq) toDomain() (
	userdomain.PasswordResetTokenValue,
	userdomain.Password,
	error,
) {
	value, err := userdomain.NewPasswordResetTokenValue(r.Token)
	if err != nil {
		return "", "", err
	}
	password, err := userdomain.NewPassword(r.Password)
	if err != nil {
		return "", "", err
	}
	return value, password, nil
}

func (r VerifyEmailChangeReq) toDomain() (
	userdomain.EmailChangeTokenValue,
	error,
) {
	value, err := userdomain.NewEmailChangeTokenValue(r.Token)
	if err != nil {
		return "", err
	}
	return value, nil
}

func (r VerifyEmailReq) toDomain() (
	userdomain.EmailVerificationTokenValue,
	error,
) {
	value, err := userdomain.NewEmailVerificationTokenValue(r.Token)
	if err != nil {
		return "", err
	}
	return value, nil
}

func (r ResendVerificationEmailReq) toDomain() (
	userdomain.Email,
	error,
) {
	email, err := userdomain.NewEmail(r.Email)
	if err != nil {
		return "", err
	}
	return email, nil
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

func (r UpdateUsernameReq) toDomain() (
	userdomain.Username,
	error,
) {
	username, err := userdomain.NewUsername(r.Username)
	if err != nil {
		return "", err
	}
	return username, nil
}

func (r UpdateEmailReq) toDomain() (
	userdomain.Password,
	userdomain.Email,
	error,
) {
	password, err := userdomain.NewPassword(r.Password)
	if err != nil {
		return "", "", err
	}
	email, err := userdomain.NewEmail(r.Email)
	if err != nil {
		return "", "", err
	}
	return password, email, nil
}

func (r UpdatePasswordReq) toDomain() (
	userdomain.Password,
	userdomain.Password,
	error,
) {
	oldPassword, err := userdomain.NewPassword(r.OldPassword)
	if err != nil {
		return "", "", err
	}
	newPassword, err := userdomain.NewPassword(r.NewPassword)
	if err != nil {
		return "", "", err
	}
	return oldPassword, newPassword, nil
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
		logging.Error(c, "request_failed", errors.New("userID in context is missing or not of type user.ID"))
		response.InternalServerError(c)
		return 0, false
	}
	return id, true
}

func getLanguage(c *gin.Context) (maildomain.Language, bool) {
	l := c.Query("language")
	if l == "" {
		l = "en"
	}
	language, err := maildomain.NewLanguage(l)
	if err != nil {
		response.InvalidInput(c, err)
		return "", false
	}
	return language, true
}

func (uh *UserHandler) GetByID(c *gin.Context) {
	ctx := c.Request.Context()

	authUserID, ok := getUserID(c)
	if !ok {
		return
	}

	u, err := uh.userUsecase.GetByID(ctx, authUserID)
	if errors.Is(err, exception.ErrNotFound) {
		response.UserNotFound(c)
		return
	}
	if err != nil {
		logging.Error(c, "request_failed", err)
		response.InternalServerError(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":  strconv.FormatUint(uint64(u.ID), 10),
		"username": string(u.Username),
		"email":    string(u.Email),
	})
}

func (uh *UserHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()

	var req UserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.MalformedBody(c)
		return
	}

	language, ok := getLanguage(c)
	if !ok {
		return
	}

	username, email, password, err := req.toDomain()
	if errors.Is(err, exception.ErrInvalid) {
		response.InvalidInput(c, err)
		return
	}
	if err != nil {
		logging.Error(c, "request_failed", err)
		response.InternalServerError(c)
		return
	}

	if err := uh.userUsecase.Create(ctx, username, email, password, language); err != nil {
		logging.Error(c, "request_failed", err)
		response.InternalServerError(c)
		return
	}

	c.Status(http.StatusNoContent)
}

func (uh *UserHandler) VerifyEmail(c *gin.Context) {
	ctx := c.Request.Context()

	var req VerifyEmailReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.MalformedBody(c)
		return
	}

	value, err := req.toDomain()
	if errors.Is(err, exception.ErrInvalid) {
		response.InvalidInput(c, err)
		return
	}
	if err != nil {
		logging.Error(c, "request_failed", err)
		response.InternalServerError(c)
		return
	}

	verifyEmailErr := uh.userUsecase.VerifyEmail(ctx, value)
	if errors.Is(verifyEmailErr, exception.ErrInvalid) {
		response.InvalidVerificationToken(c)
		return
	}
	if verifyEmailErr != nil {
		logging.Error(c, "request_failed", verifyEmailErr)
		response.InternalServerError(c)
		return
	}

	c.Status(http.StatusNoContent)
}

func (uh *UserHandler) ResendVerificationEmail(c *gin.Context) {
	ctx := c.Request.Context()

	var req ResendVerificationEmailReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.MalformedBody(c)
		return
	}

	language, ok := getLanguage(c)
	if !ok {
		return
	}

	email, err := req.toDomain()
	if errors.Is(err, exception.ErrInvalid) {
		response.InvalidInput(c, err)
		return
	}
	if err != nil {
		logging.Error(c, "request_failed", err)
		response.InternalServerError(c)
		return
	}

	if err := uh.userUsecase.ResendVerificationEmail(ctx, email, language); err != nil {
		logging.Error(c, "request_failed", err)
		response.InternalServerError(c)
		return
	}

	c.Status(http.StatusNoContent)
}

func (uh *UserHandler) UpdateUsername(c *gin.Context) {
	ctx := c.Request.Context()

	authUserID, ok := getUserID(c)
	if !ok {
		return
	}

	var req UpdateUsernameReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.MalformedBody(c)
		return
	}

	newUsername, err := req.toDomain()
	if errors.Is(err, exception.ErrInvalid) {
		response.InvalidInput(c, err)
		return
	}
	if err != nil {
		logging.Error(c, "request_failed", err)
		response.InternalServerError(c)
		return
	}

	updatedUser, updateUsernameErr := uh.userUsecase.UpdateUsername(
		ctx,
		authUserID,
		newUsername,
	)
	if errors.Is(updateUsernameErr, exception.ErrNotFound) {
		response.UserNotFound(c)
		return
	}
	if updateUsernameErr != nil {
		logging.Error(c, "request_failed", updateUsernameErr)
		response.InternalServerError(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":  strconv.FormatUint(uint64(updatedUser.ID), 10),
		"username": string(updatedUser.Username),
		"email":    string(updatedUser.Email),
	})
}

func (uh *UserHandler) UpdateEmail(c *gin.Context) {
	ctx := c.Request.Context()

	authUserID, ok := getUserID(c)
	if !ok {
		return
	}

	var req UpdateEmailReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.MalformedBody(c)
		return
	}

	language, ok := getLanguage(c)
	if !ok {
		return
	}

	password, newEmail, err := req.toDomain()
	if errors.Is(err, exception.ErrInvalid) {
		response.InvalidInput(c, err)
		return
	}
	if err != nil {
		logging.Error(c, "request_failed", err)
		response.InternalServerError(c)
		return
	}

	updateEmailErr := uh.userUsecase.UpdateEmail(
		ctx,
		authUserID,
		newEmail,
		password,
		language,
	)
	if errors.Is(updateEmailErr, exception.ErrNotFound) {
		response.UserNotFound(c)
		return
	}
	if errors.Is(updateEmailErr, exception.ErrInvalid) {
		response.InvalidCurrentPassword(c)
		return
	}
	if updateEmailErr != nil {
		logging.Error(c, "request_failed", updateEmailErr)
		response.InternalServerError(c)
		return
	}

	c.Status(http.StatusNoContent)
}

func (uh *UserHandler) VerifyEmailChange(c *gin.Context) {
	ctx := c.Request.Context()

	var req VerifyEmailChangeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.MalformedBody(c)
		return
	}

	value, err := req.toDomain()
	if errors.Is(err, exception.ErrInvalid) {
		response.InvalidInput(c, err)
		return
	}
	if err != nil {
		logging.Error(c, "request_failed", err)
		response.InternalServerError(c)
		return
	}

	verifyEmailChangeErr := uh.userUsecase.VerifyEmailChange(ctx, value)
	if errors.Is(verifyEmailChangeErr, exception.ErrInvalid) {
		response.InvalidVerificationToken(c)
		return
	}
	if errors.Is(verifyEmailChangeErr, exception.ErrAlreadyExists) {
		response.UserAlreadyExists(c)
		return
	}
	if verifyEmailChangeErr != nil {
		logging.Error(c, "request_failed", verifyEmailChangeErr)
		response.InternalServerError(c)
		return
	}

	c.Status(http.StatusNoContent)
}

func (uh *UserHandler) RequestPasswordReset(c *gin.Context) {
	ctx := c.Request.Context()

	var req RequestPasswordResetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.MalformedBody(c)
		return
	}

	language, ok := getLanguage(c)
	if !ok {
		return
	}

	email, err := req.toDomain()
	if errors.Is(err, exception.ErrInvalid) {
		response.InvalidInput(c, err)
		return
	}
	if err != nil {
		logging.Error(c, "request_failed", err)
		response.InternalServerError(c)
		return
	}

	if err := uh.userUsecase.RequestPasswordReset(ctx, email, language); err != nil {
		logging.Error(c, "request_failed", err)
		response.InternalServerError(c)
		return
	}

	c.Status(http.StatusNoContent)
}

func (uh *UserHandler) ResetPassword(c *gin.Context) {
	ctx := c.Request.Context()

	var req ResetPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.MalformedBody(c)
		return
	}

	value, newPassword, err := req.toDomain()
	if errors.Is(err, exception.ErrInvalid) {
		response.InvalidInput(c, err)
		return
	}
	if err != nil {
		logging.Error(c, "request_failed", err)
		response.InternalServerError(c)
		return
	}

	resetPasswordErr := uh.userUsecase.ResetPassword(ctx, value, newPassword)
	if errors.Is(resetPasswordErr, exception.ErrInvalid) {
		response.InvalidPasswordResetToken(c)
		return
	}
	if resetPasswordErr != nil {
		logging.Error(c, "request_failed", resetPasswordErr)
		response.InternalServerError(c)
		return
	}

	c.Status(http.StatusNoContent)
}

func (uh *UserHandler) UpdatePassword(c *gin.Context) {
	ctx := c.Request.Context()

	authUserID, ok := getUserID(c)
	if !ok {
		return
	}

	var req UpdatePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.MalformedBody(c)
		return
	}

	oldPassword, newPassword, err := req.toDomain()
	if errors.Is(err, exception.ErrInvalid) {
		response.InvalidInput(c, err)
		return
	}
	if err != nil {
		logging.Error(c, "request_failed", err)
		response.InternalServerError(c)
		return
	}

	updatedUser, updatePasswordErr := uh.userUsecase.UpdatePassword(
		ctx,
		authUserID,
		oldPassword,
		newPassword,
	)
	if errors.Is(updatePasswordErr, exception.ErrNotFound) {
		response.UserNotFound(c)
		return
	}
	if errors.Is(updatePasswordErr, exception.ErrInvalid) {
		response.InvalidCurrentPassword(c)
		return
	}
	if updatePasswordErr != nil {
		logging.Error(c, "request_failed", updatePasswordErr)
		response.InternalServerError(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":  strconv.FormatUint(uint64(updatedUser.ID), 10),
		"username": string(updatedUser.Username),
		"email":    string(updatedUser.Email),
	})
}

func (uh *UserHandler) Delete(c *gin.Context) {
	ctx := c.Request.Context()

	authUserID, ok := getUserID(c)
	if !ok {
		return
	}

	err := uh.userUsecase.Delete(ctx, authUserID)
	if errors.Is(err, exception.ErrNotFound) {
		response.UserNotFound(c)
		return
	}
	if err != nil {
		logging.Error(c, "request_failed", err)
		response.InternalServerError(c)
		return
	}
	c.Status(http.StatusNoContent)
}
