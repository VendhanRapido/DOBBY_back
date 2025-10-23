package kafka

import (
	"github.com/google/wire"
	kafkahelpers "github.com/roppenlabs/dobby-service/internal/modules/kafka/helpers"
	kafkaservices "github.com/roppenlabs/dobby-service/internal/modules/kafka/services"
	"github.com/roppenlabs/dobby-service/internal/modules/kafka/validators"
)

var WireSet = wire.NewSet(
	kafkahelpers.NewKafkaClient,
	NewHandler,
	kafkaservices.NewService,
	validators.NewValidator,
)
