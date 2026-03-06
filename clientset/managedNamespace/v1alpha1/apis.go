package v1alpha1

import (
	"context"

	managednsTypeV1 "github.com/ialexeze/multi-crd-controller/pkg/config/api/types/managedNamespace/v1alpha1"
	"github.com/ialexeze/multi-crd-controller/pkg/config/domain"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/logger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

// ManagedNamespaces implements the ManagedNamespace interface
func (p *managednsClient) ManagedNamespaces(namespace string) domain.ManagedNamespaceInterface {
	return &managednsClient{
		name:           string(domain.ProjectResource),
		restClient:     p.restClient,
		namespace:      namespace,
		scheme:         p.scheme,
		parameterCodec: p.parameterCodec,
	}
}

// API Functions
func (p *managednsClient) List(ctx context.Context, opts metav1.ListOptions) (*managednsTypeV1.ManagedNamespaceList, error) {
	if p.restClient == nil {
		logger.Fatal().Msg("restClient is nil - check client initialization")
	}

	result := managednsTypeV1.ManagedNamespaceList{}
	err := p.restClient.
		Get().
		Namespace(p.Namespace()).
		Resource(p.Name()).
		VersionedParams(&opts, p.parameterCodec).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (p *managednsClient) Get(ctx context.Context, name string, opts metav1.GetOptions) (*managednsTypeV1.ManagedNamespace, error) {
	result := managednsTypeV1.ManagedNamespace{}
	err := p.restClient.
		Get().
		Namespace(p.Namespace()).
		Resource(p.Name()).
		Name(name).
		VersionedParams(&opts, p.parameterCodec).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (p *managednsClient) Create(ctx context.Context, mns *managednsTypeV1.ManagedNamespace) (*managednsTypeV1.ManagedNamespace, error) {
	result := managednsTypeV1.ManagedNamespace{}
	err := p.restClient.
		Post().
		Namespace(p.Namespace()).
		Resource(p.Name()).
		Body(mns).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (p *managednsClient) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return p.restClient.
		Get().
		Namespace(p.Namespace()).
		Resource(p.Name()).
		VersionedParams(&opts, p.parameterCodec).
		Watch(ctx)
}
