package server

import (
	"log"

	"github.com/gin-contrib/pprof"
	"github.com/roppenlabs/dobby-service/internal/config"
	"github.com/roppenlabs/dobby-service/internal/health"
	"github.com/roppenlabs/dobby-service/internal/modules"
	kafka "github.com/roppenlabs/dobby-service/internal/modules/kafka"
)

type Handlers struct {
	HealthHandler  *health.Handler
	KafkaGroup     *kafka.Handler
	ModulesHandler *modules.Handler
}

func (s *Server) InitRoutes(h Handlers, c *config.Config) {
	router := s.routerGroups.rootRouter
	router.GET("/sanity", h.HealthHandler.CheckSanity)
	router.GET("/health", h.HealthHandler.CheckHealth)
	router.GET("/api/v0/modules/:moduleId/operations", s.middleware.Authenticate(), s.middleware.OperationListing(), h.ModulesHandler.GetOperationsHandler)
	router.GET("/api/v0/modules", s.middleware.Authenticate(), s.middleware.ModuleListing(), h.ModulesHandler.GetModulesHandler)
	router.GET("/api/v0/modules/:moduleId/operations/:operationId", s.middleware.Authenticate(), s.middleware.EnforceRouteAuthorization(), h.ModulesHandler.GetOperationDetailHandler)
	router.POST("/api/v0/modules/:moduleId/operations/:operationId", s.middleware.Authenticate(), s.middleware.EnforceRouteAuthorization(), h.KafkaGroup.OperationsHandler)

	// Register pprof handlers
	if c.ProfilingEnabled {
		log.Println("ALERT! Profiling enabled. Please be aware of the performance impact it could have")
		pprof.Register(router)
	}

}
