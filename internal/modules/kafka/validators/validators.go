package validators

import (
	"regexp"

	config "github.com/roppenlabs/dobby-service/internal/config"
	types "github.com/roppenlabs/dobby-service/internal/modules/kafka/types"
)

//go:generate mockgen -destination validator_mock.go -package=validators . Validator
type Validator interface {
	ValidateCreateTopicRequest(req *types.CreateTopicRequest) (bool, *types.KafkaError)
}

type ValidatorImpl struct {
	config *config.Config
}

func NewValidator(c *config.Config) Validator {
	return &ValidatorImpl{
		config: c,
	}
}

func createTopicError(code string, message string) *types.KafkaError {
	return &types.KafkaError{
		Error: types.Error{
			Code:           code,
			Message:        message,
			DisplayMessage: message,
		},
	}
}

func (v *ValidatorImpl) ValidateCreateTopicRequest(req *types.CreateTopicRequest) (bool, *types.KafkaError) {
	conf := v.config.GetKafkaValidationKeys()

	topicName := req.TopicName
	if topicName == "" {
		return false, createTopicError("BAD_REQUEST", topicNameCannotBeEmtpy)
	}
	match, err := regexp.MatchString(conf.TopicNameConventionPattern, topicName)
	if err != nil {
		return false, createTopicError("INTERNAL_SERVER_ERROR", somethingWentWrong)
	} else if !match {
		return false, createTopicError("BAD_REQUEST", topicNameShouldFollowNamingConvention)
	}

	numPartitions := req.Partitions
	if numPartitions > conf.MaxNumPartitions {
		return false, createTopicError("BAD_REQUEST", numPartitionsTooHigh)
	} else if numPartitions < conf.MinNumPartitions {
		return false, createTopicError("BAD_REQUEST", numPartitionsTooLow)
	}

	replicationFactor := req.ReplicationFactor
	if replicationFactor > conf.MaxReplicationFactor {
		return false, createTopicError("BAD_REQUEST", replicationFactorTooHigh)
	} else if replicationFactor < conf.MinReplicationFactor {
		return false, createTopicError("BAD_REQUEST", replicationFactorTooLow)
	}

	retentionInMs := req.RetentionMs
	if retentionInMs > conf.MaxRetentionInMS {
		return false, createTopicError("BAD_REQUEST", retentionPeriodTooHigh)
	} else if retentionInMs < conf.MinRetentionInMS {
		return false, createTopicError("BAD_REQUEST", retentionPeriodTooLow)
	}

	return true, nil
}
