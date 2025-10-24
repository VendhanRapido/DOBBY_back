package helloworld

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(s Service) *Handler {
	return &Handler{
		service: s,
	}
}

func (h *Handler) helloWorldHandler(ctx *gin.Context) {
	ctx.JSON( http.StatusOK , h.service.HelloWorld() )
}
