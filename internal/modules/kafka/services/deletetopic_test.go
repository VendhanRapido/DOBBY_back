package kafkaservices

import (
	"context"
	"errors"
	"fmt"
	"testing"

	confluentKafka "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/roppenlabs/dobby-service/internal/config"
	"github.com/roppenlabs/dobby-service/internal/modules"
	kafkahelpers "github.com/roppenlabs/dobby-service/internal/modules/kafka/helpers"
	kafkatypes "github.com/roppenlabs/dobby-service/internal/modules/kafka/types"
	"github.com/roppenlabs/dobby-service/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type DeleteTopicTestSuite struct {
	suite.Suite
	ctrl             *gomock.Controller
	mockKafkaHelper  *kafkahelpers.MockKafkaHelper
	mockAdminClient  *kafkahelpers.MockAdminClient
	mockClusterUtils *utils.MockClusterUtils
	mockConfig       *config.Config
	service          Service
}

func TestDeleteTopicTestSuite(t *testing.T) {
	suite.Run(t, new(DeleteTopicTestSuite))
}

func (c *DeleteTopicTestSuite) BeforeTest(suiteName, testName string) {
	config.AppConfig = config.Config{
		Kafka: config.KafkaConfig{
			TimeoutInMS: config.KafkaTimeouts{
				CreateTopic: 30000,
				DeleteTopic: 30000,
				GetMetadata: 5000,
			},
		},
	}

	c.ctrl = gomock.NewController(c.T())
	c.mockKafkaHelper = kafkahelpers.NewMockKafkaHelper(c.ctrl)
	c.mockAdminClient = kafkahelpers.NewMockAdminClient(c.ctrl)
	c.mockClusterUtils = utils.NewMockClusterUtils(c.ctrl)

	c.mockConfig = &config.Config{
		Kafka: config.KafkaConfig{
			TimeoutInMS: config.KafkaTimeouts{
				DeleteTopic: 30000,
			},
		},
	}

	c.service = NewService(c.mockKafkaHelper, c.mockClusterUtils, c.mockConfig)
}

func (c *DeleteTopicTestSuite) AfterTest(suiteName, testName string) {
	c.ctrl.Finish()
}

func (c *DeleteTopicTestSuite) TestDeleteTopicWhenClusterIdIsInvalid() {
	req := &kafkatypes.DeleteTopicRequest{
		ClusterID: "invalid-cluster-id",
		TopicName: "test-topic",
	}

	baseErr := fmt.Errorf("clusterId %s not found", req.ClusterID)

	c.mockClusterUtils.EXPECT().
		GetClusterUrlFromClusterId(req.ClusterID).
		Return("", baseErr).
		Times(1)

	resp, errResp := c.service.DeleteTopic(req)

	assert.Nil(c.T(), resp, "Response should be nil when cluster is not found")
	assert.NotNil(c.T(), errResp, "Error response should not be nil")
	assert.Equal(c.T(), baseErr.Error(), errResp.ErrorInfo.Message)
	assert.Equal(c.T(), modules.CLUSTER_URL_NOT_FOUND, errResp.ErrorInfo.Code)
}

func (c *DeleteTopicTestSuite) TestDeleteTopicWhenKafkaClientCreationFails() {
	req := &kafkatypes.DeleteTopicRequest{
		ClusterID: "test-cluster-id",
		TopicName: "test-topic",
	}

	clusterURL := "localhost:9092"
	createClientErr := fmt.Errorf("unable to connect to Kafka cluster")

	c.mockClusterUtils.EXPECT().
		GetClusterUrlFromClusterId(req.ClusterID).
		Return(clusterURL, nil).
		Times(1)

	c.mockKafkaHelper.EXPECT().
		CreateClient(clusterURL).
		Return(nil, createClientErr).
		Times(1)

	resp, errResp := c.service.DeleteTopic(req)

	assert.Nil(c.T(), resp, "Response should be nil when client creation fails")
	assert.NotNil(c.T(), errResp, "Error response should not be nil")
	assert.Equal(c.T(), "Failed to create Kafka client: "+createClientErr.Error(), errResp.ErrorInfo.Message)
	assert.Equal(c.T(), modules.INTERNAL_ERROR, errResp.ErrorInfo.Code)
}

