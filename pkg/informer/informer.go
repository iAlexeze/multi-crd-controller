package informer

import (
	"context"

	"github.com/ialexeze/multi-crd-controller/pkg/config/domain"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/queue"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

type InformerComponents interface {
	Store() cache.Store
	Controller() cache.Controller
	Name() string
	RestClient() rest.Interface
}

// NewProjectInformer() returns a new ProjectInformer
func NewProjectInformer(
	client domain.ProjectsV1Alpha1nterface,
	wq *queue.Workqueue,
	opts Options,
) *ProjectInformer {
	return &ProjectInformer{
		client: client,
		Informer: Informer{
			name:      string(domain.ProjectInformerResource),
			queue:     wq,
			namespace: opts.Namespace,
			resync:    opts.Resync,
		},
	}
}

// NewManagedNamespaceInformer() returns a new anagedNamespaceInformer informer
func NewManagedNamespaceInformer(
	client domain.ManagedNamespaceV1Alpha1nterface,
	wq *queue.Workqueue,
	opts Options,
) *ManagedNamespaceInformer {
	return &ManagedNamespaceInformer{
		client: client,
		Informer: Informer{
			name:      string(domain.ManagedNamespaceInformerResource),
			queue:     wq,
			namespace: opts.Namespace,
			resync:    opts.Resync,
		},
	}
}

// Shutdown gracefully stops the informer
func (i *Informer) Shutdown(ctx context.Context) {}
