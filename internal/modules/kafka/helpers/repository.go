package kafkahelpers

import (
	"context"
	"fmt"
	"time"

	confluentKafka "github.com/confluentinc/confluent-kafka-go/kafka"
)

//go:generate mockgen -destination=repository_mock.go -package=kafkahelpers . AdminClient

const RetentionMS = "retention.ms"

type AdminClient interface {
	CreateTopics(ctx context.Context, topics []confluentKafka.TopicSpecification) ([]confluentKafka.TopicResult, error)
	GetMetadata(topic *string, allTopics bool, timeoutMs int) (*confluentKafka.Metadata, error)
	DeleteTopic(ctx context.Context, topic string) ([]confluentKafka.TopicResult, error)
	GetRetentionMs(topic string) (string, error)
	Close()
}

type AdminClientWrapper struct {
	client *confluentKafka.AdminClient
}

func (w *AdminClientWrapper) CreateTopics(ctx context.Context, topics []confluentKafka.TopicSpecification) ([]confluentKafka.TopicResult, error) {
	results, err := w.client.CreateTopics(ctx, topics)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (w *AdminClientWrapper) GetMetadata(topic *string, allTopics bool, timeoutMs int) (*confluentKafka.Metadata, error) {
	metadata, err := w.client.GetMetadata(topic, allTopics, timeoutMs)
	if err != nil {
		return nil, err
	}
	return metadata, nil
}

func (w *AdminClientWrapper) DeleteTopic(ctx context.Context, topic string) ([]confluentKafka.TopicResult, error) {
	topics := []string{topic}
	results, err := w.client.DeleteTopics(ctx, topics)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (w *AdminClientWrapper) GetRetentionMs(topic string) (string, error) {
	if topic == "" {
		return "", fmt.Errorf("topic cannot be empty")
	}
	resources := []confluentKafka.ConfigResource{
		{
			Type: confluentKafka.ResourceTopic,
			Name: topic,
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	results, err := w.client.DescribeConfigs(ctx, resources)
	if err != nil {
		return "", err
	}

	for _, result := range results {
		if result.Error.Code() != confluentKafka.ErrNoError {
			return "", result.Error
		}
		for str, entry := range result.Config {
			if str == RetentionMS {
				return entry.Value, nil
			}
		}
	}
	return "", fmt.Errorf("retentionMs not found in topic config")
}

func (w *AdminClientWrapper) Close() {
	w.client.Close()
}
