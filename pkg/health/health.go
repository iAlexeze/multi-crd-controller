package health

import (
	"context"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/ialexeze/kubernetes-crd-example/pkg/config/domain"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/logger"
)

var _ domain.Component = (*HealthServer)(nil)

type HealthServer struct {
	server *http.Server
	ready  atomic.Bool
	port   string
	client string
}

func NewHealthServer(client, port string) *HealthServer {
	if client == "" {
		client = "service"
	}

	hs := &HealthServer{
		client: client,
		port:   port,
	}

	// server is not ready on startup. modified when client is ready to process requests
	hs.ready.Store(false)
	return hs
}

func (h *HealthServer) Start(ctx context.Context) error {
	if !strings.HasPrefix(h.port, ":") {
		logger.Debug().Msgf("normalizing port from %s to :%s", h.port, h.port)
		h.port = ":" + h.port
	}

	h.server = &http.Server{
		Addr:    h.port,
		Handler: h.routes(),
	}

	go func() {
		if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error().Err(err).Msg("health server error")
		}
	}()

	return nil
}

func (h *HealthServer) Shutdown(ctx context.Context) {
	if h.server != nil {
		if err := h.server.Shutdown(ctx); err != nil {
			logger.Error().Err(err).Msg("health server shutdown error")
		}
	}
	h.ready.Store(false)
}

func (h *HealthServer) Name() string {
	return "health server"
}

func (h *HealthServer) SetReady() {
	h.ready.Store(true)
}

func (h *HealthServer) routes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("/health", h.logRouteMiddleware(http.HandlerFunc(h.healthHandler)))
	mux.Handle("/ready", h.logRouteMiddleware(http.HandlerFunc(h.readyHandler)))

	return mux
}
