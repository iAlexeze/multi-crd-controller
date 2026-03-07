package kubeclient

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
)

type Options struct {
	Group        string               // Required if GroupVersion is not specified
	Version      string               // Required if GroupVersion is not specified
	GroupVersion *schema.GroupVersion // Optional (can be used if Group and Version are not specified)
	APIPath      string
}

// SharedClientFactory provides a simple way to build clients from config
func (k *Kubeclient) SharedClientFactory(opts Options) (*rest.RESTClient, error) {
	switch {
	case opts.APIPath == "":
		opts.APIPath = "/apis"
	case opts.GroupVersion == nil:
		if opts.Group == "" && opts.Version == "" {
			return nil, fmt.Errorf("required variables: Group, Version")
		}
		opts.GroupVersion = &schema.GroupVersion{
			Group:   opts.Group,
			Version: opts.Version,
		}
	}

	// Build restclient
	cfg := rest.CopyConfig(k.RestConfig())
	cfg.GroupVersion = opts.GroupVersion

	cfg.APIPath = opts.APIPath
	cfg.NegotiatedSerializer = serializer.NewCodecFactory(k.Scheme())
	cfg.UserAgent = rest.DefaultKubernetesUserAgent()

	return rest.RESTClientFor(cfg)

}

func (k *Kubeclient) RuntimeParameterCodec() runtime.ParameterCodec {
	return runtime.NewParameterCodec(k.Scheme())
}
