package kafka

import (
	"encoding/json"

	"net/http"
	"strings"
	"testing"

	"github.com/roppenlabs/dobby-service/internal/modules"
	"github.com/roppenlabs/dobby-service/internal/modules/kafka/validators"
	"github.com/roppenlabs/dobby-service/internal/types"

	"github.com/gin-gonic/gin"
	kafkaservices "github.com/roppenlabs/dobby-service/internal/modules/kafka/services"
	kafkatypes "github.com/roppenlabs/dobby-service/internal/modules/kafka/types"
	"github.com/roppenlabs/dobby-service/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type HandlerTestSuite struct {
	suite.Suite
	ctrl          *gomock.Controller
	mockService   *kafkaservices.MockService
	mockValidator *validators.MockValidator
	router        *gin.Engine
}

func (h *HandlerTestSuite) BeforeTest(suiteName, testName string) {
	h.ctrl = gomock.NewController(h.T())
	h.mockService = kafkaservices.NewMockService((h.ctrl))
	h.mockValidator = validators.NewMockValidator(h.ctrl)
	handler := NewHandler(h.mockService, h.mockValidator)
	h.router = testutils.SetupTestRouter(func(r *gin.Engine) {
		r.POST("/api/v0/modules/:moduleId/operations/:operationId", handler.OperationsHandler)
	})
}

func (h *HandlerTestSuite) AfterTest(suiteName, testName string) {
	h.ctrl.Finish()
}

func TestHandlerSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}

func (h *HandlerTestSuite) TestUnknownOperationId() {
	expectedResponse := types.ErrorResponse{
		ErrorInfo: types.ErrorContent{
			Message:        "Operation doesnt exist",
			Code:           "Bad Request",
			HTTPCode:       http.StatusNotFound,
			DisplayMessage: "Failed to reach endpoint",
		},
	}
	w := testutils.PerformRequest(h.router, http.MethodPost, "/api/v0/modules/1/operations/unknown", testutils.FaultyRequestBodyReader{})

	assert.Equal(h.T(), http.StatusNotFound, w.Code, "Status code should be 404")

	var response types.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)

	assert.NoError(h.T(), err, "Should be able to get valid json response")
	assert.Equal(h.T(), expectedResponse.ErrorInfo.DisplayMessage, response.ErrorInfo.DisplayMessage, "The response should match")
	assert.Equal(h.T(), expectedResponse.ErrorInfo.Code, response.ErrorInfo.Code, "The response should match")
}

// List Topics handler tests
func (h *HandlerTestSuite) TestListTopicsHandlerWhenBodyReadFails() {

	expectedResponse := types.ErrorResponse{
		ErrorInfo: types.ErrorContent{
			Message:        "",
			Code:           modules.INVALID_REQUEST_BODY,
			HTTPCode:       http.StatusBadRequest,
			DisplayMessage: "Invalid request",
		},
	}
	w := testutils.PerformRequest(h.router, http.MethodPost, "/api/v0/modules/1/operations/op-list-topics", testutils.FaultyRequestBodyReader{})

	assert.Equal(h.T(), http.StatusBadRequest, w.Code, "Status code should be 400")

	var response types.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)

	assert.NoError(h.T(), err, "Should be able to get valid json response")
	assert.Equal(h.T(), expectedResponse.ErrorInfo.DisplayMessage, response.ErrorInfo.DisplayMessage, "The response should match")
	assert.Equal(h.T(), expectedResponse.ErrorInfo.Code, response.ErrorInfo.Code, "The response should match")
}

func (h *HandlerTestSuite) TestListHandlerMisformedJSON() {
	requestJSON := `{"clusterID": ""`
	expectedResponse := types.ErrorResponse{
		ErrorInfo: types.ErrorContent{
			Message:        "",
			Code:           "INVALID_REQUEST_BODY",
			HTTPCode:       http.StatusBadRequest,
			DisplayMessage: "Invalid request",
		},
	}

	w := testutils.PerformRequest(h.router, http.MethodPost, "/api/v0/modules/1/operations/op-list-topics", strings.NewReader(requestJSON))

	assert.Equal(h.T(), http.StatusBadRequest, w.Code, "Status code should be 400")

	var response types.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)

	assert.NoError(h.T(), err, "Should be able to get valid json response")
	assert.Equal(h.T(), expectedResponse.ErrorInfo.DisplayMessage, response.ErrorInfo.DisplayMessage, "The response should match")
	assert.Equal(h.T(), expectedResponse.ErrorInfo.Code, response.ErrorInfo.Code, "The response should match")
}

