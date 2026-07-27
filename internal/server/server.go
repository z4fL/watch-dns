package server

import (
	"net/http"

	"github.com/z4fL/watch-dns/internal/config"
)

type Server struct {
	httpServer *http.Server
}

func New(cfg config.AppConfig, handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:    ":" + cfg.Port,
			Handler: handler,
		},
	}
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}
