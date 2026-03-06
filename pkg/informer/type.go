package informer

import (
	"time"

	"github.com/ialexeze/multi-crd-controller/pkg/config/domain"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/queue"
	"k8s.io/client-go/tools/cache"
)

type Options struct {
	Namespace string
	Resync    time.Duration
}

type Informer struct {
	name       string
	namespace  string
	resync     time.Duration
	store      cache.Store
	controller cache.Controller
	queue      *queue.Workqueue
}

type ProjectInformer struct {
	Informer
	client domain.ProjectsV1Alpha1nterface
}

type ManagedNamespaceInformer struct {
	Informer
	client domain.ManagedNamespaceV1Alpha1nterface
}
