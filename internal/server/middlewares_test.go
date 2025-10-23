package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/roppenlabs/dobby-service/internal/accesscontrol"
	entitiesclient "github.com/roppenlabs/dobby-service/internal/clients/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type MiddlewareTestSuite struct {
	suite.Suite
	mockCtrl     *gomock.Controller
	mockEnforcer *accesscontrol.MockAccessControlService
	mockMapper   *accesscontrol.MockActionMapping
	mockEntities *entitiesclient.MockEntitiesClient
	middleware   Middleware
	router       *gin.Engine
}

func (s *MiddlewareTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	s.mockCtrl = gomock.NewController(s.T())
	s.mockEnforcer = accesscontrol.NewMockAccessControlService(s.mockCtrl)
	s.mockMapper = accesscontrol.NewMockActionMapping(s.mockCtrl)
	s.mockEntities = entitiesclient.NewMockEntitiesClient(s.mockCtrl)

	s.mockMapper.EXPECT().
		LoadRBACMap(gomock.Any()).
		Return(nil)

	s.middleware = NewMiddleware(s.mockEnforcer, s.mockMapper, s.mockEntities)
	s.router = gin.New()
}

func (s *MiddlewareTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *MiddlewareTestSuite) TestAuthenticate_Success() {
	// Arrange
	s.mockEntities.EXPECT().
		GetRolesFromEmail("test@rapido.com").
		Return([]string{"admin"}, nil)

	s.router.Use(s.middleware.Authenticate())
	s.router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Act
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-User-Email", "test@rapido.com")
	resp := httptest.NewRecorder()

	s.router.ServeHTTP(resp, req)

	// Assert
	assert.Equal(s.T(), http.StatusOK, resp.Code)
}

func (s *MiddlewareTestSuite) TestAuthenticate_MissingEmail() {
	// Arrange
	s.router.Use(s.middleware.Authenticate())
	s.router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Act
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	resp := httptest.NewRecorder()
	s.router.ServeHTTP(resp, req)

	// Assert
	assert.Equal(s.T(), http.StatusBadRequest, resp.Code)
}

func (s *MiddlewareTestSuite) TestAuthenticate_NoRoles() {
	// Arrange
	s.mockEntities.EXPECT().
		GetRolesFromEmail("test@rapido.com").
		Return([]string{}, nil)

	s.router.Use(s.middleware.Authenticate())
	s.router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Act
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-User-Email", "test@rapido.com")
	resp := httptest.NewRecorder()

	s.router.ServeHTTP(resp, req)

	// Assert
	assert.Equal(s.T(), http.StatusUnauthorized, resp.Code)
}

func (s *MiddlewareTestSuite) TestModuleListing_Success() {
	// Arrange
	userCtx := &UserContext{email: "test@rapido.com", roles: []string{"admin"}}
	s.mockEnforcer.EXPECT().
		GetPermittedModulesForRoles(userCtx.roles).
		Return([]string{"mod1", "mod2"})

	s.router.Use(func(c *gin.Context) {
		c.Set("user", userCtx)
	}, s.middleware.ModuleListing())

	s.router.GET("/modules", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"modules": c.MustGet("modules")})
	})

	// Act
	req, _ := http.NewRequest(http.MethodGet, "/modules", nil)
	resp := httptest.NewRecorder()
	s.router.ServeHTTP(resp, req)

	// Assert
	assert.Equal(s.T(), http.StatusOK, resp.Code)
}

func (s *MiddlewareTestSuite) TestOperationListing_Success() {
	// Arrange
	userCtx := &UserContext{email: "test@rapido.com", roles: []string{"admin"}}
	s.mockEnforcer.EXPECT().
		GetPermittedActionsForRoles(userCtx.roles, "mod123").
		Return([]string{"create", "read"})

	s.router.Use(func(c *gin.Context) {
		c.Set("user", userCtx)
	}, s.middleware.OperationListing())

	s.router.GET("/modules/:moduleId/operations", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"allowed": c.MustGet("allowedActions")})
	})

	// Act
	req, _ := http.NewRequest(http.MethodGet, "/modules/mod123/operations", nil)
	resp := httptest.NewRecorder()
	s.router.ServeHTTP(resp, req)

	// Assert
	assert.Equal(s.T(), http.StatusOK, resp.Code)
}

