package controller

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ialexeze/multi-crd-controller/pkg/config/domain"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/events"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/informer"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/kubeclient"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/logger"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/utils"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type Controller struct {
	kube     *kubeclient.Kubeclient
	informer *informer.Informer
	events   *events.Recorder
	queue    workqueue.TypedRateLimitingInterface[string]
	wg       sync.WaitGroup
	workers  int
	opts     CustomOptions
}

var _ domain.Component = (*Controller)(nil)

func NewController(
	kube *kubeclient.Kubeclient,
	informer *informer.Informer,
	events *events.Recorder,
	workers int,
	opts CustomOptions,
) *Controller {
	return &Controller{
		kube:     kube,
		informer: informer,
		events:   events,
		workers:  workers,
		opts:     opts,
	}
}

type CustomOptions struct {
	IsCustom bool
	Group    string
	Kind     string
	Version  string
}

func (c *Controller) Start(ctx context.Context) error {
	// Confirm CRD type presence in cluster if custom
	if c.opts.IsCustom {
		logger.Info().Msg("Custom controller setup detected")
		required := map[string]string{
			"Group":   c.opts.Group,
			"Kind":    c.opts.Kind,
			"Version": c.opts.Version,
		}

		var missing []string
		for k, v := range required {
			if v == "" {
				missing = append(missing, k)
			}
		}

		if len(missing) > 0 {
			err := fmt.Sprintf("missing required parameter(s): %s", strings.Join(missing, ", "))
			logger.Error().Msgf("%s", err)
			return fmt.Errorf("%s", err)
		}

		// Try with backoff
		logger.Info().
			Msgf("checking %s CRD: %s/%s...", c.opts.Kind, c.opts.Group, c.opts.Version)
		if err := utils.RetryBackoff(
			func() error {
				return utils.WaitForCRD(
					c.kube.RestConfig(),
					c.opts.Group,
					c.opts.Kind,
					c.opts.Version,
				)
			}, 5, 2*time.Second,
		); err != nil {
			logger.Error().Err(err).
				Msgf("%s CRD: %s/%s... not found", c.opts.Kind, c.opts.Group, c.opts.Version)
			return err
		}

		logger.Info().
			Msgf("Found %s CRD: %s/%s...", c.opts.Kind, c.opts.Group, c.opts.Version)
	}

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
	return "smart controller"
}