func (c *DeleteTopicTestSuite) TestDeleteTopicWhenContextTimesOut() {
	req := &kafkatypes.DeleteTopicRequest{
		ClusterID: "test-cluster-id",
		TopicName: "test-topic",
	}

	clusterURL := "localhost:9092"
	timeoutErr := context.DeadlineExceeded

	c.mockClusterUtils.EXPECT().
		GetClusterUrlFromClusterId(req.ClusterID).
		Return(clusterURL, nil).
		Times(1)

	c.mockKafkaHelper.EXPECT().
		CreateClient(clusterURL).
		Return(c.mockAdminClient, nil).
		Times(1)

	c.mockAdminClient.EXPECT().
		DeleteTopic(gomock.Any(), gomock.Any()).
		Return(nil, timeoutErr).
		Times(1)

	c.mockAdminClient.EXPECT().
		Close().
		Times(1)

	resp, errResp := c.service.DeleteTopic(req)

	assert.Nil(c.T(), resp, "Response should be nil when context times out")
	assert.NotNil(c.T(), errResp, "Error response should not be nil")
	assert.Equal(c.T(), timeoutErr.Error(), errResp.ErrorInfo.Message)
	assert.Equal(c.T(), modules.INTERNAL_ERROR, errResp.ErrorInfo.Code)
}

func (c *DeleteTopicTestSuite) TestDeleteTopicWhenTopicDoesntExist() {
	req := &kafkatypes.DeleteTopicRequest{
		ClusterID: "test-cluster-id",
		TopicName: "non-existent-topic",
	}

	clusterURL := "localhost:9092"
	deleteErr := fmt.Errorf("topic not found")

	c.mockClusterUtils.EXPECT().
		GetClusterUrlFromClusterId(req.ClusterID).
		Return(clusterURL, nil).
		Times(1)

	c.mockKafkaHelper.EXPECT().
		CreateClient(clusterURL).
		Return(c.mockAdminClient, nil).
		Times(1)

	c.mockAdminClient.EXPECT().
		Close().
		Times(1)

	c.mockAdminClient.EXPECT().
		DeleteTopic(gomock.Any(), gomock.Any()).
		Return(nil, deleteErr).
		Times(1)

	resp, errResp := c.service.DeleteTopic(req)

	assert.Nil(c.T(), resp, "Response should be nil when topic doesn't exist")
	assert.NotNil(c.T(), errResp, "Error response should not be nil")
	assert.Equal(c.T(), deleteErr.Error(), errResp.ErrorInfo.Message)
	assert.Equal(c.T(), modules.INTERNAL_ERROR, errResp.ErrorInfo.Code)
}

func (c *DeleteTopicTestSuite) TestDeleteTopicWhenKafkaUnavailable() {
	req := &kafkatypes.DeleteTopicRequest{
		ClusterID: "test-cluster-id",
		TopicName: "test-topic",
	}

	clusterURL := "localhost:9092"
	deleteErr := errors.New("Failed to reach kafka")

	c.mockClusterUtils.EXPECT().
		GetClusterUrlFromClusterId(req.ClusterID).
		Return(clusterURL, nil).
		Times(1)

	c.mockKafkaHelper.EXPECT().
		CreateClient(clusterURL).
		Return(c.mockAdminClient, nil).
		Times(1)

	c.mockAdminClient.EXPECT().
		DeleteTopic(gomock.Any(), gomock.Any()).
		Return(nil, deleteErr).
		Times(1)

	c.mockAdminClient.EXPECT().
		Close().
		Times(1)

	resp, errResp := c.service.DeleteTopic(req)

	assert.Nil(c.T(), resp, "Response should be nil when Kafka is unavailable")
	assert.NotNil(c.T(), errResp, "Error response should not be nil")
	assert.Equal(c.T(), deleteErr.Error(), errResp.ErrorInfo.Message)
	assert.Equal(c.T(), modules.INTERNAL_ERROR, errResp.ErrorInfo.Code)
}

func (c *DeleteTopicTestSuite) TestDeleteTopicWhenKafkaReturnsError() {
	req := &kafkatypes.DeleteTopicRequest{
		ClusterID: "test-cluster-id",
		TopicName: "test-topic",
	}

	clusterURL := "localhost:9092"

	mockTopicResult := &confluentKafka.TopicResult{
		Topic: req.TopicName,
		Error: confluentKafka.NewError(confluentKafka.ErrTopicException, "Topic deletion failed", false),
	}
	results := []confluentKafka.TopicResult{*mockTopicResult}

	c.mockClusterUtils.EXPECT().
		GetClusterUrlFromClusterId(req.ClusterID).
		Return(clusterURL, nil).
		Times(1)

	c.mockKafkaHelper.EXPECT().
		CreateClient(clusterURL).
		Return(c.mockAdminClient, nil).
		Times(1)

	c.mockAdminClient.EXPECT().
		DeleteTopic(gomock.Any(), gomock.Any()).
		Return(results, nil).
		Times(1)

	c.mockAdminClient.EXPECT().
		Close().
		Times(1)

	resp, errResp := c.service.DeleteTopic(req)

	assert.Nil(c.T(), resp, "Response should be nil when Kafka returns error")
	assert.NotNil(c.T(), errResp, "Error response should not be nil")
	assert.Equal(c.T(), modules.KAFKA_TOPIC_DELETION_ERROR, errResp.ErrorInfo.Code)
	assert.Contains(c.T(), errResp.ErrorInfo.Message, "Kafka topic deletion error:")
}

