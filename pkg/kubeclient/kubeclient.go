package kubeclient

import (
	"context"
	"fmt"

	projectTypeV1 "github.com/ialexeze/multi-crd-controller/pkg/config/api/types/project/v1alpha1"
	"github.com/ialexeze/multi-crd-controller/pkg/config/domain"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/logger"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Kubeclient struct {
	isCustom   bool
	name       string
	restConfig *rest.Config
	clientset  kubernetes.Interface
	restClient rest.Interface

	Opts Options
}

type Options struct {
	Kubeconfig string
	Masterurl  string
	Scheme     *runtime.Scheme
}

var _ domain.Component = (*Kubeclient)(nil)

func NewKubeclient(isCustom bool, opts Options) *Kubeclient {
	if !isCustom && opts.Scheme == nil {
		opts.Scheme = scheme.Scheme
	}

	return &Kubeclient{
		name:     "kubeclient",
		isCustom: isCustom,
		Opts:     opts,
	}
}

func (k *Kubeclient) Start(ctx context.Context) error {
	cfg, err := k.buildConfig()
	if err != nil {
		return err
	}

	// Populate Kubeclient's rest config
	k.restConfig = cfg

	// Build clientset and restClient conditonally
	logger.Debug().Msg("creating clients...")

	// Default
	logger.Info().Msg("clientset for leader election")
	k.clientset, err = kubernetes.NewForConfig(k.restConfig)
	if err != nil {
		return err
	}

	if k.isCustom {
		logger.Info().Msg("rest client")
		k.restClient, err = k.buildRestClient()
		if err != nil {
			return err
		}

		// Add scheme
		if k.Opts.Scheme == nil {
			return fmt.Errorf("scheme not defined: custom resource scheme cannot be nil")
		}
	}

	return nil
}

func (k *Kubeclient) buildConfig() (*rest.Config, error) {
	if k.Opts.Kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags(k.Opts.Masterurl, k.Opts.Kubeconfig)
	}
	return rest.InClusterConfig()
}

func (k *Kubeclient) buildRestClient() (*rest.RESTClient, error) {
	config := *k.restConfig

	config.ContentConfig.GroupVersion = &schema.GroupVersion{
		Group:   projectTypeV1.GroupName,
		Version: projectTypeV1.GroupVersion,
	}

	config.APIPath = "/apis"
	config.NegotiatedSerializer = serializer.NewCodecFactory(k.Opts.Scheme).WithoutConversion()
	config.UserAgent = rest.DefaultKubernetesUserAgent()

	return rest.RESTClientFor(&config)
}

// Methods
func (k *Kubeclient) Shutdown(ctx context.Context) {}

func (k *Kubeclient) Name() string {
	return k.name
}

func (k *Kubeclient) RestConfig() *rest.Config {
	return k.restConfig
}

func (k *Kubeclient) Clientset() kubernetes.Interface {
	return k.clientset
}

func (k *Kubeclient) RestClient() rest.Interface {
	return k.restClient
}
