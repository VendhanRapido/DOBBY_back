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

	t.Run("should return hello world response successfully", func(t *testing.T) {
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
		mockService.On("HelloWorld").Return("")

		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		handler.helloWorldHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "", response)

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

	t.Run("should work with actual Gin router", func(t *testing.T) {
		mockService := &MockService{}
		expectedMessage := "Hello World!!!"
		mockService.On("HelloWorld").Return(expectedMessage)

		handler := NewHandler(mockService)
		router := gin.New()
		router.GET("/hello-world", handler.helloWorldHandler)

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

func TestHandler_EdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should handle service returning JSON-unsafe characters", func(t *testing.T) {
		mockService := &MockService{}
		expectedMessage := "Hello World with \"quotes\" and \n newlines and \t tabs"
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

	t.Run("should handle service returning unicode characters", func(t *testing.T) {
		mockService := &MockService{}
		expectedMessage := "Hello 世界! 🌍 测试"
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
		expectedMessage := "This is a very long message that contains multiple words and should be handled properly by the JSON marshaling and unmarshaling process without any issues. " +
			"Let's add some more content to make it even longer and test the robustness of our handler. " +
			"This message should be properly serialized and deserialized without any problems."
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
}

func TestHandler_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should work with real service implementation", func(t *testing.T) {
		realService := NewService()
		handler := NewHandler(realService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		handler.helloWorldHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Hello World!!!", response)
	})

	t.Run("should work with router integration", func(t *testing.T) {
		realService := NewService()
		handler := NewHandler(realService)
		router := gin.New()
		router.GET("/hello-world", handler.helloWorldHandler)

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
}

func TestHandler_JSONSerialization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should properly serialize string response", func(t *testing.T) {
		mockService := &MockService{}
		expectedMessage := "Hello World!!!"
		mockService.On("HelloWorld").Return(expectedMessage)

		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		handler.helloWorldHandler(c)

		// Verify the response is valid JSON
		var response string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, expectedMessage, response)

		// Verify the JSON is properly quoted
		jsonStr := w.Body.String()
		assert.Equal(t, `"Hello World!!!"`, jsonStr)

		mockService.AssertExpectations(t)
	})

	t.Run("should handle empty string serialization", func(t *testing.T) {
		mockService := &MockService{}
		mockService.On("HelloWorld").Return("")

		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		handler.helloWorldHandler(c)

		var response string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "", response)
		assert.Equal(t, `""`, w.Body.String())

		mockService.AssertExpectations(t)
	})
}