func (c *DeleteTopicTestSuite) TestDeleteTopicWhenTopicExists() {
	req := &kafkatypes.DeleteTopicRequest{
		ClusterID: "test-cluster-id",
		TopicName: "existing-topic",
	}

	clusterURL := "localhost:9092"
	mockTopicResult := &confluentKafka.TopicResult{
		Topic: req.TopicName,
		Error: confluentKafka.NewError(confluentKafka.ErrNoError, "", false),
	}
	results := []confluentKafka.TopicResult{*mockTopicResult}

	c.mockClusterUtils.EXPECT().
		GetClusterUrlFromClusterId(req.ClusterID).
		Return(clusterURL, nil).
		Times(1)

	c.mockKafkaHelper.EXPECT().
		CreateClient(clusterURL).
		Return(c.mockAdminClient, nil).
		Times(1)

	c.mockAdminClient.EXPECT().
		DeleteTopic(gomock.Any(), gomock.Any()).
		Return(results, nil).
		Times(1)

	c.mockAdminClient.EXPECT().
		Close().
		Times(1)

	resp, errResp := c.service.DeleteTopic(req)

	assert.NotNil(c.T(), resp, "Response should not be nil when topic is deleted successfully")
	assert.Nil(c.T(), errResp, "Error response should be nil when successful")
	assert.Equal(c.T(), "Topic deleted successfully", resp.Message)
}

func (c *DeleteTopicTestSuite) TestDeleteTopicWhenTopicNotFound() {
	req := &kafkatypes.DeleteTopicRequest{
		ClusterID: "test-cluster-id",
		TopicName: "non-existent-topic",
	}

	clusterURL := "localhost:9092"

	mockTopicResult := &confluentKafka.TopicResult{
		Topic: req.TopicName,
		Error: confluentKafka.NewError(confluentKafka.ErrUnknownTopicOrPart, "Unknown topic or partition", false),
	}
	results := []confluentKafka.TopicResult{*mockTopicResult}

	c.mockClusterUtils.EXPECT().
		GetClusterUrlFromClusterId(req.ClusterID).
		Return(clusterURL, nil).
		Times(1)

	c.mockKafkaHelper.EXPECT().
		CreateClient(clusterURL).
		Return(c.mockAdminClient, nil).
		Times(1)

	c.mockAdminClient.EXPECT().
		DeleteTopic(gomock.Any(), gomock.Any()).
		Return(results, nil).
		Times(1)

	c.mockAdminClient.EXPECT().
		Close().
		Times(1)

	_, errResp := c.service.DeleteTopic(req)

	assert.Equal(c.T(), modules.TOPIC_NOT_FOUND, errResp.ErrorInfo.Code)
	assert.Equal(c.T(), "Topic "+req.TopicName+" does not exist!", errResp.ErrorInfo.Message)
}

func (c *DeleteTopicTestSuite) TestDeleteTopicWhenNotAuthorized() {
	req := &kafkatypes.DeleteTopicRequest{
		ClusterID: "test-cluster-id",
		TopicName: "protected-topic",
	}

	clusterURL := "localhost:9092"

	mockTopicResult := &confluentKafka.TopicResult{
		Topic: req.TopicName,
		Error: confluentKafka.NewError(confluentKafka.ErrTopicAuthorizationFailed, "Topic authorization failed", false),
	}
	results := []confluentKafka.TopicResult{*mockTopicResult}

	c.mockClusterUtils.EXPECT().
		GetClusterUrlFromClusterId(req.ClusterID).
		Return(clusterURL, nil).
		Times(1)

	c.mockKafkaHelper.EXPECT().
		CreateClient(clusterURL).
		Return(c.mockAdminClient, nil).
		Times(1)

	c.mockAdminClient.EXPECT().
		DeleteTopic(gomock.Any(), gomock.Any()).
		Return(results, nil).
		Times(1)

	c.mockAdminClient.EXPECT().
		Close().
		Times(1)

	_, errResp := c.service.DeleteTopic(req)
	assert.Equal(c.T(), modules.KAFKA_TOPIC_DELETION_ERROR, errResp.ErrorInfo.Code)
	assert.Contains(c.T(), errResp.ErrorInfo.Message, "Kafka topic deletion error:")
}

