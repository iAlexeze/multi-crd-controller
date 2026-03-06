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

type managednsClient struct {
	restClient     rest.Interface
	kube           *kubeclient.Kubeclient
	namespace      string
	name           string
	scheme         *runtime.Scheme
	parameterCodec runtime.ParameterCodec
	opts           Options
}

var _ domain.ManagedNamespaceInterface = (*managednsClient)(nil)
var _ domain.Component = (*managednsClient)(nil)

type Options struct {
	Group     string
	Version   string
	APIPath   string
	Namespace string
}

func NewManagednsClient(kube *kubeclient.Kubeclient, scheme *runtime.Scheme, opts Options) *managednsClient {

	return &managednsClient{
		name:           string(domain.ManagedNamespaceResource),
		kube:           kube,
		scheme:         scheme,
		opts:           opts,
		parameterCodec: runtime.NewParameterCodec(scheme), // create a parameterCodec from scheme
	}
}

// Entry point
func (m *managednsClient) Start(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	switch {
	case m.opts.APIPath == "":
		m.opts.APIPath = "/apis"
	case m.opts.Group == "", m.opts.Version == "", m.opts.Namespace == "":
		return fmt.Errorf("required variables: Group, Version, Namespace")
	}

	// Build restclient
	cfg := rest.CopyConfig(m.kube.RestConfig())
	cfg.GroupVersion = &schema.GroupVersion{
		Group:   m.opts.Group,
		Version: m.opts.Version,
	}

	cfg.APIPath = m.opts.APIPath
	cfg.NegotiatedSerializer = serializer.NewCodecFactory(m.scheme)
	cfg.UserAgent = rest.DefaultKubernetesUserAgent()

	m.restClient, _ = rest.RESTClientFor(cfg)

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
