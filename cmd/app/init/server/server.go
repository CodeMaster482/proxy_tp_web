package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	handler "proxy/internal/api/handler/http"
	"proxy/pkg/config"
	"proxy/pkg/logger"

	"github.com/gin-gonic/gin"
)

type Server struct {
	*http.Server
	Logger logger.Logger
}

func NewServer(mainCtx context.Context, cfg *config.Config, r *gin.Engine, log logger.Logger, h *handler.Handler) *Server {
	// Set Up Middleware
	r.Use(gin.Recovery())

	// Set up /api router group
	api := r.Group("/api")

	api.GET("/requests", h.GetRequests)
	api.GET("/requests/:id", h.GetRequestById)

	api.GET("/repeat/:id", h.RepeatRequest)
	api.GET("/scan/:id", h.ScanRequest)

	s := &Server{
		Server: &http.Server{
			Addr:    fmt.Sprintf("%s:%s", cfg.Server.Addr, cfg.Server.Port),
			Handler: r,
			BaseContext: func(_ net.Listener) context.Context {
				return mainCtx
			},
		},
		Logger: log,
	}
	return s
}

func (s *Server) ApiRun() {
	s.Logger.Info("Router initialized")
	if err := s.ListenAndServe(); err != nil {
		s.Logger.Fatal("Server error: ", err)
	}
	s.Logger.Info("Server down")
}
