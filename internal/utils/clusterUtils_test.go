package utils

import (
	"testing"

	"github.com/roppenlabs/dobby-service/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ClusterUtilsTestSuite struct {
	suite.Suite
	clusterUtils ClusterUtils
	config       *config.Config
}

func TestClusterUtilsTestSuite(t *testing.T) {
	suite.Run(t, new(ClusterUtilsTestSuite))
}

func (c *ClusterUtilsTestSuite) BeforeTest(suiteName, testName string) {
	c.config = &config.Config{
		Kafka: config.KafkaConfig{
			ClusterInfo: map[string]string{
				"cluster1": "localhost:9092",
			},
			TimeoutInMS: config.KafkaTimeouts{
				GetMetadata: 5000,
			},
		},
	}

	c.clusterUtils = NewClusterUtils(c.config)
}

func (c *ClusterUtilsTestSuite) TestGetClusterUrlFromClusterId_Success() {
	clusterID := "cluster1"
	expectedURL := "localhost:9092"

	url, err := c.clusterUtils.GetClusterUrlFromClusterId(clusterID)

	assert.NoError(c.T(), err, "Should not return error for valid cluster ID")
	assert.Equal(c.T(), expectedURL, url, "Should return correct cluster URL")
}

func (c *ClusterUtilsTestSuite) TestGetClusterUrlFromClusterId_NotFound() {
	clusterID := "nonexistent-cluster"

	url, err := c.clusterUtils.GetClusterUrlFromClusterId(clusterID)

	assert.Error(c.T(), err, "Should return error for invalid cluster ID")
	assert.Empty(c.T(), url, "Should return empty URL for invalid cluster ID")
	assert.Contains(c.T(), err.Error(), clusterID, "Error message should contain the cluster ID")
}

func (c *ClusterUtilsTestSuite) TestGetTimeoutForGetMetadataFromConfig() {
	expectedTimeout := 5000

	timeout := c.clusterUtils.GetTimeoutForGetMetadataFromConfig()

	assert.Equal(c.T(), expectedTimeout, timeout, "Should return correct timeout value from config")
}
