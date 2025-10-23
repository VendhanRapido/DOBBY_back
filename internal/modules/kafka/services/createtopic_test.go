package kafkaservices

import (
	"fmt"
	"net/http"
	"testing"

	confluentKafka "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/roppenlabs/dobby-service/internal/modules"
	kafkahelpers "github.com/roppenlabs/dobby-service/internal/modules/kafka/helpers"
	kafkatypes "github.com/roppenlabs/dobby-service/internal/modules/kafka/types"

	"github.com/roppenlabs/dobby-service/internal/config"
	"github.com/roppenlabs/dobby-service/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type CreateTopicTestSuite struct {
	suite.Suite
	ctrl             *gomock.Controller
	mockKafkaHelper  *kafkahelpers.MockKafkaHelper
	mockAdminClient  *kafkahelpers.MockAdminClient
	mockClusterUtils *utils.MockClusterUtils
	mockConfig       *config.Config
	service          Service
}

func TestCreateTopicTestSuite(t *testing.T) {
	suite.Run(t, new(CreateTopicTestSuite))
}

func (c *CreateTopicTestSuite) BeforeTest(suiteName, testName string) {

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
				CreateTopic: 30000,
			},
		},
	}

	c.service = NewService(c.mockKafkaHelper, c.mockClusterUtils, c.mockConfig)
}

func (c *CreateTopicTestSuite) AfterTest(suiteName, testName string) {
	c.ctrl.Finish()
}

func (c *CreateTopicTestSuite) TestCreateTopicSuccess() {

	req := &kafkatypes.CreateTopicRequest{
		ClusterID:         "test_cluster",
		TopicName:         "test_topic",
		Partitions:        3,
		ReplicationFactor: 2,
		RetentionMs:       86400000,
	}

	c.mockClusterUtils.EXPECT().
		GetClusterUrlFromClusterId(req.ClusterID).
		Return("localhost:9092", nil).
		Times(1)

	c.mockKafkaHelper.EXPECT().
		CreateClient("localhost:9092").
		Return(c.mockAdminClient, nil).
		Times(1)

	successResult := []confluentKafka.TopicResult{
		{
			Topic: req.TopicName,
			Error: confluentKafka.Error{},
		},
	}

	c.mockAdminClient.EXPECT().
		CreateTopics(gomock.Any(), gomock.Any()).
		Return(successResult, nil).
		Times(1)

	c.mockAdminClient.EXPECT().
		Close().
		Times(1)

	resp, errResp := c.service.CreateTopic(req)

	assert.NotNil(c.T(), resp, "Response should not be nil when topic is created successfully")
	assert.Nil(c.T(), errResp, "Error response should be nil when successful")
	assert.Equal(c.T(), "Topic created successfully", resp.Message)
}

func (c *CreateTopicTestSuite) TestCreateTopicWhenClusterIdIsInvalid() {
	req := &kafkatypes.CreateTopicRequest{
		ClusterID:         "invalid-cluster-id",
		TopicName:         "test-topic",
		Partitions:        1,
		ReplicationFactor: 2,
		RetentionMs:       36000,
	}

	baseErr := fmt.Errorf("clusterId %s not found", req.ClusterID)

	c.mockClusterUtils.EXPECT().
		GetClusterUrlFromClusterId(req.ClusterID).
		Return("", baseErr).
		Times(1)

	resp, errResp := c.service.CreateTopic(req)

	assert.Nil(c.T(), resp, "Response should be nil when cluster is not found")
	assert.NotNil(c.T(), errResp, "Error response should not be nil")

	assert.Equal(c.T(), baseErr.Error(), errResp.ErrorInfo.Message)
	assert.Equal(c.T(), modules.CLUSTER_URL_NOT_FOUND, errResp.ErrorInfo.Code)
	assert.Equal(c.T(), http.StatusNotFound, errResp.ErrorInfo.HTTPCode)
}

