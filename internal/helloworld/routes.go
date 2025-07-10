package helloworld

import "github.com/gin-gonic/gin"

func (h *Handler) InitRoutes(router *gin.Engine) {
	router.GET("/hello-world", h.helloWorldHandler)
}
