package informer

import (
	"context"

	managednsTypeV1 "github.com/ialexeze/multi-crd-controller/pkg/config/api/types/managedNamespace/v1alpha1"
	"github.com/ialexeze/multi-crd-controller/pkg/config/domain"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/logger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

func (m *ManagedNamespaceInformer) watchManagedNamespaceResources(ctx context.Context) (cache.Store, cache.Controller) {
	return cache.NewInformerWithOptions(
		cache.InformerOptions{
			ListerWatcher: &cache.ListWatch{
				ListFunc: func(lo metav1.ListOptions) (result runtime.Object, err error) {
					logger.Debug().Msg("MNs - ListFunc called")
					result, err = m.client.ManagedNamespaces(m.namespace).List(ctx, lo)
					if err != nil {
						return nil, err
					}
					count := len(result.(*managednsTypeV1.ManagedNamespaceList).Items)
					logger.Debug().Msgf("MNs - ListFunc returned %d items", count)
					return result, err
				},
				WatchFunc: func(lo metav1.ListOptions) (watch.Interface, error) {
					logger.Debug().Msg("MNs - WatchFunc called - establishing watch")
					w, err := m.client.ManagedNamespaces(m.namespace).Watch(ctx, lo)
					if err != nil {
						return nil, err
					}
					logger.Debug().Msg("MNs - Watch established successfully")
					return w, err
				},
			},
			ObjectType: &managednsTypeV1.ManagedNamespace{},
			Handler: cache.ResourceEventHandlerFuncs{
				AddFunc:    func(obj interface{}) { m.queue.Enqueue(obj, domain.ManagedNamespaceResource) },
				UpdateFunc: func(oldObj, newObj interface{}) { m.queue.Enqueue(newObj, domain.ManagedNamespaceResource) },
				DeleteFunc: func(obj interface{}) { m.queue.Enqueue(obj, domain.ManagedNamespaceResource) },
			},
			ResyncPeriod: m.resync,
		},
	)
}

var _ domain.Component = (*ManagedNamespaceInformer)(nil)

// Methods
func (m *ManagedNamespaceInformer) Start(ctx context.Context) error {
	m.namespace = m.client.Namespace()
	m.store, m.controller = m.watchManagedNamespaceResources(ctx)
	return nil
}

func (m *ManagedNamespaceInformer) Controller() cache.Controller { return m.controller }

func (m *ManagedNamespaceInformer) Store() cache.Store { return m.store }

func (m *ManagedNamespaceInformer) Name() string { return m.name }

func (m *ManagedNamespaceInformer) RestClient() rest.Interface { return m.client.RestClient() }
