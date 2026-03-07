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
	opts           kubeclient.Options
	name           string
	namespace      string
	parameterCodec runtime.ParameterCodec
}

var _ domain.ProjectInterface = (*projectClient)(nil)
var _ domain.Component = (*projectClient)(nil)

func NewProjectClient(kube *kubeclient.Kubeclient, opts kubeclient.Options) *projectClient {

	return &projectClient{
		name: string(domain.ProjectResource),
		kube: kube,
		opts: opts,
	}
}

// Entry point
func (p *projectClient) Start(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// create a parameterCodec from scheme
	p.parameterCodec = p.kube.RuntimeParameterCodec()

	// Assign rest client
	restClient, err := p.kube.SharedClientFactory(p.opts)
	if err != nil {
		return err
	}

	p.restClient = restClient
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
