package kafkaservices

import (
	"context"

	"net/http"
	"time"

	confluentKafka "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/roppenlabs/dobby-service/internal/modules"
	kafkatypes "github.com/roppenlabs/dobby-service/internal/modules/kafka/types"
	"github.com/roppenlabs/dobby-service/internal/types"
)

func (s *serviceImpl) DeleteTopic(req *kafkatypes.DeleteTopicRequest) (*kafkatypes.DeleteTopicResponse, *types.ErrorResponse) {
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

	topic := req.TopicName
	timeoutSeconds := s.config.GetDeleteTopicTimeout()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	results, err := adminClient.DeleteTopic(ctx, topic)
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
		case confluentKafka.ErrUnknownTopicOrPart:
			return nil, types.NewNotFoundError(
				"Topic "+req.TopicName+" does not exist!",
				modules.TOPIC_NOT_FOUND,
				"Kafka topic doesn't exist",
			)

		default:
			return nil, types.CustomError(
				"Kafka topic deletion error: "+result.Error.String(),
				modules.KAFKA_TOPIC_DELETION_ERROR,
				http.StatusInternalServerError,
				"Kafka topic deletion failed",
			)
		}
	}

	return &kafkatypes.DeleteTopicResponse{
		Message: "Topic deleted successfully",
	}, nil
}
