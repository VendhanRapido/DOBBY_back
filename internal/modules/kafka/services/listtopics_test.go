package kafkaservices

import (
	"fmt"
	"testing"

	confluentKafka "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/roppenlabs/dobby-service/internal/config"
	kafkahelpers "github.com/roppenlabs/dobby-service/internal/modules/kafka/helpers"
	kafkatypes "github.com/roppenlabs/dobby-service/internal/modules/kafka/types"
	"github.com/roppenlabs/dobby-service/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type ListTopicsTestSuite struct {
	suite.Suite
	ctrl             *gomock.Controller
	mockKafkaHelper  *kafkahelpers.MockKafkaHelper
	mockAdminClient  *kafkahelpers.MockAdminClient
	mockClusterUtils *utils.MockClusterUtils
	mockConfig       *config.Config
	service          Service
}

func TestListTopicsTestSuite(t *testing.T) {
	suite.Run(t, new(ListTopicsTestSuite))
}

func (c *ListTopicsTestSuite) BeforeTest(suiteName, testName string) {
	c.ctrl = gomock.NewController(c.T())
	c.mockKafkaHelper = kafkahelpers.NewMockKafkaHelper(c.ctrl)
	c.mockAdminClient = kafkahelpers.NewMockAdminClient(c.ctrl)
	c.mockClusterUtils = utils.NewMockClusterUtils(c.ctrl)

	c.mockConfig = &config.Config{
		Kafka: config.KafkaConfig{
			TimeoutInMS: config.KafkaTimeouts{
				GetMetadata: 5000,
			},
		},
	}

	c.service = NewService(c.mockKafkaHelper, c.mockClusterUtils, c.mockConfig)
}

func (c *ListTopicsTestSuite) AfterTest(suiteName, testName string) {
	c.ctrl.Finish()
}

func (c *ListTopicsTestSuite) TestListTopicsWhenClusterIdIsInvalid() {
	req := &kafkatypes.ListTopicsRequest{
		ClusterID: "invalid-cluster-id",
	}

	baseErr := fmt.Errorf("clusterId %s not found", req.ClusterID)

	c.mockClusterUtils.EXPECT().GetClusterUrlFromClusterId(req.ClusterID).Return("", baseErr).Times(1)

	resp, errResp := c.service.ListTopics(req)

	assert.Nil(c.T(), resp, "Response should be nil when cluster is not found")
	assert.NotNil(c.T(), errResp, "ErrorResponse should not be nil when cluster is not found")

	assert.Equal(c.T(), baseErr.Error(), errResp.ErrorInfo.Message)
	assert.Equal(c.T(), "CLUSTER_URL_NOT_FOUND", errResp.ErrorInfo.Code)
}

func (c *ListTopicsTestSuite) TestListTopicsWhenKafkaClientCreationFails() {
	req := &kafkatypes.ListTopicsRequest{
		ClusterID: "test-cluster-id",
	}
	baseErr := fmt.Errorf("unable to connect to Kafka cluster")
	clusterURL := "localhost:9092"

	c.mockClusterUtils.EXPECT().
		GetClusterUrlFromClusterId(req.ClusterID).
		Return(clusterURL, nil).
		Times(1)

	c.mockKafkaHelper.EXPECT().
		CreateClient(clusterURL).
		Return(nil, baseErr).
		Times(1)

	resp, errResp := c.service.ListTopics(req)

	assert.Nil(c.T(), resp, "Response should be nil when kafka creation fails")
	assert.NotNil(c.T(), errResp, "ErrorResponse should not be nil when kafka creation fails")

	assert.Equal(c.T(), baseErr.Error(), errResp.ErrorInfo.Message)
	assert.Equal(c.T(), "INTERNAL_ERROR", errResp.ErrorInfo.Code)
}

func (c *ListTopicsTestSuite) TestListTopicsWhenKafkaIsUnavailable() {
	req := &kafkatypes.ListTopicsRequest{
		ClusterID: "test-cluster-id",
	}
	baseErr := fmt.Errorf("kafka not available")
	clusterURL := "localhost:9092"
	timeout := 5000

	c.mockClusterUtils.EXPECT().
		GetClusterUrlFromClusterId(req.ClusterID).
		Return(clusterURL, nil).
		Times(1)

	c.mockKafkaHelper.EXPECT().
		CreateClient(clusterURL).
		Return(c.mockAdminClient, nil).
		Times(1)

	c.mockClusterUtils.EXPECT().
		GetTimeoutForGetMetadataFromConfig().
		Return(timeout).
		Times(1)

	c.mockAdminClient.EXPECT().
		GetMetadata(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, baseErr).
		Times(1)

	c.mockAdminClient.EXPECT().
		Close().
		Times(1)

	resp, errResp := c.service.ListTopics(req)

	assert.Nil(c.T(), resp, "Response should be nil when get metadata fails")
	assert.NotNil(c.T(), errResp, "ErrorResponse should not be nil when get metadata fails")

	assert.Equal(c.T(), baseErr.Error(), errResp.ErrorInfo.Message)
	assert.Equal(c.T(), "INTERNAL_ERROR", errResp.ErrorInfo.Code)
}