func (h *HandlerTestSuite) TestListTopicsHandlerClusterIdAbsent() {
	requestJSON := `{"clusterID": ""}`
	expectedResponse := types.ErrorResponse{
		ErrorInfo: types.ErrorContent{
			Message:        "",
			Code:           "INVALID_REQUEST_BODY",
			HTTPCode:       http.StatusBadRequest,
			DisplayMessage: "Invalid request",
		},
	}

	w := testutils.PerformRequest(h.router, http.MethodPost, "/api/v0/modules/1/operations/op-list-topics", strings.NewReader(requestJSON))

	assert.Equal(h.T(), http.StatusBadRequest, w.Code, "Status code should be 400")

	var response types.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)

	assert.NoError(h.T(), err, "Should be able to get valid json response")
	assert.Equal(h.T(), expectedResponse.ErrorInfo.DisplayMessage, response.ErrorInfo.DisplayMessage, "The response should match")
	assert.Equal(h.T(), expectedResponse.ErrorInfo.Code, response.ErrorInfo.Code, "The response should match")
}

func (h *HandlerTestSuite) TestListTopicsHandlerOnErrorFromService() {
	requestJSON := `{"clusterID": "1"}`
	request := kafkatypes.ListTopicsRequest{ClusterID: "1"}
	expectedResponse := types.ErrorResponse{
		ErrorInfo: types.ErrorContent{
			Message:        modules.INTERNAL_SERVER_ERROR,
			Code:           "INTERNAL_ERROR",
			HTTPCode:       http.StatusInternalServerError,
			DisplayMessage: "An internal error occurred",
		},
	}

	h.mockService.EXPECT().ListTopics(&request).Return(&kafkatypes.ListTopicsResponse{}, types.InternalServerError("internal server error"))

	w := testutils.PerformRequest(h.router, "POST", "/api/v0/modules/1/operations/op-list-topics", strings.NewReader(requestJSON))

	assert.Equal(h.T(), http.StatusInternalServerError, w.Code, "Status code should be 500")

	var response types.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)

	assert.NoError(h.T(), err, "Should be able to get valid json response")
	assert.Equal(h.T(), expectedResponse.ErrorInfo.DisplayMessage, response.ErrorInfo.DisplayMessage, "The response should match")
	assert.Equal(h.T(), expectedResponse.ErrorInfo.Code, response.ErrorInfo.Code, "The response should match")
}
func (h *HandlerTestSuite) TestListTopicsHandlerSuccessResponse() {
	requestJSON := `{"clusterID": "1"}`
	request := kafkatypes.ListTopicsRequest{ClusterID: "1"}
	expectedResponse := kafkatypes.ListTopicsResponse{
		Topics: []kafkatypes.TopicInfo{
			{
				Name:              "cities",
				Partitions:        3,
				ReplicationFactor: 1,
				RetentionMS:       "604800000",
			},
		},
	}

	h.mockService.EXPECT().ListTopics(&request).Return(&expectedResponse, nil)
	w := testutils.PerformRequest(h.router, "POST", "/api/v0/modules/1/operations/op-list-topics", strings.NewReader(requestJSON))

	assert.Equal(h.T(), http.StatusOK, w.Code, "Status code should be 200")

	var response kafkatypes.ListTopicsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)

	assert.NoError(h.T(), err, "Should be able to get valid json response")
	assert.Equal(h.T(), expectedResponse, response, "The response should match")
}

