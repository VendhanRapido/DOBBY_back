package kafka

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/roppenlabs/dobby-service/internal/modules"
	kafkaservices "github.com/roppenlabs/dobby-service/internal/modules/kafka/services"
	kafkatypes "github.com/roppenlabs/dobby-service/internal/modules/kafka/types"
	"github.com/roppenlabs/dobby-service/internal/modules/kafka/validators"
	"github.com/roppenlabs/dobby-service/internal/types"
)

type Handler struct {
	service   kafkaservices.Service
	validator validators.Validator
}

func NewHandler(s kafkaservices.Service, v validators.Validator) *Handler {
	return &Handler{
		service:   s,
		validator: v,
	}
}
func (h *Handler) OperationsHandler(ctx *gin.Context) {
	operationID := ctx.Param("operationId")
	switch operationID {
	case "op-create-topic":
		h.CreateTopic(ctx)
	case "op-list-topics":
		h.ListTopics(ctx)
	case "op-delete-topic":
		h.DeleteTopic(ctx)
	default:
		ctx.JSON(http.StatusNotFound, types.BadRequestError("Operation doesnt exist", http.StatusText(http.StatusBadRequest), "Failed to reach endpoint"))
	}
}

func (h *Handler) CreateTopic(ctx *gin.Context) {
	var req kafkatypes.CreateTopicRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.BadRequestError(err.Error(), modules.INVALID_REQUEST_BODY, "Invalid request body"))
		return
	}
	ok, errorPayload := h.validator.ValidateCreateTopicRequest(&req)
	if !ok {
		ctx.JSON(http.StatusBadRequest, errorPayload)
		return
	}
	result, err := h.service.CreateTopic(&req)
	if err != nil {
		ctx.JSON(err.ErrorInfo.HTTPCode, err)
		return
	}
	ctx.JSON(http.StatusOK, result)
}

func (h *Handler) ListTopics(ctx *gin.Context) {
	var req kafkatypes.ListTopicsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.BadRequestError(err.Error(), modules.INVALID_REQUEST_BODY, "Invalid request"))
		return
	}
	result, err := h.service.ListTopics(&req)
	if err != nil {
		ctx.JSON(err.ErrorInfo.HTTPCode, err)
		return
	}
	ctx.JSON(http.StatusOK, result)
}
func (h *Handler) DeleteTopic(ctx *gin.Context) {
	var req kafkatypes.DeleteTopicRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.BadRequestError(err.Error(), modules.INVALID_REQUEST_BODY, "Invalid request"))
		return
	}
	result, err := h.service.DeleteTopic(&req)
	if err != nil {
		ctx.JSON(err.ErrorInfo.HTTPCode, err)
		return
	}
	ctx.JSON(http.StatusOK, result)
}
