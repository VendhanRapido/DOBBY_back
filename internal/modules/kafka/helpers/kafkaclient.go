package kafkahelpers

import (
	confluentKafka "github.com/confluentinc/confluent-kafka-go/kafka"
)

//go:generate mockgen -destination=kafkaclient_mock.go -package=kafkahelpers . KafkaClient

type KafkaClient interface {
	CreateClient(clusterUrl string) (AdminClient, error)
}

type kafkaClient struct {
	KafkaClient
}

func NewKafkaClient() KafkaClient {
	return &kafkaClient{}
}

func (kc *kafkaClient) CreateClient(clusterUrl string) (AdminClient, error) {
	adminClient, err := confluentKafka.NewAdminClient(&confluentKafka.ConfigMap{"bootstrap.servers": clusterUrl})
	if err != nil {
		return nil, err
	}
	return &AdminClientWrapper{client: adminClient}, nil
}
