package kafkaservices

import (
	"context"
	"net/http"
	"strconv"
	"time"

	confluentKafka "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/roppenlabs/dobby-service/internal/modules"
	kafkatypes "github.com/roppenlabs/dobby-service/internal/modules/kafka/types"
	"github.com/roppenlabs/dobby-service/internal/types"
)

func (s *serviceImpl) CreateTopic(req *kafkatypes.CreateTopicRequest) (*kafkatypes.CreateTopicResponse, *types.ErrorResponse) {
	clusterUrl, err := s.clusterUtils.GetClusterUrlFromClusterId(req.ClusterID)
	if err != nil {
		return nil, types.NewNotFoundError(
			err.Error(),
			modules.CLUSTER_URL_NOT_FOUND,
			"ClusterId could not be found",
		)
	}

	adminClient, err := s.kafkaHelper.CreateClient(clusterUrl)
	if err != nil {
		return nil, types.InternalServerError("Failed to create Kafka client: " + err.Error())
	}
	defer adminClient.Close()

	topicSpec := confluentKafka.TopicSpecification{
		Topic:             req.TopicName,
		NumPartitions:     req.Partitions,
		ReplicationFactor: req.ReplicationFactor,
		Config: map[string]string{
			"retention.ms": strconv.Itoa(req.RetentionMs),
		},
	}

	timeoutSeconds := s.config.GetCreateTopicTimeout()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	results, err := adminClient.CreateTopics(ctx, []confluentKafka.TopicSpecification{topicSpec})
	if err != nil {
		return nil, types.InternalServerError(err.Error())
	}

	if len(results) == 0 {
		return nil, types.InternalServerError(
			"Something went wrong, Please check if the operation has succeeded",
		)
	}

	result := results[0]

	if result.Error.Code() != confluentKafka.ErrNoError {
		switch result.Error.Code() {
		case confluentKafka.ErrInvalidConfig:
			return nil, types.BadRequestError(
				"Invalid Kafka configuration received",
				modules.INVALID_KAFKA_CONFIG,
				"Invalid Kafka config",
			)
		case confluentKafka.ErrTopicAlreadyExists:
			return nil, types.CustomError(
				"Kafka topic "+topicSpec.Topic+" already exists",
				modules.KAFKA_TOPIC_EXISTS,
				http.StatusConflict,
				"Kafka topic "+topicSpec.Topic+" already exists",
			)
		default:
			return nil, types.InternalServerError(
				"Kafka topic creation failed",
			)
		}
	}

	return &kafkatypes.CreateTopicResponse{
		Message: "Topic created successfully",
	}, nil
}
