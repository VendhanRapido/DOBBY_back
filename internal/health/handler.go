package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
}

type Response struct {
	Client string `json:"client,omitempty"`
	Status string `json:"status"`
}

const (
	UP string = "UP"
)

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) CheckSanity(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, Response{Status: UP})
}

func (h *Handler) CheckHealth(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, Response{Status: UP})
}
