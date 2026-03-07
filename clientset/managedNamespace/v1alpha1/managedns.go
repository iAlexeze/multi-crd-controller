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
	parameterCodec runtime.ParameterCodec
	opts           kubeclient.Options
}

var _ domain.ManagedNamespaceInterface = (*managednsClient)(nil)
var _ domain.Component = (*managednsClient)(nil)

func NewManagednsClient(kube *kubeclient.Kubeclient, opts kubeclient.Options) *managednsClient {

	return &managednsClient{
		name: string(domain.ManagedNamespaceResource),
		kube: kube,
		opts: opts,
	}
}

// Entry point
func (m *managednsClient) Start(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// create a parameterCodec from scheme
	m.parameterCodec = m.kube.RuntimeParameterCodec()

	// Assign rest client
	restClient, err := m.kube.SharedClientFactory(m.opts)
	if err != nil {
		return err
	}

	m.restClient = restClient
	return nil
}

// Just to fully implement the components interface
func (c *managednsClient) Shutdown(ctx context.Context) {}

// Getters
func (m *managednsClient) Name() string {
	return m.name
}

func (m *managednsClient) Namespace() string {
	return m.namespace
}

func (m *managednsClient) RestClient() rest.Interface {
	return m.restClient
}