func (c *DeleteTopicTestSuite) TestDeleteTopicSuccess() {
	req := &kafkatypes.DeleteTopicRequest{
		ClusterID: "test-cluster-id",
		TopicName: "existing-topic",
	}

	clusterURL := "localhost:9092"

	mockTopicResult := &confluentKafka.TopicResult{
		Topic: req.TopicName,
		Error: confluentKafka.NewError(confluentKafka.ErrNoError, "", false),
	}
	results := []confluentKafka.TopicResult{*mockTopicResult}

	c.mockClusterUtils.EXPECT().
		GetClusterUrlFromClusterId(req.ClusterID).
		Return(clusterURL, nil).
		Times(1)

	c.mockKafkaHelper.EXPECT().
		CreateClient(clusterURL).
		Return(c.mockAdminClient, nil).
		Times(1)

	c.mockAdminClient.EXPECT().
		DeleteTopic(gomock.Any(), gomock.Any()).
		Return(results, nil).
		Times(1)

	c.mockAdminClient.EXPECT().
		Close().
		Times(1)

	resp, err := c.service.DeleteTopic(req)

	assert.Nil(c.T(), err, "Error response should be nil when successful")
	assert.Equal(c.T(), "Topic deleted successfully", resp.Message)
}

func (c *DeleteTopicTestSuite) TestDeleteTopicWhenGetTimeoutFails() {
	originalConfig := config.AppConfig
	defer func() {
		config.AppConfig = originalConfig
	}()

	config.AppConfig = config.Config{
		Kafka: config.KafkaConfig{
			TimeoutInMS: config.KafkaTimeouts{
				CreateTopic: 30000,
				DeleteTopic: 0,
				GetMetadata: 5000,
			},
		},
	}

	req := &kafkatypes.DeleteTopicRequest{
		ClusterID: "test-cluster-id",
		TopicName: "test-topic",
	}

	clusterURL := "localhost:9092"
	deleteTopicError := fmt.Errorf("Invalid DeleteTopic timeout configuration: must be greater than 0")

	c.mockClusterUtils.EXPECT().
		GetClusterUrlFromClusterId(req.ClusterID).
		Return(clusterURL, nil).
		Times(1)

	c.mockKafkaHelper.EXPECT().
		CreateClient(clusterURL).
		Return(c.mockAdminClient, nil).
		Times(1)

	c.mockAdminClient.EXPECT().
		DeleteTopic(gomock.Any(), gomock.Any()).
		Return(nil, deleteTopicError).
		Times(1)

	c.mockAdminClient.EXPECT().
		Close().
		Times(1)

	resp, errResp := c.service.DeleteTopic(req)

	assert.Nil(c.T(), resp, "Response should be nil when config timeout fails")
	assert.NotNil(c.T(), errResp, "Error response should not be nil")

	assert.Equal(c.T(), modules.INTERNAL_ERROR, errResp.ErrorInfo.Code)
	assert.Equal(c.T(), "Invalid DeleteTopic timeout configuration: must be greater than 0", errResp.ErrorInfo.Message)
}

func (c *DeleteTopicTestSuite) TestDeleteTopicWhenNoResultsReturned() {
	req := &kafkatypes.DeleteTopicRequest{
		ClusterID: "test-cluster-id",
		TopicName: "test-topic",
	}

	clusterURL := "localhost:9092"

	c.mockClusterUtils.EXPECT().
		GetClusterUrlFromClusterId(req.ClusterID).
		Return(clusterURL, nil).
		Times(1)

	c.mockKafkaHelper.EXPECT().
		CreateClient(clusterURL).
		Return(c.mockAdminClient, nil).
		Times(1)

	// Return empty results array
	c.mockAdminClient.EXPECT().
		DeleteTopic(gomock.Any(), gomock.Any()).
		Return([]confluentKafka.TopicResult{}, nil).
		Times(1)

	c.mockAdminClient.EXPECT().
		Close().
		Times(1)

	resp, errResp := c.service.DeleteTopic(req)

	assert.Nil(c.T(), resp, "Response should be nil when no results returned")
	assert.NotNil(c.T(), errResp, "Error response should not be nil when no results")
	assert.Equal(c.T(), "Something went wrong, Please check if the operation has succeeded", errResp.ErrorInfo.Message)
	assert.Equal(c.T(), modules.INTERNAL_ERROR, errResp.ErrorInfo.Code)
}
