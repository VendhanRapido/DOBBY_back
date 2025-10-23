package server

import (
	"fmt"
	"log"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/roppenlabs/dobby-service/internal/accesscontrol"
	entitiesclient "github.com/roppenlabs/dobby-service/internal/clients/entities"
	"github.com/roppenlabs/dobby-service/internal/modules"
	"github.com/roppenlabs/dobby-service/internal/types"
)

type ResourceAction struct {
	Resource string
	Action   string
}

var TaskResource = map[string]*ResourceAction{
	"GET /api/v0/modules": {
		Resource: "task",
		Action:   "create",
	},
}

func ErrorResponse(c *gin.Context, message string, code string, httpCode int) {
	c.JSON(httpCode, types.ErrorResponse{
		ErrorInfo: types.ErrorContent{
			Message:        message,
			Code:           code,
			HTTPCode:       httpCode,
			DisplayMessage: message,
		},
	})
	c.Abort()
}

func Deny(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, types.NewNotFoundError(message, modules.OPERATION_NOT_FOUND, http.StatusText(http.StatusNotFound)))
	c.Abort()
}

//go:generate mockgen -destination middlewares_mock.go -package=server . Middleware

type Middleware interface {
	EnforceRouteAuthorization() gin.HandlerFunc
	OperationListing() gin.HandlerFunc
	ModuleListing() gin.HandlerFunc
	Authenticate() gin.HandlerFunc
}

type middleware struct {
	enforcer       accesscontrol.AccessControlService
	actionMapper   accesscontrol.ActionMapping
	entitiesClient entitiesclient.EntitiesClient
}

type UserContext struct {
	email string
	roles []string
}

func NewMiddleware(enforcer accesscontrol.AccessControlService, actionMap accesscontrol.ActionMapping, entitiesClient entitiesclient.EntitiesClient) Middleware {
	if err := actionMap.LoadRBACMap(modules.DatabaseYAML_FILE_PATH); err != nil {
		fmt.Printf("failed to initialize middleware: %v", err)
	}

	return &middleware{
		enforcer:       enforcer,
		actionMapper:   actionMap,
		entitiesClient: entitiesClient,
	}
}

func (m *middleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		email := c.GetHeader("X-User-Email")
		if email == "" {
			log.Println("Missing user email in request header")
			ErrorResponse(c, "Missing user email", modules.BAD_REQUEST, http.StatusBadRequest)
			return
		}

		roles, err := m.entitiesClient.GetRolesFromEmail(email)
		if err != nil || len(roles) == 0 {
			log.Printf("Failed to get roles for email %s: %v", email, err)
			ErrorResponse(c, "user not exist", modules.UNAUTHORIZED, http.StatusUnauthorized)
		}

		userCtx := &UserContext{
			email: email,
			roles: roles,
		}
		c.Set("user", userCtx)
		c.Next()
	}
}

func (m *middleware) OperationListing() gin.HandlerFunc {
	return func(c *gin.Context) {
		userCtxVal, _ := c.Get("user")

		userCtx, _ := userCtxVal.(*UserContext)

		moduleID := c.Param("moduleId")
		module := moduleID

		hasAccess := false

		allowedActions := m.enforcer.GetPermittedActionsForRoles(userCtx.roles, module)
		if len(allowedActions) > 0 {
			hasAccess = true
		}

		if !hasAccess {
			ErrorResponse(c, "User not authorized for any operations on this module", modules.UNAUTHORIZED, http.StatusForbidden)
		}
		c.Set("userRoles", userCtx.roles)
		c.Set("userEmail", userCtx.email)
		c.Set("allowedActions", allowedActions)
		c.Next()
	}
}
func (m *middleware) EnforceRouteAuthorization() gin.HandlerFunc {
	return func(c *gin.Context) {
		userCtxVal, _ := c.Get("user")
		userCtx, _ := userCtxVal.(*UserContext)

		module := c.Param("moduleId")
		operation := c.Param("operationId")

		action, err := m.actionMapper.GetValidAction(module, operation)

		if err != nil {
			fmt.Println("ActionMap error: ", err)
			Deny(c, "Something Went wrong")
			return
		}

		if m.enforcer.VerifyActionForRoles(userCtx.roles, module, action) {
			c.Next()
			return
		}

		ErrorResponse(c, "You have no access to this operation in this module", modules.ACCESS_DENIED, http.StatusForbidden)
	}
}

func (m *middleware) ModuleListing() gin.HandlerFunc {
	return func(c *gin.Context) {
		userCtxVal, _ := c.Get("user")

		userCtx, _ := userCtxVal.(*UserContext)

		allUserModules := m.enforcer.GetPermittedModulesForRoles(userCtx.roles)

		c.Set("modules", allUserModules)
		c.Set("roles", userCtx.roles)
		c.Next()
	}
}