func (c *ListTopicsTestSuite) TestListTopicsWhenRetentionMSisNil() {
	req := &kafkatypes.ListTopicsRequest{
		ClusterID: "test-cluster-id",
	}

	clusterURL := "localhost:9092"
	timeout := 5000

	mockMetadata := &confluentKafka.Metadata{
		Brokers: []confluentKafka.BrokerMetadata{
			{ID: 1, Host: "localhost", Port: 9092},
		},
		Topics: map[string]confluentKafka.TopicMetadata{
			"test-topic": {
				Partitions: []confluentKafka.PartitionMetadata{
					{
						Replicas: []int32{1, 2},
					},
				},
			},
		},
		OriginatingBroker: confluentKafka.BrokerMetadata{
			ID:   1,
			Host: "localhost",
			Port: 9092,
		},
	}

	c.mockClusterUtils.EXPECT().
		GetClusterUrlFromClusterId(req.ClusterID).
		Return(clusterURL, nil).
		Times(1)

	c.mockKafkaHelper.EXPECT().
		CreateClient(clusterURL).
		Return(c.mockAdminClient, nil).
		Times(1)

	c.mockClusterUtils.EXPECT().
		GetTimeoutForGetMetadataFromConfig().
		Return(timeout).
		Times(1)

	c.mockAdminClient.EXPECT().
		GetMetadata(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(mockMetadata, nil).
		Times(1)

	c.mockAdminClient.EXPECT().
		GetRetentionMs("test-topic").
		Return("", fmt.Errorf("retentionMs not found in topic config")).
		Times(1)

	c.mockAdminClient.EXPECT().
		Close().
		Times(1)

	resp, errResp := c.service.ListTopics(req)

	assert.NotNil(c.T(), resp, "Response should not be nil when operation succeeds")
	assert.Nil(c.T(), errResp, "ErrorResponse should be nil when operation succeeds")

	assert.Equal(c.T(), "test-topic", resp.Topics[0].Name)
	assert.Equal(c.T(), "", resp.Topics[0].RetentionMS, "RetentionMS should be empty when not found")
}

func (c *ListTopicsTestSuite) TestListTopicsNoTopics() {
	req := &kafkatypes.ListTopicsRequest{
		ClusterID: "test-cluster-id",
	}

	clusterURL := "localhost:9092"
	timeout := 5000

	mockMetadata := &confluentKafka.Metadata{
		Brokers:           []confluentKafka.BrokerMetadata{},
		Topics:            map[string]confluentKafka.TopicMetadata{},
		OriginatingBroker: confluentKafka.BrokerMetadata{},
	}

	c.mockClusterUtils.EXPECT().
		GetClusterUrlFromClusterId(req.ClusterID).
		Return(clusterURL, nil).
		Times(1)

	c.mockKafkaHelper.EXPECT().
		CreateClient(clusterURL).
		Return(c.mockAdminClient, nil).
		Times(1)

	c.mockClusterUtils.EXPECT().
		GetTimeoutForGetMetadataFromConfig().
		Return(timeout).
		Times(1)

	c.mockAdminClient.EXPECT().
		GetMetadata(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(mockMetadata, nil).
		Times(1)

	c.mockAdminClient.EXPECT().
		Close().
		Times(1)

	resp, _ := c.service.ListTopics(req)

	assert.Len(c.T(), resp.Topics, 0, "Should have no topic")

}

func (c *ListTopicsTestSuite) TestListTopicsSuccess() {
	req := &kafkatypes.ListTopicsRequest{
		ClusterID: "test-cluster-id",
	}

	clusterURL := "localhost:9092"
	timeout := 5000

	mockMetadata := &confluentKafka.Metadata{
		Brokers: []confluentKafka.BrokerMetadata{
			{ID: 1, Host: "localhost", Port: 9092},
		},
		Topics: map[string]confluentKafka.TopicMetadata{
			"test-topic": {
				Partitions: []confluentKafka.PartitionMetadata{
					{
						Replicas: []int32{1, 2},
					},
				},
			},
		},
		OriginatingBroker: confluentKafka.BrokerMetadata{
			ID:   1,
			Host: "localhost",
			Port: 9092,
		},
	}

	c.mockClusterUtils.EXPECT().
		GetClusterUrlFromClusterId(req.ClusterID).
		Return(clusterURL, nil).
		Times(1)

	c.mockKafkaHelper.EXPECT().
		CreateClient(clusterURL).
		Return(c.mockAdminClient, nil).
		Times(1)

	c.mockClusterUtils.EXPECT().
		GetTimeoutForGetMetadataFromConfig().
		Return(timeout).
		Times(1)

	c.mockAdminClient.EXPECT().
		GetMetadata(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(mockMetadata, nil).
		Times(1)

	c.mockAdminClient.EXPECT().
		GetRetentionMs("test-topic").
		Return("604800000", nil).
		Times(1)

	c.mockAdminClient.EXPECT().
		Close().
		Times(1)

	resp, _ := c.service.ListTopics(req)

	assert.NotNil(c.T(), resp, "Response should not be nil when operation succeeds")

	assert.Len(c.T(), resp.Topics, 1, "Should have one topic")
	assert.Equal(c.T(), "test-topic", resp.Topics[0].Name)
	assert.Equal(c.T(), "604800000", resp.Topics[0].RetentionMS)
	assert.Equal(c.T(), 1, resp.Topics[0].Partitions)
	assert.Equal(c.T(), 2, resp.Topics[0].ReplicationFactor)
}
