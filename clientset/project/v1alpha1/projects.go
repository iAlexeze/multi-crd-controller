package v1alpha1

import (
	"context"

	"github.com/ialexeze/multi-crd-controller/pkg/config/domain"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/kubeclient"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
)

type projectClient struct {
	restClient     rest.Interface
	kube           *kubeclient.Kubeclient
	namespace      string
	name           string
	scheme         *runtime.Scheme
	parameterCodec runtime.ParameterCodec
}

var _ domain.ProjectInterface = (*projectClient)(nil)
var _ domain.Component = (*projectClient)(nil)

func NewProjectClient(kube *kubeclient.Kubeclient, scheme *runtime.Scheme, namespace string) *projectClient {

	return &projectClient{
		name:           "projects",
		kube:           kube,
		namespace:      namespace,
		scheme:         scheme,
		parameterCodec: runtime.NewParameterCodec(scheme), // create a parameterCodec from scheme
	}
}

func (p *projectClient) Start(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	p.restClient = p.kube.RestClient()
	return nil
}

// Just to fully implement the components interface
func (c *projectClient) Shutdown(ctx context.Context) {}

// Getters
func (p *projectClient) Name() string {
	return p.name
}

func (p *projectClient) Namespace() string {
	return p.namespace
}

func (p *projectClient) RestClient() rest.Interface {
	return p.restClient
}
