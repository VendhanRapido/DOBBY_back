package utils

import (
	"fmt"
	"strings"

	"github.com/roppenlabs/dobby-service/internal/config"
)

//go:generate mockgen -destination=clusterUtils_mock.go -package=utils . ClusterUtils

type ClusterUtils interface {
	GetClusterUrlFromClusterId(cid string) (string, error)
	GetTimeoutForGetMetadataFromConfig() int
}

type clusterUtils struct {
	config *config.Config
}

func NewClusterUtils(config *config.Config) ClusterUtils {
	return &clusterUtils{
		config: config,
	}
}

func (c *clusterUtils) GetClusterUrlFromClusterId(cid string) (string, error) {
	clusterInfo := c.config.GetKafkaClusterInfo()

	for key, url := range clusterInfo {
		if strings.EqualFold(key, cid) {
			return url, nil
		}
	}

	return "", fmt.Errorf("%s not found", cid)
}

func (c *clusterUtils) GetTimeoutForGetMetadataFromConfig() int {
	timeoutForGetMetadata := c.config.GetMetadataTimeoutFromConfig()
	return timeoutForGetMetadata
}