func (s *MiddlewareTestSuite) TestOperationListing_NoAccess() {
	// Arrange
	userCtx := &UserContext{email: "test@rapido.com", roles: []string{"read"}}
	s.mockEnforcer.EXPECT().
		GetPermittedActionsForRoles(userCtx.roles, "mod123").
		Return([]string{})

	s.router.Use(func(c *gin.Context) {
		c.Set("user", userCtx)
	}, s.middleware.OperationListing())

	s.router.GET("/modules/:moduleId/operations", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Act
	req, _ := http.NewRequest(http.MethodGet, "/modules/mod123/operations", nil)
	resp := httptest.NewRecorder()
	s.router.ServeHTTP(resp, req)

	// Assert
	assert.Equal(s.T(), http.StatusForbidden, resp.Code)
}

func (s *MiddlewareTestSuite) TestEnforceRouteAuthorization_Success() {
	// Arrange
	userCtx := &UserContext{email: "test@rapido.com", roles: []string{"admin"}}
	s.mockMapper.EXPECT().
		GetValidAction("mod123", "op456").
		Return("write", nil)
	s.mockEnforcer.EXPECT().
		VerifyActionForRoles(userCtx.roles, "mod123", "write").
		Return(true)

	s.router.Use(func(c *gin.Context) {
		c.Set("user", userCtx)
	}, s.middleware.EnforceRouteAuthorization())

	s.router.GET("/modules/:moduleId/operations/:operationId", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "authorized"})
	})

	// Act
	req, _ := http.NewRequest(http.MethodGet, "/modules/mod123/operations/op456", nil)
	resp := httptest.NewRecorder()
	s.router.ServeHTTP(resp, req)

	// Assert
	assert.Equal(s.T(), http.StatusOK, resp.Code)
}

func (s *MiddlewareTestSuite) TestEnforceRouteAuthorization_Unauthorized() {
	// Arrange
	userCtx := &UserContext{email: "test@rapido.com", roles: []string{"admin"}}
	s.mockMapper.EXPECT().
		GetValidAction("mod123", "op456").
		Return("write", nil)
	s.mockEnforcer.EXPECT().
		VerifyActionForRoles(userCtx.roles, "mod123", "write").
		Return(false)

	s.router.Use(func(c *gin.Context) {
		c.Set("user", userCtx)
	}, s.middleware.EnforceRouteAuthorization())

	s.router.GET("/modules/:moduleId/operations/:operationId", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "unauthorized"})
	})

	// Act
	req, _ := http.NewRequest(http.MethodGet, "/modules/mod123/operations/op456", nil)
	resp := httptest.NewRecorder()
	s.router.ServeHTTP(resp, req)

	// Assert
	assert.Equal(s.T(), http.StatusForbidden, resp.Code)
}

func (s *MiddlewareTestSuite) TestEnforceRouteAuthorization_MapperError() {
	// Arrange
	userCtx := &UserContext{email: "test@rapido.com", roles: []string{"admin"}}
	s.mockMapper.EXPECT().
		GetValidAction("mod123", "op456").
		Return("", errors.New("mapping error"))

	s.router.Use(func(c *gin.Context) {
		c.Set("user", userCtx)
	}, s.middleware.EnforceRouteAuthorization())

	s.router.GET("/modules/:moduleId/operations/:operationId", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "unauthorized"})
	})

	// Act
	req, _ := http.NewRequest(http.MethodGet, "/modules/mod123/operations/op456", nil)
	resp := httptest.NewRecorder()
	s.router.ServeHTTP(resp, req)

	// Assert
	assert.Equal(s.T(), http.StatusNotFound, resp.Code)
}

func TestMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, new(MiddlewareTestSuite))
}