func (h *HandlerTestSuite) TestListTopicsHandlerWhenClusterIdIsInvalid() {
	requestJSON := `{"clusterID": "invalid-cluster-id"}`
	request := kafkatypes.ListTopicsRequest{ClusterID: "invalid-cluster-id"}

	// Service returns an error when cluster ID is invalid
	serviceError := types.NewNotFoundError("clusterId invalid-cluster-id not found", modules.CLUSTER_URL_NOT_FOUND, "ClusterId could not be found")

	expectedResponse := types.ErrorResponse{
		ErrorInfo: types.ErrorContent{
			Message:        "clusterId invalid-cluster-id not found",
			Code:           modules.CLUSTER_URL_NOT_FOUND,
			HTTPCode:       http.StatusNotFound,
			DisplayMessage: "ClusterId could not be found",
		},
	}

	h.mockService.EXPECT().ListTopics(&request).Return(&kafkatypes.ListTopicsResponse{}, serviceError)

	w := testutils.PerformRequest(h.router, "POST", "/api/v0/modules/1/operations/op-list-topics", strings.NewReader(requestJSON))

	assert.Equal(h.T(), http.StatusNotFound, w.Code, "Status code should be 404")

	var response types.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)

	assert.NoError(h.T(), err, "Should be able to get valid json response")
	assert.Equal(h.T(), expectedResponse.ErrorInfo.DisplayMessage, response.ErrorInfo.DisplayMessage)
	assert.Equal(h.T(), expectedResponse.ErrorInfo.Code, response.ErrorInfo.Code)
}

func (h *HandlerTestSuite) TestCreateTopicHandler200Response() {

	requestJSON := `{
		"clusterId":"1",
		"topicName":"abcd",
		"numPartitions":1,
		"replicationFactor":1,
		"retentionMs":10000}`

	reqBody := kafkatypes.CreateTopicRequest{
		ClusterID:         "1",
		TopicName:         "abcd",
		Partitions:        1,
		ReplicationFactor: 1,
		RetentionMs:       10000,
	}
	expectedResp := kafkatypes.CreateTopicResponse{
		Message: "Topic created successfully",
	}

	h.mockValidator.EXPECT().ValidateCreateTopicRequest(gomock.Any()).Return(true, nil)
	h.mockService.EXPECT().CreateTopic(&reqBody).Return(&expectedResp, nil)

	w := testutils.PerformRequest(h.router, "POST", "/api/v0/modules/1/operations/op-create-topic", strings.NewReader(requestJSON))

	var response kafkatypes.CreateTopicResponse

	err := json.Unmarshal(w.Body.Bytes(), &response)

	assert.NoError(h.T(), err)
	assert.Equal(h.T(), http.StatusOK, w.Code, "Status code should be 200")
	assert.Equal(h.T(), response.Message, expectedResp.Message, "Messages should be matched")

}

func (h *HandlerTestSuite) TestCreateTopicHandlerMalformedJson() {

	requestJSON := `{
		"clusterId":"1",
		"topicName":"abcd",
		"numPartitions":1,
		"replicationFactor":1,
		"retentionMs":"10 }`

	expectedResponse := types.ErrorResponse{
		ErrorInfo: types.ErrorContent{
			Message:        "",
			Code:           "INVALID_REQUEST_BODY",
			HTTPCode:       http.StatusBadRequest,
			DisplayMessage: "Invalid request body",
		},
	}

	w := testutils.PerformRequest(h.router, "POST", "/api/v0/modules/1/operations/op-create-topic", strings.NewReader(requestJSON))

	assert.Equal(h.T(), http.StatusBadRequest, w.Code, "Status code should be 400- json bind error")

	var response types.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)

	assert.NoError(h.T(), err, "Should be able to get valid json response")
	assert.Equal(h.T(), expectedResponse.ErrorInfo.DisplayMessage, response.ErrorInfo.DisplayMessage, "The response should match")
	assert.Equal(h.T(), expectedResponse.ErrorInfo.Code, response.ErrorInfo.Code, "The response should match")
}

func (h *HandlerTestSuite) TestCreateTopic500Response() {
	requestJSON := `{
		"clusterId":"1",
		"topicName":"abcd",
		"numPartitions":1,
		"replicationFactor":1,
		"retentionMs":10000}`

	reqBody := kafkatypes.CreateTopicRequest{
		ClusterID:         "1",
		TopicName:         "abcd",
		Partitions:        1,
		ReplicationFactor: 1,
		RetentionMs:       10000,
	}
	expectedResponse := types.ErrorResponse{
		ErrorInfo: types.ErrorContent{
			Message:        "internal server error",
			Code:           modules.INTERNAL_ERROR,
			HTTPCode:       http.StatusInternalServerError,
			DisplayMessage: "An internal error occurred",
		},
	}

	h.mockValidator.EXPECT().ValidateCreateTopicRequest(gomock.Any()).Return(true, nil).Times(1)
	h.mockService.EXPECT().CreateTopic(&reqBody).Return(&kafkatypes.CreateTopicResponse{}, types.InternalServerError("internal server error"))

	w := testutils.PerformRequest(h.router, "POST", "/api/v0/modules/1/operations/op-create-topic", strings.NewReader(requestJSON))

	assert.Equal(h.T(), http.StatusInternalServerError, w.Code, "Status code should be 500")

	var response types.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)

	assert.NoError(h.T(), err, "Should be able to get valid json response")
	assert.Equal(h.T(), expectedResponse.ErrorInfo.DisplayMessage, response.ErrorInfo.DisplayMessage, "The response should match")
	assert.Equal(h.T(), expectedResponse.ErrorInfo.Code, response.ErrorInfo.Code, "The response should match")

}