func (c *CreateTopicTestSuite) TestCreateTopicWhenKafkaClientCreationFails() {
	req := &kafkatypes.CreateTopicRequest{
		ClusterID:         "test-cluster-id",
		TopicName:         "test-topic",
		Partitions:        1,
		ReplicationFactor: 2,
		RetentionMs:       36000,
	}

	clusterURL := "localhost:9092"
	baseErr := fmt.Errorf("failed to create confluentKafka admin client: unable to connect to Kafka cluster")

	c.mockClusterUtils.EXPECT().
		GetClusterUrlFromClusterId(req.ClusterID).
		Return(clusterURL, nil).
		Times(1)

	c.mockKafkaHelper.EXPECT().
		CreateClient(clusterURL).
		Return(nil, baseErr).
		Times(1)

	resp, errResp := c.service.CreateTopic(req)

	assert.Nil(c.T(), resp, "Response should be nil when client creation fails")
	assert.NotNil(c.T(), errResp, "Error response should not be nil")

	assert.Equal(c.T(), modules.INTERNAL_ERROR, errResp.ErrorInfo.Code)
	assert.Equal(c.T(), http.StatusInternalServerError, errResp.ErrorInfo.HTTPCode)
}
func (c *CreateTopicTestSuite) TestCreateTopicWhenTopicAlreadyExists() {
	req := &kafkatypes.CreateTopicRequest{
		ClusterID:         "test-cluster-id",
		TopicName:         "existing-topic",
		Partitions:        1,
		ReplicationFactor: 2,
		RetentionMs:       36000,
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

	failureResult := []confluentKafka.TopicResult{
		{
			Error: confluentKafka.NewError(confluentKafka.ErrTopicAlreadyExists, "Topic already exists", false),
		},
	}

	c.mockAdminClient.EXPECT().
		CreateTopics(gomock.Any(), gomock.Any()).
		Return(failureResult, nil).
		Times(1)

	c.mockAdminClient.EXPECT().
		Close().
		Times(1)

	resp, errResp := c.service.CreateTopic(req)

	assert.Nil(c.T(), resp, "Response should be nil when topic already exists")
	assert.NotNil(c.T(), errResp, "Error response should not be nil")

	assert.Equal(c.T(), "Kafka topic existing-topic already exists", errResp.ErrorInfo.Message)
	assert.Equal(c.T(), modules.KAFKA_TOPIC_EXISTS, errResp.ErrorInfo.Code)
	assert.Equal(c.T(), http.StatusConflict, errResp.ErrorInfo.HTTPCode)
}

func (c *CreateTopicTestSuite) TestCreateTopicWhenInvalidConfig() {
	req := &kafkatypes.CreateTopicRequest{
		ClusterID:         "test-cluster-id",
		TopicName:         "test-topic",
		Partitions:        1,
		ReplicationFactor: 2,
		RetentionMs:       36000,
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

	failureResult := []confluentKafka.TopicResult{
		{
			Topic: req.TopicName,
			Error: confluentKafka.NewError(confluentKafka.ErrInvalidConfig, "Invalid topic configuration", false),
		},
	}

	c.mockAdminClient.EXPECT().
		CreateTopics(gomock.Any(), gomock.Any()).
		Return(failureResult, nil).
		Times(1)

	c.mockAdminClient.EXPECT().
		Close().
		Times(1)

	resp, errResp := c.service.CreateTopic(req)

	assert.Nil(c.T(), resp, "Response should be nil when topic config is invalid")
	assert.NotNil(c.T(), errResp, "Error response should not be nil")

	assert.Equal(c.T(), "Invalid Kafka configuration received", errResp.ErrorInfo.Message)
	assert.Equal(c.T(), modules.INVALID_KAFKA_CONFIG, errResp.ErrorInfo.Code)
	assert.Equal(c.T(), http.StatusBadRequest, errResp.ErrorInfo.HTTPCode)
}

func (c *CreateTopicTestSuite) TestCreateTopicWhenKafkaUnavailable() {
	req := &kafkatypes.CreateTopicRequest{
		ClusterID:         "test-cluster-id",
		TopicName:         "test-topic",
		Partitions:        1,
		ReplicationFactor: 2,
		RetentionMs:       36000,
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

	failureResult := []confluentKafka.TopicResult{
		{
			Topic: req.TopicName,
			Error: confluentKafka.NewError(confluentKafka.ErrAllBrokersDown, "All broker connections are down", false),
		},
	}

	c.mockAdminClient.EXPECT().
		CreateTopics(gomock.Any(), gomock.Any()).
		Return(failureResult, nil).
		Times(1)

	c.mockAdminClient.EXPECT().
		Close().
		Times(1)

	resp, errResp := c.service.CreateTopic(req)

	assert.Nil(c.T(), resp, "Response should be nil when Kafka is unavailable")
	assert.NotNil(c.T(), errResp, "Error response should not be nil")

	assert.Equal(c.T(), "Kafka topic creation failed", errResp.ErrorInfo.Message)
	assert.Equal(c.T(), modules.INTERNAL_ERROR, errResp.ErrorInfo.Code)
	assert.Equal(c.T(), http.StatusInternalServerError, errResp.ErrorInfo.HTTPCode)
}

func (c *CreateTopicTestSuite) TestCreateTopicWhenGetTimeoutFails() {
	originalConfig := config.AppConfig
	defer func() {
		config.AppConfig = originalConfig
	}()

	config.AppConfig = config.Config{
		Kafka: config.KafkaConfig{
			TimeoutInMS: config.KafkaTimeouts{
				CreateTopic: 0,
				DeleteTopic: 30000,
				GetMetadata: 5000,
			},
		},
	}

	req := &kafkatypes.CreateTopicRequest{
		ClusterID:         "test-cluster-id",
		TopicName:         "test-topic",
		Partitions:        1,
		ReplicationFactor: 2,
		RetentionMs:       36000,
	}

	clusterURL := "localhost:9092"
	createTopicsErr := fmt.Errorf("Invalid CreateTopic timeout configuration: must be greater than 0")

	c.mockClusterUtils.EXPECT().
		GetClusterUrlFromClusterId(req.ClusterID).
		Return(clusterURL, nil).
		Times(1)

	c.mockKafkaHelper.EXPECT().
		CreateClient(clusterURL).
		Return(c.mockAdminClient, nil).
		Times(1)

	c.mockAdminClient.EXPECT().
		CreateTopics(gomock.Any(), gomock.Any()).
		Return(nil, createTopicsErr).
		Times(1)

	c.mockAdminClient.EXPECT().
		Close().
		Times(1)

	resp, errResp := c.service.CreateTopic(req)

	assert.Nil(c.T(), resp, "Response should be nil when config timeout fails")
	assert.NotNil(c.T(), errResp, "Error response should not be nil")

	assert.Equal(c.T(), modules.INTERNAL_ERROR, errResp.ErrorInfo.Code)
	assert.Equal(c.T(), http.StatusInternalServerError, errResp.ErrorInfo.HTTPCode)
	assert.Equal(c.T(), "Invalid CreateTopic timeout configuration: must be greater than 0", errResp.ErrorInfo.Message)
}

func (c *CreateTopicTestSuite) TestCreateTopicWhenCreateTopicsReturnsError() {
	req := &kafkatypes.CreateTopicRequest{
		ClusterID:         "test-cluster-id",
		TopicName:         "test-topic",
		Partitions:        1,
		ReplicationFactor: 2,
		RetentionMs:       36000,
	}

	clusterURL := "localhost:9092"
	createTopicsErr := fmt.Errorf("failed to create topics: connection lost")

	c.mockClusterUtils.EXPECT().
		GetClusterUrlFromClusterId(req.ClusterID).
		Return(clusterURL, nil).
		Times(1)

	c.mockKafkaHelper.EXPECT().
		CreateClient(clusterURL).
		Return(c.mockAdminClient, nil).
		Times(1)

	c.mockAdminClient.EXPECT().
		CreateTopics(gomock.Any(), gomock.Any()).
		Return(nil, createTopicsErr).
		Times(1)

	c.mockAdminClient.EXPECT().
		Close().
		Times(1)

	resp, errResp := c.service.CreateTopic(req)

	assert.Nil(c.T(), resp, "Response should be nil when CreateTopics returns error")
	assert.NotNil(c.T(), errResp, "Error response should not be nil")

	assert.Equal(c.T(), modules.INTERNAL_ERROR, errResp.ErrorInfo.Code)
	assert.Equal(c.T(), http.StatusInternalServerError, errResp.ErrorInfo.HTTPCode)
	assert.Equal(c.T(), createTopicsErr.Error(), errResp.ErrorInfo.Message)
}

func (c *CreateTopicTestSuite) TestCreateTopicWhenNoResultsReturned() {
	req := &kafkatypes.CreateTopicRequest{
		ClusterID:         "test-cluster-id",
		TopicName:         "test-topic",
		Partitions:        1,
		ReplicationFactor: 2,
		RetentionMs:       36000,
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
		CreateTopics(gomock.Any(), gomock.Any()).
		Return([]confluentKafka.TopicResult{}, nil).
		Times(1)

	c.mockAdminClient.EXPECT().
		Close().
		Times(1)

	resp, errResp := c.service.CreateTopic(req)

	assert.Nil(c.T(), resp, "Response should be nil when no results returned")
	assert.NotNil(c.T(), errResp, "Error response should not be nil when no results")
	assert.Equal(c.T(), "Something went wrong, Please check if the operation has succeeded", errResp.ErrorInfo.Message)
	assert.Equal(c.T(), modules.INTERNAL_ERROR, errResp.ErrorInfo.Code)
	assert.Equal(c.T(), http.StatusInternalServerError, errResp.ErrorInfo.HTTPCode)
}
