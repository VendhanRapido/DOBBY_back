package testutils

import (
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
)

func SetupTestRouter(registerRoutes func(r *gin.Engine)) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	registerRoutes(router)
	return router
}

func PerformRequest(r http.Handler, method, path string, requestByteArray io.Reader) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, requestByteArray)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
