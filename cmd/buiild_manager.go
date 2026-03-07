package main

import (
	"fmt"
	"strings"

	"github.com/ialexeze/multi-crd-controller/pkg/config/domain"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/config"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/controller"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/event"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/health"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/informer"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/kubeclient"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/queue"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/reconciler"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	mnsTypev1 "github.com/ialexeze/multi-crd-controller/pkg/config/api/types/managedNamespace/v1alpha1"
	projectTypev1 "github.com/ialexeze/multi-crd-controller/pkg/config/api/types/project/v1alpha1"
	mnsClientV1alpha1 "github.com/ialexeze/multi-crd-controller/pkg/config/clientset/managedNamespace/v1alpha1"
	projectsClientV1alpha1 "github.com/ialexeze/multi-crd-controller/pkg/config/clientset/project/v1alpha1"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/logger"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/manager"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

type startupCfg struct {
	controller *controller.Controller
	event      *event.Event
	kube       *kubeclient.Kubeclient
	manager    *manager.Manager
}

func buildManager(cfg *config.Config) *startupCfg {
	scheme, err := buildScheme()
	if err != nil {
		logger.Fatal().Err(err).Msg("scheme creation error")
	}

	var components []domain.Component

	// health
	hs := health.NewHealthServer("projects", cfg)
	components = append(components, hs)

	// kube
	kube := kubeclient.NewKubeclient(kubeclient.Config{
		Kubeconfig: cfg.Cluster().KubeconfigPath,
		Masterurl:  cfg.Cluster().MasterURL,
		Scheme:     scheme,
	})
	components = append(components, kube)

	// queue
	wq := queue.NewWorkqueue()
	components = append(components, wq)

	// clients
	projectsClient := projectsClientV1alpha1.NewProjectClient(kube, kubeclient.Options{
		Group:   projectTypev1.Group,
		Version: projectTypev1.Version,
		APIPath: projectTypev1.APIPath,
	})
	components = append(components, projectsClient)

	managedNamespaceClient := mnsClientV1alpha1.NewManagednsClient(kube, kubeclient.Options{
		Group:   mnsTypev1.Group,
		Version: mnsTypev1.Version,
		APIPath: mnsTypev1.APIPath,
	})
	components = append(components, managedNamespaceClient)

	// informers
	projInformer := informer.NewProjectInformer(
		projectsClient,
		wq,
		informer.Options{
			Namespace: cfg.Cluster().Namespace,
			Resync:    cfg.Cluster().DefaultResync,
		},
	)
	components = append(components, projInformer)

	mnsInformer := informer.NewManagedNamespaceInformer(
		managedNamespaceClient,
		wq,
		informer.Options{
			Namespace: cfg.Cluster().Namespace,
			Resync:    cfg.Cluster().DefaultResync,
		},
	)
	components = append(components, mnsInformer)

	// events
	ev := event.NewEvent(kube, scheme, event.Options{Component: cfg.App().Name})
	components = append(components, ev)

	// reconcilers
	projReconciler := reconciler.NewProjectReconciler(projInformer, ev)
	mnsReconciler := reconciler.NewManagedNamespaceReconciler(kube, mnsInformer, ev)

	// registry
	reg := controller.NewRegistry()
	reg.Register(
		domain.ProjectResource,
		controller.CRDInfo{
			Group:   projectTypev1.Group,
			Version: projectTypev1.Version,
			Kind:    projectTypev1.Kind,
			APIPath: projectTypev1.APIPath,
		},
		projInformer,
		projReconciler,
	)
	reg.Register(
		domain.ManagedNamespaceResource,
		controller.CRDInfo{
			Group:   mnsTypev1.Group,
			Version: mnsTypev1.Version,
			Kind:    mnsTypev1.Kind,
			APIPath: mnsTypev1.APIPath,
		},
		mnsInformer,
		mnsReconciler,
	)

	// controller
	ctrl := controller.NewController(
		kube,
		reg,
		ev,
		wq,
		cfg.Cluster().Workers,
	)
	components = append(components, ctrl)

	// manager
	mgr := manager.NewManager(hs, cfg.Cluster().DefaultResync)

	fmt.Println("==========================")
	fmt.Println("REGISTERING MANAGER COMPONENTS...")
	for _, comp := range components {
		mgr.Register(comp)
		logger.Info().Msgf("[%s] component registered", comp.Name())
	}
	var names []string
	for _, comp := range components {
		names = append(names, comp.Name())
	}
	fmt.Printf("Available Components: %s\n", strings.Join(names, ", "))
	fmt.Println("==========================")

	return &startupCfg{
		event:      ev,
		controller: ctrl,
		kube:       kube,
		manager:    mgr,
	}
}

func buildScheme() (*runtime.Scheme, error) {
	scheme := runtime.NewScheme()

	// 1. Register built-in Kubernetes types
	metav1.AddToGroupVersion(scheme, metav1.SchemeGroupVersion)

	// 2. Register core Kubernetes types
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return nil, err
	}

	// 3. Register your CRDs
	if err := projectTypev1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := mnsTypev1.AddToScheme(scheme); err != nil {
		return nil, err
	}

	return scheme, nil
}
