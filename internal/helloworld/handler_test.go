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
		mockService := new(MockService)
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
		mockService := new(MockService)
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
		mockService := new(MockService)
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
		mockService := new(MockService)
		specialMessage := "Hello World! 🚀 测试中文"
		mockService.On("HelloWorld").Return(specialMessage)

		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		handler.helloWorldHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, specialMessage, response)

		mockService.AssertExpectations(t)
	})

	t.Run("should handle service returning long message", func(t *testing.T) {
		mockService := new(MockService)
		longMessage := "This is a very long message that contains multiple words and should be handled properly by the JSON marshaling and unmarshaling process without any issues"
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

	t.Run("should handle multiple concurrent requests", func(t *testing.T) {
		mockService := new(MockService)
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

	t.Run("should handle service panic gracefully", func(t *testing.T) {
		mockService := new(MockService)
		mockService.On("HelloWorld").Panic("service panic")

		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		assert.Panics(t, func() {
			handler.helloWorldHandler(c)
		})

		mockService.AssertExpectations(t)
	})
}

func TestHandler_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should work with actual Gin router", func(t *testing.T) {
		mockService := new(MockService)
		expectedMessage := "Hello World!!!"
		mockService.On("HelloWorld").Return(expectedMessage)

		handler := NewHandler(mockService)
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
		assert.Equal(t, expectedMessage, response)

		mockService.AssertExpectations(t)
	})

	t.Run("should handle different HTTP methods", func(t *testing.T) {
		mockService := new(MockService)
		expectedMessage := "Hello World!!!"
		mockService.On("HelloWorld").Return(expectedMessage)

		handler := NewHandler(mockService)
		router := gin.New()
		router.GET("/hello-world", handler.helloWorldHandler)

		httpMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

		for _, method := range httpMethods {
			t.Run("method_"+method, func(t *testing.T) {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest(method, "/hello-world", nil)

				router.ServeHTTP(w, req)

				if method == "GET" {
					assert.Equal(t, http.StatusOK, w.Code)
					var response string
					err := json.Unmarshal(w.Body.Bytes(), &response)
					require.NoError(t, err)
					assert.Equal(t, expectedMessage, response)
				} else {
					assert.Equal(t, http.StatusNotFound, w.Code)
				}
			})
		}

		mockService.AssertExpectations(t)
	})
}

func TestHandler_EdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should handle nil context gracefully", func(t *testing.T) {
		mockService := new(MockService)
		handler := NewHandler(mockService)

		// This test ensures the handler doesn't panic with nil context
		// In real scenarios, Gin would never pass nil, but it's good to test
		assert.Panics(t, func() {
			handler.helloWorldHandler(nil)
		})
	})

	t.Run("should handle context with nil writer", func(t *testing.T) {
		mockService := new(MockService)
		handler := NewHandler(mockService)
		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		// Reset the writer to nil to test edge case
		c.Writer = nil

		// This should panic as we're trying to write to nil writer
		assert.Panics(t, func() {
			handler.helloWorldHandler(c)
		})
	})

	t.Run("should handle service returning nil", func(t *testing.T) {
		// Test with nil service
		handler := NewHandler(nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// This should panic as we're calling method on nil service
		assert.Panics(t, func() {
			handler.helloWorldHandler(c)
		})
	})
}

