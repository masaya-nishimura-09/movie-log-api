package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/model"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/repository"
)

func JWTAuth(
	userRepo repository.UserRepository,
	tokenService repository.TokenRepository,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHENTICATED",
				"message": "missing or malformed authorization header",
			})
			return
		}

		token := model.Token(strings.TrimPrefix(auth, "Bearer "))
		principal, err := tokenService.Validate(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "INVALID_TOKEN",
				"message": "invalid or expired token",
			})
			return
		}

		_, err = userRepo.GetByID(c.Request.Context(), principal.UserID)
		if errors.Is(err, model.ErrUserNotFound) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "USER_NOT_FOUND",
				"message": "user no longer exists",
			})
			return
		}

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "internal server error",
			})
			return
		}

		c.Set("userID", principal.UserID)
		c.Set("role", principal.Role)
		c.Next()
	}
}
