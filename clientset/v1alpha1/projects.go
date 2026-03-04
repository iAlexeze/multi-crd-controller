package v1alpha1

import (
	"context"

	"github.com/ialexeze/kubernetes-crd-example/pkg/config/domain"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/kubeclient"
	"k8s.io/client-go/rest"
)

type projectClient struct {
	restClient rest.Interface
	kube       *kubeclient.Kubeclient
	namespace  string
	name       string
}

var _ domain.ProjectInterface = (*projectClient)(nil)
var _ domain.Component = (*projectClient)(nil)

func NewProjectClient(kube *kubeclient.Kubeclient, namespace string) *projectClient {
	return &projectClient{
		kube:      kube,
		namespace: namespace,
	}
}

func (p *projectClient) Start(ctx context.Context) error {
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