func TestHandler_JSONResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should return valid JSON", func(t *testing.T) {
		mockService := new(MockService)
		expectedMessage := "Hello World!!!"
		mockService.On("HelloWorld").Return(expectedMessage)

		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		handler.helloWorldHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify it's valid JSON
		var response string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, expectedMessage, response)

		mockService.AssertExpectations(t)
	})

	t.Run("should handle JSON with quotes and special characters", func(t *testing.T) {
		mockService := new(MockService)
		messageWithQuotes := `Hello "World" with 'quotes' and \backslashes\`
		mockService.On("HelloWorld").Return(messageWithQuotes)

		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		handler.helloWorldHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, messageWithQuotes, response)

		mockService.AssertExpectations(t)
	})

	t.Run("should handle JSON with newlines and tabs", func(t *testing.T) {
		mockService := new(MockService)
		messageWithNewlines := "Hello\nWorld\twith\ttabs"
		mockService.On("HelloWorld").Return(messageWithNewlines)

		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		handler.helloWorldHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, messageWithNewlines, response)

		mockService.AssertExpectations(t)
	})
}

func TestHandler_ServiceInteraction(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should call service exactly once", func(t *testing.T) {
		mockService := new(MockService)
		expectedMessage := "Hello World!!!"
		mockService.On("HelloWorld").Return(expectedMessage).Once()

		handler := NewHandler(mockService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		handler.helloWorldHandler(c)

		// Verify service was called exactly once
		mockService.AssertExpectations(t)
	})

	t.Run("should handle service returning different messages", func(t *testing.T) {
		mockService := new(MockService)
		messages := []string{
			"Hello World!!!",
			"Hello Universe!!!",
			"Hello Galaxy!!!",
			"Hello Cosmos!!!",
		}

		for i, message := range messages {
			mockService.On("HelloWorld").Return(message).Once()

			handler := NewHandler(mockService)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			handler.helloWorldHandler(c)

			assert.Equal(t, http.StatusOK, w.Code)

			var response string
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Equal(t, message, response, "Failed for message %d: %s", i, message)
		}

		mockService.AssertExpectations(t)
	})
}

func TestHandler_InitRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should register hello world route", func(t *testing.T) {
		mockService := new(MockService)
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

	t.Run("should not register other routes", func(t *testing.T) {
		mockService := new(MockService)
		handler := NewHandler(mockService)
		router := gin.New()
		
		handler.InitRoutes(router)

		// Test that other routes are not registered
		testRoutes := []string{
			"/health",
			"/sanity",
			"/api/hello",
			"/hello",
		}

		for _, route := range testRoutes {
			t.Run("route_"+route, func(t *testing.T) {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", route, nil)

				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusNotFound, w.Code)
			})
		}
	})

	t.Run("should work with fresh router each time", func(t *testing.T) {
		mockService := new(MockService)
		expectedMessage := "Hello World!!!"
		mockService.On("HelloWorld").Return(expectedMessage).Times(2)

		handler := NewHandler(mockService)
		
		// Test with first router
		router1 := gin.New()
		handler.InitRoutes(router1)
		
		w1 := httptest.NewRecorder()
		req1, _ := http.NewRequest("GET", "/hello-world", nil)
		router1.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusOK, w1.Code)

		// Test with second router
		router2 := gin.New()
		handler.InitRoutes(router2)
		
		w2 := httptest.NewRecorder()
		req2, _ := http.NewRequest("GET", "/hello-world", nil)
		router2.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusOK, w2.Code)

		mockService.AssertExpectations(t)
	})
}

func TestService_Implementation(t *testing.T) {
	t.Run("should create service implementation", func(t *testing.T) {
		service := NewService()
		assert.NotNil(t, service)
	})

	t.Run("should return hello world message", func(t *testing.T) {
		service := NewService()
		message := service.HelloWorld()
		assert.Equal(t, "Hello World!!!", message)
	})

	t.Run("should return consistent message", func(t *testing.T) {
		service := NewService()
		
		// Call multiple times to ensure consistency
		for i := 0; i < 5; i++ {
			message := service.HelloWorld()
			assert.Equal(t, "Hello World!!!", message)
		}
	})

	t.Run("should work with different service instances", func(t *testing.T) {
		service1 := NewService()
		service2 := NewService()
		
		message1 := service1.HelloWorld()
		message2 := service2.HelloWorld()
		
		assert.Equal(t, "Hello World!!!", message1)
		assert.Equal(t, "Hello World!!!", message2)
		assert.Equal(t, message1, message2)
	})
}

func TestHandler_WithRealService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should work with real service implementation", func(t *testing.T) {
		realService := NewService()
		handler := NewHandler(realService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		handler.helloWorldHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

		var response string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Hello World!!!", response)
	})

	t.Run("should work with real service in router", func(t *testing.T) {
		realService := NewService()
		handler := NewHandler(realService)
		router := gin.New()
		
		handler.InitRoutes(router)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/hello-world", nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Hello World!!!", response)
	})
}