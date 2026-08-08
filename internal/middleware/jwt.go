package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	authdomain "github.com/masaya-nishimura-09/movie-log-api/internal/domain/auth"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	authusecase "github.com/masaya-nishimura-09/movie-log-api/internal/usecase/auth"
	userusecase "github.com/masaya-nishimura-09/movie-log-api/internal/usecase/user"
)

func JWTAuth(
	authUsecase *authusecase.AuthUsecase,
	userUsecase *userusecase.UserUsecase,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		authorization := c.GetHeader("Authorization")
		if authorization == "" || !strings.HasPrefix(authorization, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHENTICATED",
				"message": "missing or malformed authorization header",
			})
			return
		}

		accessToken := authdomain.AccessToken{
			Value: authdomain.AccessTokenValue(
				strings.TrimPrefix(authorization, "Bearer "),
			),
		}
		principal, err := authUsecase.ValidateAccessToken(ctx, &accessToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "INVALID_TOKEN",
				"message": "invalid or expired token",
			})
			return
		}

		_, err = userUsecase.GetByID(c.Request.Context(), principal.UserID)
		if errors.Is(err, exception.ErrUserNotFound) {
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
