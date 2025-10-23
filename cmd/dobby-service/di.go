//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/roppenlabs/dobby-service/internal/accesscontrol"
	"github.com/roppenlabs/dobby-service/internal/clients"
	"github.com/roppenlabs/dobby-service/internal/config"
	"github.com/roppenlabs/dobby-service/internal/health"
	"github.com/roppenlabs/dobby-service/internal/modules"
	"github.com/roppenlabs/dobby-service/internal/modules/kafka"
	"github.com/roppenlabs/dobby-service/internal/server"
	"github.com/roppenlabs/dobby-service/internal/utils"
	"github.com/roppenlabs/dobby-service/internal/utils/filereader"
)

type ServerDependencies struct {
	config   *config.Config
	server   *server.Server
	handlers server.Handlers
}

func InitDependencies() (ServerDependencies, error) {
	wire.Build(
		wire.Struct(new(ServerDependencies), "*"),
		wire.Struct(new(server.Handlers), "*"),
		server.WireSet,
		health.WireSet,
		modules.WireSet,
		clients.WireSet,
		config.GetConfig,
		kafka.WireSet,
		utils.WireSet,
		accesscontrol.WireSet,
		filereader.WireSet,
	)

	return ServerDependencies{}, nil
}
