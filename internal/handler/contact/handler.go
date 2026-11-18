package contact

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	contactdomain "github.com/masaya-nishimura-09/movie-log-api/internal/domain/contact"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	maildomain "github.com/masaya-nishimura-09/movie-log-api/internal/domain/mail"
	"github.com/masaya-nishimura-09/movie-log-api/internal/handler/response"
	"github.com/masaya-nishimura-09/movie-log-api/internal/logging"
	contactusecase "github.com/masaya-nishimura-09/movie-log-api/internal/usecase/contact"
)

type ContactReq struct {
	Email   string `json:"email"`
	Message string `json:"message"`
}

func (r ContactReq) toDomain() (
	maildomain.Address,
	contactdomain.Message,
	error,
) {
	email, err := maildomain.NewAddress(r.Email)
	if err != nil {
		return "", "", err
	}
	message, err := contactdomain.NewMessage(r.Message)
	if err != nil {
		return "", "", err
	}
	return email, message, nil
}

type ContactHandler struct {
	contactUsecase contactusecase.Usecase
}

func NewContactHandler(usecase contactusecase.Usecase) *ContactHandler {
	return &ContactHandler{contactUsecase: usecase}
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

func (ch *ContactHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()

	var req ContactReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.MalformedBody(c)
		return
	}

	language, ok := getLanguage(c)
	if !ok {
		return
	}

	email, message, err := req.toDomain()
	if errors.Is(err, exception.ErrInvalid) {
		response.InvalidInput(c, err)
		return
	}
	if err != nil {
		logging.Error(c, "request_failed", err)
		response.InternalServerError(c)
		return
	}

	if err := ch.contactUsecase.Send(ctx, email, language, message); err != nil {
		logging.Error(c, "request_failed", err)
		response.InternalServerError(c)
		return
	}

	c.Status(http.StatusNoContent)
}
