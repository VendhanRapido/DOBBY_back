package modules

import (
	"github.com/google/wire"
)

var WireSet = wire.NewSet(
	NewHandler,
	NewService,
)
