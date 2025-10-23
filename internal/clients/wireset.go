package clients

import (
	"github.com/google/wire"
	entitiesclient "github.com/roppenlabs/dobby-service/internal/clients/entities"
)

var WireSet = wire.NewSet(
	entitiesclient.NewEntitiesClient,
)
