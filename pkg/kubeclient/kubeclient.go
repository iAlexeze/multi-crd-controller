package kubeclient

import (
	"context"
	"fmt"

	"github.com/ialexeze/multi-crd-controller/pkg/config/domain"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/logger"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Kubeclient struct {
	name       string
	restConfig *rest.Config
	clientset  kubernetes.Interface
	dynamic    dynamic.Interface
	scheme     *runtime.Scheme
	Opts       Options
}

type Options struct {
	Kubeconfig string
	Masterurl  string
	Scheme     *runtime.Scheme // REQUIRED
}

var _ domain.Component = (*Kubeclient)(nil)

func NewKubeclient(opts Options) *Kubeclient {
	if opts.Scheme == nil {
		panic("kubeclient.Options.Scheme cannot be nil")
	}

	return &Kubeclient{
		name:   "kubeclient",
		scheme: opts.Scheme,
		Opts:   opts,
	}
}

func (k *Kubeclient) Start(ctx context.Context) error {
	cfg, err := k.buildConfig()
	if err != nil {
		return err
	}

	// Store config
	k.restConfig = cfg

	// Build core clientset
	logger.Info().Msg("creating core clientset")
	k.clientset, err = kubernetes.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to create clientset: %w", err)
	}

	// Build dynamic client
	logger.Info().Msg("creating dynamic client")
	k.dynamic, err = dynamic.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %w", err)
	}

	return nil
}

func (k *Kubeclient) buildConfig() (*rest.Config, error) {
	if k.restConfig != nil {
		return k.restConfig, nil
	}

	if k.scheme == nil {
		return nil, fmt.Errorf("scheme is nil in kubeclient")
	}

	var cfg *rest.Config
	var err error

	if k.Opts.Kubeconfig != "" {
		logger.Info().Msg("using kubeconfig")
		cfg, err = clientcmd.BuildConfigFromFlags(k.Opts.Masterurl, k.Opts.Kubeconfig)
	} else {
		cfg, err = rest.InClusterConfig()
	}
	if err != nil {
		return nil, err
	}

	// Ensure the config uses our global scheme
	cfg.NegotiatedSerializer = serializer.NewCodecFactory(k.scheme)

	return cfg, nil
}

func (k *Kubeclient) Shutdown(ctx context.Context) {}

func (k *Kubeclient) Name() string { return k.name }

func (k *Kubeclient) RestConfig() *rest.Config { return k.restConfig }

func (k *Kubeclient) Clientset() kubernetes.Interface { return k.clientset }

func (k *Kubeclient) Dynamic() dynamic.Interface { return k.dynamic }

func (k *Kubeclient) Scheme() *runtime.Scheme { return k.scheme }
