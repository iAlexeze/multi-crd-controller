package manager

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ialexeze/kubernetes-crd-example/pkg/config/domain"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/health"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/logger"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/utils"
)

type Manager struct {
	components []domain.Component
	timeout    time.Duration
	done       chan struct{}
	hs         *health.HealthServer
}

func NewManager(hs *health.HealthServer, timeout time.Duration) *Manager {
	return &Manager{
		timeout: timeout,
		hs:      hs,
		done:    make(chan struct{}),
	}
}

func (m *Manager) Start(ctx context.Context) error {
	mCtx, mCancel := context.WithCancel(ctx)
	defer mCancel()

	for _, comp := range m.components {
		name := comp.Name()

		logger.Info().Msgf("starting: %s...", name)
		if err := comp.Start(mCtx); err != nil {
			logger.Error().Err(err).Msgf("failed to start: %s", name)
			return err
		}
		logger.Info().Msgf("%s status: %v", name, utils.StatusOnline)
	}

	logger.Info().Msg("✅ All services started successfully")

	m.setReady()
	logger.Info().Msg("controller is ready...")

	m.gracefulShutdown(mCtx, mCancel)
	return nil
}

func (m *Manager) Shutdown(ctx context.Context) {}

func (m *Manager) gracefulShutdown(ctx context.Context, cancel context.CancelFunc) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		logger.Info().Msgf("recieved shutdown signal: %v", sig)
		cancel()

		// shutdown components
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), m.timeout)
		defer shutdownCancel()

		for _, comp := range m.components {
			name := comp.Name()
			logger.Info().Msgf("shutting down: %s...", name)
			if comp != nil {
				comp.Shutdown(shutdownCtx)
			}
			logger.Info().Msgf("%s status: %v", name, utils.StatusOffline)
		}

		logger.Info().Msg("⚠️  All services shut down gracefully")

		// Notify Wait() to terminate
		close(m.done)

	case <-ctx.Done():
		return
	}
}

// Register all components
func (m *Manager) Register(c domain.Component) {
	m.components = append(m.components, c)
	logger.Info().Msgf("%s registered", c.Name())
}

// setReady sets the controller ready after all startup is completed
func (m *Manager) setReady() {
	m.hs.SetReady()
}

// Listening to done channel
func (m *Manager) Wait() {
	<-m.done
}
