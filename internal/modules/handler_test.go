package modules

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/roppenlabs/dobby-service/internal/accesscontrol"
	"github.com/roppenlabs/dobby-service/internal/modules/testutils"
	"github.com/roppenlabs/dobby-service/internal/types"
)

type OperationsTestSuite struct {
	suite.Suite
	ctrl         *gomock.Controller
	mockService  *MockService
	mockEnforcer *accesscontrol.MockAccessControlService
	handler      *Handler
	router       *gin.Engine
}

func injectContextMiddleware(userRoles []string, allowedActions []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("userRoles", userRoles)
		c.Set("allowedActions", allowedActions)
		c.Next()
	}
}

func (s *OperationsTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockService = NewMockService(s.ctrl)
	s.mockEnforcer = accesscontrol.NewMockAccessControlService(s.ctrl)

	s.handler = NewHandler(s.mockService, s.mockEnforcer)
	s.router = testutils.SetupTestRouter()
	s.router.GET("/api/v0/modules", s.handler.GetModulesHandler)

	s.router.GET(
		"/api/v0/modules/:moduleId/operations",
		injectContextMiddleware([]string{"jrDev"}, []string{"read", "write"}),
		s.handler.GetOperationsHandler,
	)

	s.router.GET("/api/v0/modules/:moduleId/operations/:operationId", s.handler.GetOperationDetailHandler)
}

func (s *OperationsTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *OperationsTestSuite) TestNewHandler() {

	assert.NotNil(s.T(), s.handler)
	assert.Equal(s.T(), s.mockService, s.handler.service)
}

func (s *OperationsTestSuite) TestHandler_GetModulesHandler_Success() {
	expectedModules := []ModuleResponse{
		{
			ID:          "m1",
			Name:        "Test Module 1",
			Description: "Test Description 1",
		},
		{
			ID:          "m2",
			Name:        "Test Module 2",
			Description: "Test Description 2",
		},
	}

	s.mockService.EXPECT().
		GetModules([]string{"kafka"}).
		Return(expectedModules, nil)

	router := testutils.SetupTestRouterWithModules([]string{"kafka"}, s.handler)

	w := testutils.PerformRequest(router, "GET", "/api/v0/modules", nil)

	assert.Equal(s.T(), http.StatusOK, w.Code)

	var response struct {
		Data []ModuleResponse `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), expectedModules, response.Data)
}

func (s *OperationsTestSuite) TestHandler_GetModulesHandler_ServiceError() {
	s.mockService.EXPECT().
		GetModules([]string{"kafka"}).
		Return(nil, types.InternalServerError("internal server error"))

	router := testutils.SetupTestRouterWithModules([]string{"kafka"}, s.handler)

	w := testutils.PerformRequest(router, "GET", "/api/v0/modules", nil)

	assert.Equal(s.T(), http.StatusInternalServerError, w.Code)

	var response types.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "INTERNAL_ERROR", response.ErrorInfo.Code)
}

func (s *OperationsTestSuite) TestHandler_GetModulesHandler_InvalidModulesType() {
	router := testutils.SetupTestRouter()
	router.GET("/api/v0/modules", func(c *gin.Context) {
		c.Set("modules", "invalid-type")
		c.Next()
	}, s.handler.GetModulesHandler)

	w := testutils.PerformRequest(router, "GET", "/api/v0/modules", nil)

	assert.Equal(s.T(), http.StatusInternalServerError, w.Code)

	var response types.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "INTERNAL_ERROR", response.ErrorInfo.Code)
	assert.Contains(s.T(), response.ErrorInfo.Message, "invalid type for modules")
}

func (s *OperationsTestSuite) TestHandler_GetModulesHandler_GenericError() {
	s.mockService.EXPECT().
		GetModules([]string{"kafka"}).
		Return(nil, fmt.Errorf("some generic error"))

	router := testutils.SetupTestRouterWithModules([]string{"kafka"}, s.handler)

	w := testutils.PerformRequest(router, "GET", "/api/v0/modules", nil)

	assert.Equal(s.T(), http.StatusInternalServerError, w.Code)
}

func (s *OperationsTestSuite) TestHandler_GetOperationsHandler_Success() {
	moduleID := "m1"
	allowedActions := []string{"read", "write"}
	expectedOps := []Operations{
		{ID: "op1", Name: "Operation 1", Description: "desc", SchemaId: "schema1"},
	}

	s.mockService.EXPECT().
		GetFilteredOperations(moduleID, allowedActions).
		Return(expectedOps, nil)

	w := testutils.PerformRequest(s.router, "GET", fmt.Sprintf("/api/v0/modules/%s/operations", moduleID), nil)

	assert.Equal(s.T(), http.StatusOK, w.Code)
	var response struct {
		Data []Operations `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(s.T(), err)
	assert.ElementsMatch(s.T(), expectedOps, response.Data)
}

func (s *OperationsTestSuite) TestHandler_GetOperationsHandler_ModuleNotFound() {
	moduleID := "nonexistent"
	allowedActions := []string{"read", "write"}
	s.mockService.EXPECT().
		GetFilteredOperations(moduleID, allowedActions).
		Return(nil, types.NewNotFoundError("module not found", MODULE_NOT_FOUND, "The requested module could not be found"))

	w := testutils.PerformRequest(s.router, "GET", fmt.Sprintf("/api/v0/modules/%s/operations", moduleID), nil)

	assert.Equal(s.T(), http.StatusNotFound, w.Code)

	var response types.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "MODULE_NOT_FOUND", response.ErrorInfo.Code)
	assert.Equal(s.T(), "module not found", response.ErrorInfo.Message)
}
func (s *OperationsTestSuite) TestHandler_GetOperationsHandler_ServiceError() {
	moduleID := "m1"
	allowedActions := []string{"read", "write"}

	s.mockService.EXPECT().
		GetFilteredOperations(moduleID, allowedActions).
		Return(nil, types.InternalServerError("internal service error"))

	w := testutils.PerformRequest(s.router, "GET", fmt.Sprintf("/api/v0/modules/%s/operations", moduleID), nil)

	assert.Equal(s.T(), http.StatusInternalServerError, w.Code)

	var response types.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "INTERNAL_ERROR", response.ErrorInfo.Code)
}

