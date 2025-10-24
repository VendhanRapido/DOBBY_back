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
		mockService := NewMockService(t)
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

	t.Run("should return hello world response successfully", func(t *testing.T) {
		mockService := NewMockService(t)
		expectedResponse := "Hello World!!!"
		mockService.On("HelloWorld").Return(expectedResponse)

		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		handler.helloWorldHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

		var response string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, expectedResponse, response)

		mockService.AssertExpectations(t)
	})

	t.Run("should return different response from service", func(t *testing.T) {
		mockService := NewMockService(t)
		expectedResponse := "Custom Hello Message"
		mockService.On("HelloWorld").Return(expectedResponse)

		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		handler.helloWorldHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, expectedResponse, response)

		mockService.AssertExpectations(t)
	})

	t.Run("should handle empty response from service", func(t *testing.T) {
		mockService := NewMockService(t)
		expectedResponse := ""
		mockService.On("HelloWorld").Return(expectedResponse)

		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		handler.helloWorldHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, expectedResponse, response)

		mockService.AssertExpectations(t)
	})

	t.Run("should handle special characters in response", func(t *testing.T) {
		mockService := NewMockService(t)
		expectedResponse := "Hello 世界! 🌍"
		mockService.On("HelloWorld").Return(expectedResponse)

		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		handler.helloWorldHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, expectedResponse, response)

		mockService.AssertExpectations(t)
	})

	t.Run("should handle JSON escape characters", func(t *testing.T) {
		mockService := NewMockService(t)
		expectedResponse := "Hello \"World\" with\nnewlines\tand\ttabs"
		mockService.On("HelloWorld").Return(expectedResponse)

		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		handler.helloWorldHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, expectedResponse, response)

		mockService.AssertExpectations(t)
	})

	t.Run("should handle very long response", func(t *testing.T) {
		mockService := NewMockService(t)
		expectedResponse := "This is a very long response that contains many characters and should test the handler's ability to handle large responses without any issues. " +
			"Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. " +
			"Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. " +
			"Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. " +
			"Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."
		mockService.On("HelloWorld").Return(expectedResponse)

		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		handler.helloWorldHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, expectedResponse, response)

		mockService.AssertExpectations(t)
	})
}

func TestHandler_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should work with actual Gin router", func(t *testing.T) {
		mockService := NewMockService(t)
		expectedResponse := "Hello World!!!"
		mockService.On("HelloWorld").Return(expectedResponse)

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
		assert.Equal(t, expectedResponse, response)

		mockService.AssertExpectations(t)
	})

	t.Run("should handle multiple concurrent requests", func(t *testing.T) {
		mockService := NewMockService(t)
		expectedResponse := "Hello World!!!"
		mockService.On("HelloWorld").Return(expectedResponse).Times(10)

		handler := NewHandler(mockService)
		router := gin.New()
		handler.InitRoutes(router)

		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func() {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/hello-world", nil)
				router.ServeHTTP(w, req)

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

func TestHandler_EdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should panic with nil context", func(t *testing.T) {
		mockService := NewMockService(t)
		expectedResponse := "Hello World!!!"
		mockService.On("HelloWorld").Return(expectedResponse)

		handler := NewHandler(mockService)

		// This test verifies the handler panics with nil context
		// In real scenarios, Gin would never pass nil, but it's good to test
		assert.Panics(t, func() {
			handler.helloWorldHandler(nil)
		})
	})

	t.Run("should panic with context with no writer", func(t *testing.T) {
		mockService := NewMockService(t)
		expectedResponse := "Hello World!!!"
		mockService.On("HelloWorld").Return(expectedResponse)

		handler := NewHandler(mockService)
		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		// Reset the writer to nil to test edge case
		c.Writer = nil

		// This should panic
		assert.Panics(t, func() {
			handler.helloWorldHandler(c)
		})
	})

	t.Run("should handle service returning error-like string", func(t *testing.T) {
		mockService := NewMockService(t)
		expectedResponse := "error: something went wrong"
		mockService.On("HelloWorld").Return(expectedResponse)

		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		handler.helloWorldHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, expectedResponse, response)

		mockService.AssertExpectations(t)
	})
}

func TestHandler_Concurrency(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should be thread safe", func(t *testing.T) {
		mockService := NewMockService(t)
		expectedResponse := "Hello World!!!"
		mockService.On("HelloWorld").Return(expectedResponse).Times(100)

		handler := NewHandler(mockService)

		done := make(chan bool, 100)

		for i := 0; i < 100; i++ {
			go func() {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				handler.helloWorldHandler(c)

				assert.Equal(t, http.StatusOK, w.Code)
				done <- true
			}()
		}

		for i := 0; i < 100; i++ {
			<-done
		}

		mockService.AssertExpectations(t)
	})
}

func TestHandler_JSONResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should return valid JSON", func(t *testing.T) {
		mockService := NewMockService(t)
		expectedResponse := "Hello World!!!"
		mockService.On("HelloWorld").Return(expectedResponse)

		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		handler.helloWorldHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify it's valid JSON
		var response string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, expectedResponse, response)

		// Verify the JSON string is properly quoted
		jsonBytes := w.Body.Bytes()
		assert.True(t, len(jsonBytes) > 0)
		assert.Equal(t, byte('"'), jsonBytes[0])
		assert.Equal(t, byte('"'), jsonBytes[len(jsonBytes)-1])

		mockService.AssertExpectations(t)
	})

	t.Run("should handle unicode characters in JSON", func(t *testing.T) {
		mockService := NewMockService(t)
		expectedResponse := "Hello 世界! 🌍"
		mockService.On("HelloWorld").Return(expectedResponse)

		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		handler.helloWorldHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, expectedResponse, response)

		mockService.AssertExpectations(t)
	})
}