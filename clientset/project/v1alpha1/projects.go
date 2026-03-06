package v1alpha1

import (
	"context"
	"fmt"

	"github.com/ialexeze/multi-crd-controller/pkg/config/domain"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/kubeclient"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
)

type projectClient struct {
	restClient     rest.Interface
	kube           *kubeclient.Kubeclient
	namespace      string
	name           string
	scheme         *runtime.Scheme
	parameterCodec runtime.ParameterCodec
	opts           Options
}

type Options struct {
	Group     string
	Version   string
	APIPath   string
	Namespace string
}

var _ domain.ProjectInterface = (*projectClient)(nil)
var _ domain.Component = (*projectClient)(nil)

func NewProjectClient(kube *kubeclient.Kubeclient, scheme *runtime.Scheme, opts Options) *projectClient {

	return &projectClient{
		name:           string(domain.ProjectResource),
		kube:           kube,
		scheme:         scheme,
		opts:           opts,
		parameterCodec: runtime.NewParameterCodec(scheme), // create a parameterCodec from scheme
	}
}

// Entry point
func (p *projectClient) Start(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	switch {
	case p.opts.APIPath == "":
		p.opts.APIPath = "/apis"
	case p.opts.Group == "", p.opts.Version == "", p.opts.Namespace == "":
		return fmt.Errorf("required variables: Group, Version, Namespace")
	}

	// Build restclient
	cfg := rest.CopyConfig(p.kube.RestConfig())
	cfg.GroupVersion = &schema.GroupVersion{
		Group:   p.opts.Group,
		Version: p.opts.Version,
	}

	cfg.APIPath = p.opts.APIPath
	cfg.NegotiatedSerializer = serializer.NewCodecFactory(p.scheme)
	cfg.UserAgent = rest.DefaultKubernetesUserAgent()

	p.restClient, _ = rest.RESTClientFor(cfg)
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
