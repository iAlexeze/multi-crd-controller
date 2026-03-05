package controller

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ialexeze/kubernetes-crd-example/pkg/config/domain"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/events"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/informer"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/logger"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type Controller struct {
	name     string
	informer *informer.Informer
	events   *events.Recorder
	queue    workqueue.TypedRateLimitingInterface[string]
	workers  int
	wg       sync.WaitGroup
}

var _ domain.Component = (*Controller)(nil)

func NewController(
	informer *informer.Informer,
	events *events.Recorder,
	workers int,
) *Controller {
	return &Controller{
		name:     "smart Controller",
		events:   events,
		informer: informer,
		workers:  workers,
	}
}

func (c *Controller) Start(ctx context.Context) error {
	informer := c.informer

	if informer == nil {
		return fmt.Errorf("controller error: informer not initialized")
	}

	// instantiate queue
	c.queue = informer.Queue()
	ctrl := informer.Controller()

	// Start the controller (Controller)
	logger.Debug().Msg("starting Controller controller...")
	go ctrl.Run(wait.NeverStop)

	// Wait for cache to sync
	logger.Debug().Msg("waiting for cache sync...")
	if !cache.WaitForCacheSync(ctx.Done(), ctrl.HasSynced) {
		return fmt.Errorf("failed to sync Controller cache")
	}
	logger.Info().Msg("Controller cache synced")

	return nil
}

func (c *Controller) Run(ctx context.Context) {
	logger.Info().Msgf("starting %d workers", c.workers)

	// Start workers
	for i := 0; i < c.workers; i++ {
		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			wait.UntilWithContext(ctx, c.runWorker, time.Second)
		}()
	}

	// BLOCK until leadership is lost
	<-ctx.Done()

	logger.Info().Msg("leadership lost — draining workers...")

	// Stop accepting new items
	c.queue.ShutDown()

	// Wait for all workers to finish
	c.wg.Wait()

	logger.Info().Msg("controller drained and stopped")
}

// Shutdown gracefully stops the Controller
func (c *Controller) Shutdown(ctx context.Context) {
	logger.Info().Msg("shutting down Controller")
	c.queue.ShutDown()
}

// Controller name
func (c *Controller) Name() string {
	return c.name
}
