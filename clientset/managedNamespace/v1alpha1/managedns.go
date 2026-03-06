package v1alpha1

import (
	"context"

	"github.com/ialexeze/multi-crd-controller/pkg/config/domain"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/kubeclient"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
)

type managednsClient struct {
	restClient     rest.Interface
	kube           *kubeclient.Kubeclient
	namespace      string
	name           string
	scheme         *runtime.Scheme
	parameterCodec runtime.ParameterCodec
}

var _ domain.ManagedNamespaceInterface = (*managednsClient)(nil)
var _ domain.Component = (*managednsClient)(nil)

func NewManagednsClient(kube *kubeclient.Kubeclient, scheme *runtime.Scheme, namespace string) *managednsClient {

	return &managednsClient{
		name:           string(domain.ManagedNamespaceResource),
		kube:           kube,
		namespace:      namespace,
		scheme:         scheme,
		parameterCodec: runtime.NewParameterCodec(scheme), // create a parameterCodec from scheme
	}
}

func (p *managednsClient) Start(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	p.restClient = p.kube.RestClient()
	return nil
}

// Just to fully implement the components interface
func (c *managednsClient) Shutdown(ctx context.Context) {}

// Getters
func (p *managednsClient) Name() string {
	return p.name
}

func (p *managednsClient) Namespace() string {
	return p.namespace
}

func (p *managednsClient) RestClient() rest.Interface {
	return p.restClient
}