func (s *OperationsTestSuite) TestHandler_GetOperationsHandler_EmptyModuleID() {
	w := testutils.PerformRequest(s.router, "GET", "/api/v0/modules//operations", nil)

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)

	var response types.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "MISSING_MODULE_ID", response.ErrorInfo.Code)
	assert.Equal(s.T(), "moduleId is required", response.ErrorInfo.Message)
}

func (s *OperationsTestSuite) TestHandler_GetOperationDetailHandler_Success() {
	moduleID := "m1"
	operationID := "op1"
	expectedResponse := &OperationDetailResponse{
		Name:        "Operation 1",
		Description: "Test operation",
		Schema: []FormField{
			{
				Name: "field1",
				Type: "string",
				Options: []Option{
					{Name: "option1", ID: "opt1"},
				},
			},
		},
	}

	s.mockService.EXPECT().
		GetOperationDetail(moduleID, operationID).
		Return(expectedResponse, nil)

	w := testutils.PerformRequest(s.router, "GET", "/api/v0/modules/m1/operations/op1", nil)

	assert.Equal(s.T(), http.StatusOK, w.Code)

	var response struct {
		Data OperationDetailResponse `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), *expectedResponse, response.Data)
}

func (s *OperationsTestSuite) TestHandler_GetOperationDetailHandler_ModuleNotFound() {
	moduleID := "nonexistent"
	operationID := "op1"

	s.mockService.EXPECT().
		GetOperationDetail(moduleID, operationID).
		Return(nil, types.NewNotFoundError("module not found", MODULE_NOT_FOUND, "The requested module could not be found"))

	w := testutils.PerformRequest(s.router, "GET", "/api/v0/modules/nonexistent/operations/op1", nil)

	assert.Equal(s.T(), http.StatusNotFound, w.Code)

	var response types.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "MODULE_NOT_FOUND", response.ErrorInfo.Code)
	assert.Equal(s.T(), "module not found", response.ErrorInfo.Message)
}

func (s *OperationsTestSuite) TestHandler_GetOperationDetailHandler_OperationNotFound() {
	moduleID := "m1"
	operationID := "nonexistent"

	s.mockService.EXPECT().
		GetOperationDetail(moduleID, operationID).
		Return(nil, types.NewNotFoundError("operation not found", OPERATION_NOT_FOUND, "The requested operation could not be found"))

	w := testutils.PerformRequest(s.router, "GET", "/api/v0/modules/m1/operations/nonexistent", nil)

	assert.Equal(s.T(), http.StatusNotFound, w.Code)

	var response types.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "OPERATION_NOT_FOUND", response.ErrorInfo.Code)
	assert.Equal(s.T(), "operation not found", response.ErrorInfo.Message)
}

func (s *OperationsTestSuite) TestHandler_GetOperationDetailHandler_SchemaNotFound() {
	moduleID := "m1"
	operationID := "op1"

	s.mockService.EXPECT().
		GetOperationDetail(moduleID, operationID).
		Return(nil, types.NewNotFoundError("schema not found", SCHEMA_NOT_FOUND, "The requested schema could not be found"))

	w := testutils.PerformRequest(s.router, "GET", "/api/v0/modules/m1/operations/op1", nil)

	assert.Equal(s.T(), http.StatusNotFound, w.Code)

	var response types.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "SCHEMA_NOT_FOUND", response.ErrorInfo.Code)
	assert.Equal(s.T(), "schema not found", response.ErrorInfo.Message)
}

func (s *OperationsTestSuite) TestHandler_GetOperationDetailHandler_EmptyModuleID() {
	w := testutils.PerformRequest(s.router, "GET", "/api/v0/modules//operations/op1", nil)

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)

	var response types.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "MISSING_MODULE_ID", response.ErrorInfo.Code)
	assert.Equal(s.T(), "moduleId is required", response.ErrorInfo.Message)
}

func (s *OperationsTestSuite) TestHandler_GetOperationDetailHandler_EmptyOperationID() {

	s.router.GET("api/v0/modules/:moduleId/operations/", s.handler.GetOperationDetailHandler)

	w := testutils.PerformRequest(s.router, "GET", "/api/v0/modules/m1/operations/", nil)

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)

	var response types.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "MISSING_OPERATION_ID", response.ErrorInfo.Code)
	assert.Equal(s.T(), "operationId is required", response.ErrorInfo.Message)
}

func (s *OperationsTestSuite) TestHandler_GetOperationDetailHandler_ServiceError() {
	moduleID := "m1"
	operationID := "op1"

	s.mockService.EXPECT().
		GetOperationDetail(moduleID, operationID).
		Return(nil, types.InternalServerError("internal service error"))

	w := testutils.PerformRequest(s.router, "GET", "/api/v0/modules/m1/operations/op1", nil)
	assert.Equal(s.T(), http.StatusInternalServerError, w.Code)

	var response types.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "INTERNAL_ERROR", response.ErrorInfo.Code)
}

func (s *OperationsTestSuite) TestHandler_GetOperationsHandler_AllowedActionsNotInContext() {
	s.router = testutils.SetupTestRouter()

	// No middleware - context is empty
	s.router.GET(
		"/api/v0/modules/:moduleId/operations",
		s.handler.GetOperationsHandler,
	)

	w := testutils.PerformRequest(s.router, "GET", "/api/v0/modules/m1/operations", nil)
	assert.Equal(s.T(), http.StatusInternalServerError, w.Code)

	var response types.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "INTERNAL_ERROR", response.ErrorInfo.Code)
	assert.Contains(s.T(), response.ErrorInfo.Message, "Allowed actions not found in context")
}

func (s *OperationsTestSuite) TestHandler_GetOperationsHandler_AccessDenied() {
	moduleID := "m1"
	allowedActions := []string{"read", "write"}
	s.mockService.EXPECT().
		GetFilteredOperations(moduleID, allowedActions).
		Return(nil, types.UnauthorizedError("Access denied", ACCESS_DENIED, "Access denied"))

	w := testutils.PerformRequest(s.router, "GET", fmt.Sprintf("/api/v0/modules/%s/operations", moduleID), nil)

	assert.Equal(s.T(), http.StatusForbidden, w.Code)

	var response types.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "ACCESS_DENIED", response.ErrorInfo.Code)
	assert.Equal(s.T(), "Access denied", response.ErrorInfo.Message)
}
func (s *OperationsTestSuite) TestHandler_GetOperationsHandler_GenericError() {
	moduleID := "m1"
	allowedActions := []string{"read", "write"}

	// Return a raw error (not types.ErrorResponse)
	s.mockService.EXPECT().
		GetFilteredOperations(moduleID, allowedActions).
		Return(nil, fmt.Errorf("some generic error"))

	w := testutils.PerformRequest(s.router, "GET", fmt.Sprintf("/api/v0/modules/%s/operations", moduleID), nil)

	assert.Equal(s.T(), http.StatusInternalServerError, w.Code)
}
func (s *OperationsTestSuite) TestHandler_GetOperationDetailHandler_GenericError() {
	moduleID := "m1"
	operationID := "op1"

	s.mockService.EXPECT().
		GetOperationDetail(moduleID, operationID).
		Return(nil, fmt.Errorf("something unexpected"))

	w := testutils.PerformRequest(s.router, "GET", fmt.Sprintf("/api/v0/modules/%s/operations/%s", moduleID, operationID), nil)

	assert.Equal(s.T(), http.StatusInternalServerError, w.Code)
}

// Run the test suite
func TestOperationsTestSuite(t *testing.T) {
	suite.Run(t, new(OperationsTestSuite))
}
