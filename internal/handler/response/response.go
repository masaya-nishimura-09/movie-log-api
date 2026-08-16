package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func InvalidInput(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{
		"code":    "INVALID_INPUT",
		"message": err.Error(),
	})
}

func MalformedBody(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{
		"code":    "INVALID_INPUT",
		"message": "malformed request body",
	})
}

func InvalidAccessToken(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, gin.H{
		"code":    "INVALID_ACCESS_TOKEN",
		"message": "invalid or expired access token",
	})
}

func InvalidRefreshToken(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, gin.H{
		"code":    "INVALID_REFRESH_TOKEN",
		"message": "invalid or expired refresh token",
	})
}

func InvalidCredentials(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, gin.H{
		"code":    "INVALID_CREDENTIALS",
		"message": "invalid email or password",
	})
}

func Unauthenticated(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, gin.H{
		"code":    "UNAUTHENTICATED",
		"message": "authentication required",
	})
}

func UserNotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{
		"code":    "USER_NOT_FOUND",
		"message": "user not found",
	})
}

func UserAlreadyExists(c *gin.Context) {
	c.JSON(http.StatusConflict, gin.H{
		"code":    "USER_ALREADY_EXISTS",
		"message": "user with this email already exists",
	})
}

func InternalServerError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"code":    "INTERNAL_SERVER_ERROR",
		"message": "internal server error",
	})
}
