package informer

import (
	"context"

	projectTypeV1 "github.com/ialexeze/multi-crd-controller/pkg/config/api/types/project/v1alpha1"
	"github.com/ialexeze/multi-crd-controller/pkg/config/domain"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/logger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

func (p *ProjectInformer) watchProjectResources(ctx context.Context) (cache.Store, cache.Controller) {
	return cache.NewInformerWithOptions(
		cache.InformerOptions{
			ListerWatcher: &cache.ListWatch{
				ListFunc: func(lo metav1.ListOptions) (result runtime.Object, err error) {
					logger.Debug().Msg("PROJ - ListFunc called")
					result, err = p.client.Projects(p.namespace).List(ctx, lo)
					if err != nil {
						return nil, err
					}
					count := len(result.(*projectTypeV1.ProjectList).Items)
					logger.Debug().Msgf("PROJ - ListFunc returned %d items", count)
					return result, err
				},
				WatchFunc: func(lo metav1.ListOptions) (watch.Interface, error) {
					logger.Debug().Msg("PROJ - WatchFunc called - establishing watch")
					w, err := p.client.Projects(p.namespace).Watch(ctx, lo)
					if err != nil {
						return nil, err
					}
					logger.Debug().Msg("PROJ - Watch established successfully")
					return w, err
				},
			},
			ObjectType: &projectTypeV1.Project{},
			Handler: cache.ResourceEventHandlerFuncs{
				AddFunc:    func(obj interface{}) { p.queue.Enqueue(obj, domain.ProjectResource) },
				UpdateFunc: func(oldObj, newObj interface{}) { p.queue.Enqueue(newObj, domain.ProjectResource) },
				DeleteFunc: func(obj interface{}) { p.queue.Enqueue(obj, domain.ProjectResource) },
			},
			ResyncPeriod: p.resync,
		},
	)
}

var _ domain.Component = (*ProjectInformer)(nil)

// Methods
func (p *ProjectInformer) Start(ctx context.Context) error {
	p.namespace = p.client.Namespace()
	p.store, p.controller = p.watchProjectResources(ctx)
	return nil
}

func (p *ProjectInformer) Controller() cache.Controller { return p.controller }

func (p *ProjectInformer) Store() cache.Store { return p.store }

func (p *ProjectInformer) Name() string { return p.name }

func (p *ProjectInformer) RestClient() rest.Interface { return p.client.RestClient() }
