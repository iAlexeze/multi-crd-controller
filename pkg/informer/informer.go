package informer

import (
	"context"
	"time"

	projectTypeV1 "github.com/ialexeze/multi-crd-controller/pkg/config/api/types/project/v1alpha1"
	"github.com/ialexeze/multi-crd-controller/pkg/config/domain"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/logger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type Informer struct {
	name          string
	namespace     string
	projectClient domain.ProjectsV1Alpha1nterface
	resync        time.Duration
	store         cache.Store
	controller    cache.Controller
	queue         workqueue.TypedRateLimitingInterface[string]
}

var _ domain.Component = (*Informer)(nil)

func NewInformer(
	projectClient domain.ProjectsV1Alpha1nterface,
	resync time.Duration,
) *Informer {
	return &Informer{
		name:          "smart informer",
		projectClient: projectClient,
		resync:        resync,
		queue:         workqueue.NewTypedRateLimitingQueue(workqueue.DefaultTypedControllerRateLimiter[string]()),
	}
}

func (i *Informer) Start(ctx context.Context) error {
	i.namespace = i.projectClient.Namespace()
	i.store, i.controller = i.watchResources(ctx)
	return nil
}

func (i *Informer) watchResources(ctx context.Context) (cache.Store, cache.Controller) {
	return cache.NewInformerWithOptions(
		cache.InformerOptions{
			ListerWatcher: &cache.ListWatch{
				ListFunc: func(lo metav1.ListOptions) (result runtime.Object, err error) {
					logger.Debug().Msg("📋 ListFunc called")
					result, err = i.projectClient.Projects(i.namespace).List(ctx, lo)
					if err != nil {
						return nil, err
					}
					count := len(result.(*projectTypeV1.ProjectList).Items)
					logger.Debug().Msgf("📋 ListFunc returned %d items", count)
					return result, err
				},
				WatchFunc: func(lo metav1.ListOptions) (watch.Interface, error) {
					logger.Debug().Msg("👀 WatchFunc called - establishing watch")
					w, err := i.projectClient.Projects(i.namespace).Watch(ctx, lo)
					if err != nil {
						return nil, err
					}
					logger.Debug().Msg("👀 Watch established successfully")
					return w, err
				},
			},
			ObjectType: &projectTypeV1.Project{},
			Handler: cache.ResourceEventHandlerFuncs{
				AddFunc:    func(obj interface{}) { i.enqueue(obj) },
				UpdateFunc: func(oldObj, newObj interface{}) { i.enqueue(newObj) },
				DeleteFunc: func(obj interface{}) { i.enqueue(obj) },
			},
			ResyncPeriod: i.resync,
		},
	)
}

// enqueue adds the object's key to the workqueue
func (i *Informer) enqueue(obj interface{}) {
	// Handle tombstone (deleted objects)
	if tombstone, ok := obj.(cache.DeletedFinalStateUnknown); ok {
		obj = tombstone.Obj
	}

	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get key")
		return
	}

	i.queue.Add(key)
	logger.Debug().Msgf("Enqueued: %s", key)
}

// Shutdown gracefully stops the informer
func (i *Informer) Shutdown(ctx context.Context) {
	logger.Info().Msg("shutting down informer")
	i.queue.ShutDown()
}

// Methods
func (i *Informer) Controller() cache.Controller {
	return i.controller
}

func (i *Informer) Queue() workqueue.TypedRateLimitingInterface[string] {
	return i.queue
}

func (i *Informer) Store() cache.Store {
	return i.store
}

func (i *Informer) Name() string {
	return i.name
}