func (h *HandlerTestSuite) TestCreateTopicHandlerMissingField() {
	requestJSON := `{
		"clusterId": "1",
		"numPartitions": 1,
		"replicationFactor": 1,
		"retentionMs": "10000"
	}`

	expectedResponse := types.ErrorResponse{
		ErrorInfo: types.ErrorContent{
			Message:        "",
			Code:           modules.INVALID_REQUEST_BODY,
			HTTPCode:       http.StatusBadRequest,
			DisplayMessage: "Invalid request body",
		},
	}

	w := testutils.PerformRequest(h.router, "POST", "/api/v0/modules/1/operations/op-create-topic", strings.NewReader(requestJSON))

	assert.Equal(h.T(), http.StatusBadRequest, w.Code, "Should return 400 - json bind error")

	var response types.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)

	assert.NoError(h.T(), err, "Should be able to get valid json response")
	assert.Equal(h.T(), expectedResponse.ErrorInfo.DisplayMessage, response.ErrorInfo.DisplayMessage, "The response should match")
	assert.Equal(h.T(), expectedResponse.ErrorInfo.Code, response.ErrorInfo.Code, "The response should match")
}

func (h *HandlerTestSuite) TestCreateTopicHandlerInvalidType() {
	requestJSON := `{
		"clusterId": "1",
		"topicName": "abcd",
		"numPartitions": "one",
		"replicationFactor": 1,
		"retentionMs": "10000"
	}`

	expectedResponse := types.ErrorResponse{
		ErrorInfo: types.ErrorContent{
			Message:        "",
			Code:           modules.INVALID_REQUEST_BODY,
			HTTPCode:       http.StatusBadRequest,
			DisplayMessage: "Invalid request body",
		},
	}

	w := testutils.PerformRequest(h.router, "POST", "/api/v0/modules/1/operations/op-create-topic", strings.NewReader(requestJSON))

	assert.Equal(h.T(), http.StatusBadRequest, w.Code, "Should return 400 - json bind error")

	var response types.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)

	assert.NoError(h.T(), err, "Should be able to get valid json response")
	assert.Equal(h.T(), expectedResponse.ErrorInfo.DisplayMessage, response.ErrorInfo.DisplayMessage, "The response should match")
	assert.Equal(h.T(), expectedResponse.ErrorInfo.Code, response.ErrorInfo.Code, "The response should match")
}

func (h *HandlerTestSuite) TestCreateTopicHandlerEmptyBody() {
	requestJSON := ``

	expectedResponse := types.ErrorResponse{
		ErrorInfo: types.ErrorContent{
			Message:        "",
			Code:           modules.INVALID_REQUEST_BODY,
			HTTPCode:       http.StatusBadRequest,
			DisplayMessage: "Invalid request body",
		},
	}

	w := testutils.PerformRequest(h.router, "POST", "/api/v0/modules/1/operations/op-create-topic", strings.NewReader(requestJSON))

	assert.Equal(h.T(), http.StatusBadRequest, w.Code, "Empty body should return 400- json bind error")

	var response types.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)

	assert.NoError(h.T(), err, "Should be able to get valid json response")
	assert.Equal(h.T(), expectedResponse.ErrorInfo.DisplayMessage, response.ErrorInfo.DisplayMessage, "The response should match")
	assert.Equal(h.T(), expectedResponse.ErrorInfo.Code, response.ErrorInfo.Code, "The response should match")
}

