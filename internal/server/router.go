package server

import (
	"github.com/gin-contrib/pprof"
	"github.com/roppenlabs/dobby-service/internal/config"
	"github.com/roppenlabs/dobby-service/internal/health"
	"github.com/roppenlabs/dobby-service/internal/helloworld"
	"log"
)

type Handlers struct {
	HealthHandler     *health.Handler
	HelloWorldHandler *helloworld.Handler
}

func (s *Server) InitRoutes(h Handlers, c *config.Config) {
	router := s.routerGroups.rootRouter
	router.GET("/sanity", h.HealthHandler.CheckSanity)
	router.GET("/health", h.HealthHandler.CheckHealth)

	// Register pprof handlers
	if c.ProfilingEnabled {
		log.Println("ALERT! Profiling enabled. Please be aware of the performance impact it could have")
		pprof.Register(router)
	}

	h.HelloWorldHandler.InitRoutes(router)
}
