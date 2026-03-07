package controller

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ialexeze/multi-crd-controller/pkg/config/domain"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/event"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/informer"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/kubeclient"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/logger"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/queue"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/utils"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
)

var _ domain.Component = (*Controller)(nil)

type Controller struct {
	kube        *kubeclient.Kubeclient
	informers   []informer.InformerComponents
	event       *event.Event
	q           *queue.Workqueue
	wg          sync.WaitGroup
	workers     int
	reconcilers []domain.Reconciler
	crds        []CRDInfo
}

func NewController(
	kube *kubeclient.Kubeclient,
	registry *ResourceRegistry,
	event *event.Event,
	q *queue.Workqueue,
	workers int,
) *Controller {
	c := &Controller{
		kube:    kube,
		event:   event,
		q:       q,
		workers: workers,
	}

	// Load registry entries
	for _, entry := range registry.entries {
		c.informers = append(c.informers, entry.Informer)
		c.reconcilers = append(c.reconcilers, entry.Reconciler)
		c.crds = append(c.crds, entry.CRD)
	}

	return c
}

func (c *Controller) Start(ctx context.Context) error {
	// CRD check (you may later generalize this per-CRD)
	for _, crd := range c.crds {
		logger.Info().Msgf("checking CRD %s/%s (%s)...", crd.Group, crd.Version, crd.Kind)

		err := utils.RetryBackoff(
			func() error {
				return utils.WaitForCRD(
					c.kube.RestConfig(),
					crd.Group,
					crd.Kind,
					crd.Version,
				)
			},
			5,
			2*time.Second,
		)

		if err != nil {
			return fmt.Errorf("CRD %s/%s (%s) not found: %w",
				crd.Group, crd.Version, crd.Kind, err)
		}

		logger.Info().Msgf("CRD %s/%s (%s) detected", crd.Group, crd.Version, crd.Kind)
	}

	if len(c.informers) == 0 {
		return fmt.Errorf("controller error: no informers registered")
	}

	var hasSyncedFns []cache.InformerSynced

	for _, inf := range c.informers {
		ctrl := inf.Controller()
		if ctrl == nil {
			return fmt.Errorf("informer %s has no controller", inf.Name())
		}

		logger.Debug().Msgf("starting informer controller: %s", inf.Name())
		go ctrl.Run(wait.NeverStop)

		hasSyncedFns = append(hasSyncedFns, ctrl.HasSynced)
	}

	logger.Debug().Msg("waiting for all informer caches to sync...")
	if !cache.WaitForCacheSync(ctx.Done(), hasSyncedFns...) {
		return fmt.Errorf("failed to sync one or more informer caches")
	}
	logger.Info().Msg("all informer caches synced")

	return nil
}

func (c *Controller) RunOrDie(ctx context.Context) {
	logger.Info().Msgf("starting %d workers", c.workers)

	// Start workers
	for i := 0; i < c.workers; i++ {
		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			wait.UntilWithContext(
				ctx,
				func(ctx context.Context) {
					c.runWorker(ctx)
				}, time.Second)
		}()
	}

	// BLOCK until leadership is lost
	<-ctx.Done()

	logger.Info().Msg("leadership lost — draining workers...")

	// Stop accepting new items
	c.q.Shutdown(ctx)

	// Wait for all workers to finish
	c.wg.Wait()

	logger.Info().Msg("controller drained and stopped")
}

// Shutdown gracefully stops the Controller
func (c *Controller) Shutdown(ctx context.Context) {
	logger.Info().Msg("shutting down Controller")
	c.q.Shutdown(ctx)
}

// Controller name
func (c *Controller) Name() string {
	return "smart controller"
}
