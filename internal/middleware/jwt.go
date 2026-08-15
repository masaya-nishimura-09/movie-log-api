package middleware

import (
	"errors"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	authdomain "github.com/masaya-nishimura-09/movie-log-api/internal/domain/auth"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/handler/response"
	authusecase "github.com/masaya-nishimura-09/movie-log-api/internal/usecase/auth"
	userusecase "github.com/masaya-nishimura-09/movie-log-api/internal/usecase/user"
)

func JWTAuth(
	authUsecase authusecase.Usecase,
	userUsecase userusecase.Usecase,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		authorization := c.GetHeader("Authorization")
		if authorization == "" || !strings.HasPrefix(authorization, "Bearer ") {
			response.Unauthenticated(c)
			c.Abort()
			return
		}

		accessToken := authdomain.AccessToken{
			Value: authdomain.AccessTokenValue(
				strings.TrimPrefix(authorization, "Bearer "),
			),
		}
		principal, err := authUsecase.ValidateAccessToken(ctx, &accessToken)
		if err != nil {
			response.InvalidAccessToken(c)
			c.Abort()
			return
		}

		_, err = userUsecase.GetByID(ctx, principal.UserID)
		if errors.Is(err, exception.ErrNotFound) {
			response.Unauthenticated(c)
			c.Abort()
			return
		}

		if err != nil {
			log.Println(err)
			response.InternalServerError(c)
			c.Abort()
			return
		}

		c.Set("userID", principal.UserID)
		c.Set("role", principal.Role)
		c.Next()
	}
}
