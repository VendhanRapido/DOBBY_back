//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/roppenlabs/dobby-service/internal/config"
	"github.com/roppenlabs/dobby-service/internal/health"
	"github.com/roppenlabs/dobby-service/internal/helloworld"
	"github.com/roppenlabs/dobby-service/internal/server"
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
		helloworld.WireSet,
		health.WireSet,
		config.GetConfig,
	)

	return ServerDependencies{}, nil
}
