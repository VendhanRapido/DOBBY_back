package kafkaservices

import (
	"fmt"

	"github.com/roppenlabs/dobby-service/internal/modules"
	kafkatypes "github.com/roppenlabs/dobby-service/internal/modules/kafka/types"
	"github.com/roppenlabs/dobby-service/internal/types"
)

func (s *serviceImpl) ListTopics(req *kafkatypes.ListTopicsRequest) (*kafkatypes.ListTopicsResponse, *types.ErrorResponse) {

	clientUrl, err := s.clusterUtils.GetClusterUrlFromClusterId(req.ClusterID)
	if err != nil {
		return nil, types.NewNotFoundError(
			err.Error(),
			modules.CLUSTER_URL_NOT_FOUND,
			"ClusterId could not be found",
		)
	}
	adminClient, err := s.kafkaHelper.CreateClient(clientUrl)

	if err != nil {
		return nil, types.InternalServerError(err.Error())
	}
	defer adminClient.Close()

	metadata, err := adminClient.GetMetadata(nil, true, s.clusterUtils.GetTimeoutForGetMetadataFromConfig())
	if err != nil {
		return nil, types.InternalServerError(err.Error())
	}

	var topics []kafkatypes.TopicInfo

	for topic, topicMetadata := range metadata.Topics {
		retentionMs, err := adminClient.GetRetentionMs(topic)
		if err != nil {
			fmt.Printf("Error while fetching retentionMs for %s: %s", topic, err)
		}
		topics = append(topics, kafkatypes.TopicInfo{
			Name:              topic,
			Partitions:        len(topicMetadata.Partitions),
			ReplicationFactor: len(topicMetadata.Partitions[0].Replicas),
			RetentionMS:       retentionMs,
		})
	}
	if topics == nil {
		return &kafkatypes.ListTopicsResponse{
			Topics: []kafkatypes.TopicInfo{},
		}, nil
	}
	return &kafkatypes.ListTopicsResponse{
		Topics: topics,
	}, nil
}