func (h *HandlerTestSuite) TestCreateTopicHandlerValidationFailure() {
	requestJSON := `{
		"clusterId":"1",
		"topicName":"abcd",
		"numPartitions":1,
		"replicationFactor":1,
		"retentionMs":10000}`

	reqBody := kafkatypes.CreateTopicRequest{
		ClusterID:         "1",
		TopicName:         "abcd",
		Partitions:        1,
		ReplicationFactor: 1,
		RetentionMs:       10000,
	}

	// Create a KafkaError as expected by the validator interface
	validationError := &kafkatypes.KafkaError{
		Error: kafkatypes.Error{
			Code:           modules.BAD_REQUEST,
			Message:        "Topic name should follow naming convention",
			DisplayMessage: "Topic name should follow naming convention",
		},
	}

	// Mock validator to return false with a KafkaError
	h.mockValidator.EXPECT().ValidateCreateTopicRequest(&reqBody).Return(false, validationError)

	w := testutils.PerformRequest(h.router, "POST", "/api/v0/modules/1/operations/op-create-topic", strings.NewReader(requestJSON))

	assert.Equal(h.T(), http.StatusBadRequest, w.Code, "Status code should be 400 for validation failure")

	var response kafkatypes.KafkaError
	err := json.Unmarshal(w.Body.Bytes(), &response)

	assert.NoError(h.T(), err, "Should be able to get valid json response")
	assert.Equal(h.T(), validationError.Error.Code, response.Error.Code, "The response code should match")
	assert.Equal(h.T(), validationError.Error.Message, response.Error.Message, "The response message should match")
	assert.Equal(h.T(), validationError.Error.DisplayMessage, response.Error.DisplayMessage, "The response display message should match")
}

//Delete Topic handler tests

func (h *HandlerTestSuite) TestDeleteTopicHandler200Response() {
	requestJSON := `
		{
			"clusterId":"abcd",
			"topicName":"test"
		}
	`

	reqBody := kafkatypes.DeleteTopicRequest{
		ClusterID: "abcd",
		TopicName: "test",
	}

	expectedResponse := kafkatypes.DeleteTopicResponse{
		Message: "Topic deleted successfully",
	}

	h.mockService.EXPECT().DeleteTopic(&reqBody).Return(&expectedResponse, nil)

	w := testutils.PerformRequest(h.router, "POST", "/api/v0/modules/1/operations/op-delete-topic", strings.NewReader(requestJSON))

	var response kafkatypes.DeleteTopicResponse

	err := json.Unmarshal(w.Body.Bytes(), &response)

	assert.NoError(h.T(), err)
	assert.Equal(h.T(), http.StatusOK, w.Code, "Status code should be 200")
	assert.Equal(h.T(), response.Message, expectedResponse.Message, "Messages should be matched")
}

func (h *HandlerTestSuite) TestDeleteTopicMalformedJson() {
	requestJSON := `
		{
			"clusterId":"abcd",
			"topicName":"te
		}
	`

	expectedResponse := types.ErrorResponse{
		ErrorInfo: types.ErrorContent{
			Message:        "",
			Code:           modules.INVALID_REQUEST_BODY,
			HTTPCode:       http.StatusBadRequest,
			DisplayMessage: "Invalid request",
		},
	}

	w := testutils.PerformRequest(h.router, "POST", "/api/v0/modules/1/operations/op-delete-topic", strings.NewReader(requestJSON))

	assert.Equal(h.T(), http.StatusBadRequest, w.Code, "Status code should be 400- json bind error")

	var response types.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)

	assert.NoError(h.T(), err, "Should be able to get valid json response")
	assert.Equal(h.T(), expectedResponse.ErrorInfo.DisplayMessage, response.ErrorInfo.DisplayMessage, "The response should match")
	assert.Equal(h.T(), expectedResponse.ErrorInfo.Code, response.ErrorInfo.Code, "The response should match")
}

