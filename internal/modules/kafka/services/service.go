package kafkaservices

import (
	"github.com/roppenlabs/dobby-service/internal/config"
	kafkahelpers "github.com/roppenlabs/dobby-service/internal/modules/kafka/helpers"
	kafkatypes "github.com/roppenlabs/dobby-service/internal/modules/kafka/types"
	"github.com/roppenlabs/dobby-service/internal/types"
	"github.com/roppenlabs/dobby-service/internal/utils"
)

//go:generate mockgen -destination=service_mock.go -package=kafkaservices . Service

type serviceImpl struct {
	kafkaHelper  kafkahelpers.KafkaClient
	clusterUtils utils.ClusterUtils
	config       *config.Config
}

type Service interface {
	CreateTopic(req *kafkatypes.CreateTopicRequest) (*kafkatypes.CreateTopicResponse, *types.ErrorResponse)
	ListTopics(*kafkatypes.ListTopicsRequest) (*kafkatypes.ListTopicsResponse, *types.ErrorResponse)
	DeleteTopic(*kafkatypes.DeleteTopicRequest) (*kafkatypes.DeleteTopicResponse, *types.ErrorResponse)
}

func NewService(kafkaHelper kafkahelpers.KafkaClient, clusterUtils utils.ClusterUtils, config *config.Config) Service {
	return &serviceImpl{
		kafkaHelper:  kafkaHelper,
		clusterUtils: clusterUtils,
		config:       config,
	}
}
