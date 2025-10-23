package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/roppenlabs/dobby-service/internal/config"
)

type Server struct {
	config       *config.Config
	engine       *gin.Engine
	routerGroups RouterGroups
	middleware   Middleware
}

type RouterGroups struct {
	rootRouter *gin.Engine
}

func NewServer(c *config.Config, m Middleware) *Server {
	if c.IsProductionEnv() {
		log.Println("Setting gin server to release mode for production environment")
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
			log.Printf("%s", fmt.Sprintf("Endpoint %s is declared via handler %s, method: %v", absolutePath, handlerName, httpMethod))
		}
	}
	engine := gin.New()
	engine.Use(gin.Recovery())

	return &Server{
		config:     c,
		engine:     engine,
		middleware: m,
		routerGroups: RouterGroups{
			rootRouter: engine,
		},
	}
}

func (s *Server) Run(h Handlers) {
	s.InitRoutes(h, s.config)
	srv := &http.Server{
		Addr:    s.config.ListenAddress(),
		Handler: s.engine,
	}
	go listenServer(srv)
	waitForShutdown(srv)
}

func listenServer(server *http.Server) {
	err := server.ListenAndServe()
	if err != http.ErrServerClosed {
		panic(err)
	}
}

func waitForShutdown(server *http.Server) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig,
		syscall.SIGINT,
		syscall.SIGTERM)
	<-sig
	log.Println("server shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := server.Shutdown(ctx)
	if err != nil {
		log.Printf("%s", fmt.Sprintf("[error] Server forced to shutdown: %v", err))
	}
	os.RemoveAll("staticfiles")
	log.Println("server shutdown complete")
}