func (h *HandlerTestSuite) TestDeleteTopic500Response() {
	requestJSON := `
		{
			"clusterId": "abcd",
			"topicName": "test"
		}
	`

	reqBody := kafkatypes.DeleteTopicRequest{
		ClusterID: "abcd",
		TopicName: "test",
	}

	expectedResponse := types.ErrorResponse{
		ErrorInfo: types.ErrorContent{
			Message:        "internal server error",
			Code:           modules.INTERNAL_ERROR,
			HTTPCode:       http.StatusInternalServerError,
			DisplayMessage: "An internal error occurred",
		},
	}

	h.mockService.EXPECT().
		DeleteTopic(&reqBody).
		Return(&kafkatypes.DeleteTopicResponse{}, types.InternalServerError("internal server error"))

	w := testutils.PerformRequest(h.router, "POST", "/api/v0/modules/1/operations/op-delete-topic", strings.NewReader(requestJSON))

	assert.Equal(h.T(), http.StatusInternalServerError, w.Code, "Status code should be 500 - internal server error")

	var response types.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(h.T(), err, "Should be able to parse valid JSON error response")

	assert.Equal(h.T(), expectedResponse.ErrorInfo.DisplayMessage, response.ErrorInfo.DisplayMessage, "DisplayMessage should match")
	assert.Equal(h.T(), expectedResponse.ErrorInfo.Code, response.ErrorInfo.Code, "Error code should match")
	assert.Equal(h.T(), expectedResponse.ErrorInfo.Message, response.ErrorInfo.Message, "Message should match")
}

func (h *HandlerTestSuite) TestDeleteTopicMissingField() {
	requestJSON := `
		{
			"topicName": "test"
		}
	`

	expectedResponse := types.ErrorResponse{
		ErrorInfo: types.ErrorContent{
			Message:        "",
			Code:           modules.INVALID_REQUEST_BODY,
			HTTPCode:       http.StatusBadRequest,
			DisplayMessage: "Invalid request",
		},
	}

	w := testutils.PerformRequest(h.router, "POST", "/api/v0/modules/1/operations/op-delete-topic", strings.NewReader(requestJSON))

	assert.Equal(h.T(), http.StatusBadRequest, w.Code, "Status code should be 400- missing field")

	var response types.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)

	assert.NoError(h.T(), err, "Should be able to get valid json response")
	assert.Equal(h.T(), expectedResponse.ErrorInfo.DisplayMessage, response.ErrorInfo.DisplayMessage, "The response should match")
	assert.Equal(h.T(), expectedResponse.ErrorInfo.Code, response.ErrorInfo.Code, "The response should match")
}

func (h *HandlerTestSuite) TestDeleteTopicEmptyBody() {
	requestJSON := ``

	expectedResponse := types.ErrorResponse{
		ErrorInfo: types.ErrorContent{
			Message:        "",
			Code:           modules.INVALID_REQUEST_BODY,
			HTTPCode:       http.StatusBadRequest,
			DisplayMessage: "Invalid request",
		},
	}

	w := testutils.PerformRequest(h.router, "POST", "/api/v0/modules/1/operations/op-delete-topic", strings.NewReader(requestJSON))

	assert.Equal(h.T(), http.StatusBadRequest, w.Code, "Status code should be 400- empty body")

	var response types.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)

	assert.NoError(h.T(), err, "Should be able to get valid json response")
	assert.Equal(h.T(), expectedResponse.ErrorInfo.DisplayMessage, response.ErrorInfo.DisplayMessage, "The response should match")
	assert.Equal(h.T(), expectedResponse.ErrorInfo.Code, response.ErrorInfo.Code, "The response should match")
}

func (h *HandlerTestSuite) TestDeleteTopicHandlerInvalidType() {
	requestJSON := `{
			"clusterId":1,
			"topicName":"test"
		}`

	expectedResponse := types.ErrorResponse{
		ErrorInfo: types.ErrorContent{
			Message:        "",
			Code:           modules.INVALID_REQUEST_BODY,
			HTTPCode:       http.StatusBadRequest,
			DisplayMessage: "Invalid request",
		},
	}

	w := testutils.PerformRequest(h.router, "POST", "/api/v0/modules/1/operations/op-delete-topic", strings.NewReader(requestJSON))

	assert.Equal(h.T(), http.StatusBadRequest, w.Code, "Should return 400 - json bind error")

	var response types.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)

	assert.NoError(h.T(), err, "Should be able to get valid json response")
	assert.Equal(h.T(), expectedResponse.ErrorInfo.DisplayMessage, response.ErrorInfo.DisplayMessage, "The response should match")
	assert.Equal(h.T(), expectedResponse.ErrorInfo.Code, response.ErrorInfo.Code, "The response should match")
}
