package informer

import (
	"context"
	"time"

	"github.com/ialexeze/multi-crd-controller/pkg/config/domain"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/queue"
	"k8s.io/client-go/tools/cache"
)

type InformerComponents interface {
	Store() cache.Store
	Controller() cache.Controller
	Name() string
}

// NewProjectInformer() returns a new ProjectInformer
func NewProjectInformer(
	client domain.ProjectsV1Alpha1nterface,
	queue *queue.Workqueue,
	namespace string,
	resync time.Duration,
) *ProjectInformer {
	return &ProjectInformer{
		client: client,
		Informer: Informer{
			name:      string(domain.ProjectInformerResource),
			namespace: namespace,
			queue:     *queue,
			resync:    resync,
		},
	}
}

// NewManagedNamespaceInformer() returns a new anagedNamespaceInformer informer
func NewManagedNamespaceInformer(
	client domain.ManagedNamespaceV1Alpha1nterface,
	queue *queue.Workqueue,
	namespace string,
	resync time.Duration,
) *ManagedNamespaceInformer {
	return &ManagedNamespaceInformer{
		client: client,
		Informer: Informer{
			name:      string(domain.ManagedNamespaceInformerResource),
			namespace: namespace,
			queue:     *queue,
			resync:    resync,
		},
	}
}

// Shutdown gracefully stops the informer
func (i *Informer) Shutdown(ctx context.Context) {}
