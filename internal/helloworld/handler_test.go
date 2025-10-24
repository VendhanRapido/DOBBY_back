package helloworld

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHandler(t *testing.T) {
	t.Run("should create handler with service", func(t *testing.T) {
		mockService := &MockService{}
		handler := NewHandler(mockService)

		assert.NotNil(t, handler)
		assert.Equal(t, mockService, handler.service)
	})

	t.Run("should create handler with nil service", func(t *testing.T) {
		handler := NewHandler(nil)

		assert.NotNil(t, handler)
		assert.Nil(t, handler.service)
	})
}

func TestHandler_helloWorldHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should return hello world message successfully", func(t *testing.T) {
		mockService := &MockService{}
		expectedMessage := "Hello World!!!"
		mockService.On("HelloWorld").Return(expectedMessage)

		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		handler.helloWorldHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

		var response string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, expectedMessage, response)

		mockService.AssertExpectations(t)
	})

	t.Run("should handle service returning empty string", func(t *testing.T) {
		mockService := &MockService{}
		expectedMessage := ""
		mockService.On("HelloWorld").Return(expectedMessage)

		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		handler.helloWorldHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, expectedMessage, response)

		mockService.AssertExpectations(t)
	})

	t.Run("should handle service returning special characters", func(t *testing.T) {
		mockService := &MockService{}
		expectedMessage := "Hello World! 🎉 测试"
		mockService.On("HelloWorld").Return(expectedMessage)

		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		handler.helloWorldHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, expectedMessage, response)

		mockService.AssertExpectations(t)
	})

	t.Run("should handle service returning long message", func(t *testing.T) {
		mockService := &MockService{}
		expectedMessage := "This is a very long message that contains multiple words and should be handled properly by the JSON marshaling and unmarshaling process without any issues"
		mockService.On("HelloWorld").Return(expectedMessage)

		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		handler.helloWorldHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, expectedMessage, response)

		mockService.AssertExpectations(t)
	})

	t.Run("should handle multiple concurrent requests", func(t *testing.T) {
		mockService := &MockService{}
		expectedMessage := "Hello World!!!"
		mockService.On("HelloWorld").Return(expectedMessage).Times(10)

		handler := NewHandler(mockService)
		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func() {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				handler.helloWorldHandler(c)

				assert.Equal(t, http.StatusOK, w.Code)
				done <- true
			}()
		}

		for i := 0; i < 10; i++ {
			<-done
		}

		mockService.AssertExpectations(t)
	})

}

func TestHandler_InitRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should register hello world route", func(t *testing.T) {
		mockService := &MockService{}
		expectedMessage := "Hello World!!!"
		mockService.On("HelloWorld").Return(expectedMessage)

		handler := NewHandler(mockService)
		router := gin.New()
		handler.InitRoutes(router)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/hello-world", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

		var response string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, expectedMessage, response)

		mockService.AssertExpectations(t)
	})

	t.Run("should handle GET request to hello-world endpoint", func(t *testing.T) {
		mockService := &MockService{}
		expectedMessage := "Hello World!!!"
		mockService.On("HelloWorld").Return(expectedMessage)

		handler := NewHandler(mockService)
		router := gin.New()
		handler.InitRoutes(router)

		// Test that the route is properly registered
		routes := router.Routes()
		found := false
		for _, route := range routes {
			if route.Path == "/hello-world" && route.Method == "GET" {
				found = true
				break
			}
		}
		assert.True(t, found, "Route /hello-world should be registered")

		// Test the actual endpoint
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/hello-world", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("should return 404 for non-existent route", func(t *testing.T) {
		handler := NewHandler(&MockService{})
		router := gin.New()
		handler.InitRoutes(router)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/non-existent", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("should return 404 for POST request to hello-world", func(t *testing.T) {
		handler := NewHandler(&MockService{})
		router := gin.New()
		handler.InitRoutes(router)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/hello-world", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("should work with single route registration", func(t *testing.T) {
		mockService := &MockService{}
		expectedMessage := "Hello World!!!"
		mockService.On("HelloWorld").Return(expectedMessage)

		handler := NewHandler(mockService)
		router := gin.New()

		// Register the handler once
		handler.InitRoutes(router)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/hello-world", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, expectedMessage, response)

		mockService.AssertExpectations(t)
	})
}

func TestHandler_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should work end-to-end with real service", func(t *testing.T) {
		realService := NewService()
		handler := NewHandler(realService)
		router := gin.New()
		handler.InitRoutes(router)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/hello-world", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

		var response string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Hello World!!!", response)
	})

	t.Run("should handle multiple concurrent requests with real service", func(t *testing.T) {
		realService := NewService()
		handler := NewHandler(realService)
		router := gin.New()
		handler.InitRoutes(router)

		done := make(chan bool, 5)

		for i := 0; i < 5; i++ {
			go func() {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/hello-world", nil)
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusOK, w.Code)
				done <- true
			}()
		}

		for i := 0; i < 5; i++ {
			<-done
		}
	})
}

func TestHandler_EdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should handle service returning JSON-unsafe characters", func(t *testing.T) {
		mockService := &MockService{}
		expectedMessage := "Hello World!\n\t\"quoted\""
		mockService.On("HelloWorld").Return(expectedMessage)

		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		handler.helloWorldHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, expectedMessage, response)

		mockService.AssertExpectations(t)
	})

	t.Run("should handle service returning very long string", func(t *testing.T) {
		mockService := &MockService{}
		// Create a very long string
		longMessage := ""
		for i := 0; i < 1000; i++ {
			longMessage += "Hello World! "
		}
		mockService.On("HelloWorld").Return(longMessage)

		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		handler.helloWorldHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, longMessage, response)

		mockService.AssertExpectations(t)
	})
}