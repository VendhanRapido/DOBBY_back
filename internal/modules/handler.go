package modules

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/roppenlabs/dobby-service/internal/accesscontrol"
	"github.com/roppenlabs/dobby-service/internal/types"
)

type Handler struct {
	service  Service
	enforcer accesscontrol.AccessControlService
}

func NewHandler(s Service, e accesscontrol.AccessControlService) *Handler {
	return &Handler{
		service:  s,
		enforcer: e,
	}
}

// GET /modules - List all modules
func (h *Handler) GetModulesHandler(ctx *gin.Context) {
	userModulesVal, _ := ctx.Get("modules")

	userModules, ok := userModulesVal.([]string)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, types.InternalServerError("invalid type for modules"))
		return
	}

	modules, err := h.service.GetModules(userModules)
	if err != nil {
		if errorResp, ok := err.(*types.ErrorResponse); ok {
			ctx.JSON(errorResp.ErrorInfo.HTTPCode, errorResp)
		} else {
			ctx.JSON(http.StatusInternalServerError, err)
		}
		return
	}
	apiResponse := types.ResponseData{
		Data: modules,
	}
	ctx.JSON(http.StatusOK, apiResponse)
}

// GET /modules/:moduleId/operations - List operations for a module
func (h *Handler) GetOperationsHandler(ctx *gin.Context) {
	moduleID := ctx.Param("moduleId")
	if moduleID == "" {
		ctx.JSON(http.StatusBadRequest, types.BadRequestError("moduleId is required", MISSING_MODULE_ID, "Module Id is required"))
		return
	}

	actionsAny, exists := ctx.Get("allowedActions")
	if !exists {
		ctx.JSON(http.StatusInternalServerError, types.InternalServerError("Allowed actions not found in context"))
		return
	}

	actionsToCheck := actionsAny.([]string)
	filteredOps, err := h.service.GetFilteredOperations(moduleID, actionsToCheck)
	if err != nil {
		if errorResp, ok := err.(*types.ErrorResponse); ok {
			ctx.JSON(errorResp.ErrorInfo.HTTPCode, errorResp)
		} else {
			ctx.JSON(http.StatusInternalServerError, err)
		}
		return
	}
	ctx.JSON(http.StatusOK, types.ResponseData{
		Data: filteredOps,
	})

}

// GET /modules/:moduleId/operations/:operationId - Get operation details with schema
func (h *Handler) GetOperationDetailHandler(ctx *gin.Context) {
	moduleID := ctx.Param("moduleId")
	operationID := ctx.Param("operationId")

	if moduleID == "" {
		ctx.JSON(http.StatusBadRequest, types.BadRequestError("moduleId is required", MISSING_MODULE_ID, "Module Id is required"))
		return
	}

	if operationID == "" {
		ctx.JSON(http.StatusBadRequest, types.BadRequestError("operationId is required", MISSING_OPERATION_ID, "Operation Id is required"))
		return
	}

	operationDetail, err := h.service.GetOperationDetail(moduleID, operationID)
	if err != nil {
		if errorResp, ok := err.(*types.ErrorResponse); ok {
			ctx.JSON(errorResp.ErrorInfo.HTTPCode, errorResp)
		} else {
			ctx.JSON(http.StatusInternalServerError, err)
		}
		return
	}
	apiResponse := types.ResponseData{
		Data: operationDetail,
	}
	ctx.JSON(http.StatusOK, apiResponse)
}
