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
	"golang.org/x/sys/windows/registry"
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

type reconcilerCfg struct {
	event        *event.Event
	projInformer informer.InformerComponents
	mnsInformer  informer.InformerComponents
	kube         *kubeclient.Kubeclient
}

func buildManager(cfg *config.Config) *startupCfg {
	// create scheme
	scheme, err := buildScheme()
	if err != nil {
		logger.Fatal().Err(err).Msg("scheme creation error")
	}

	// create domain components and informers
	var components []domain.Component
	var informers []informer.InformerComponents

	// health server
	hs := health.NewHealthServer("projects", cfg)
	components = append(components, hs)

	// kube client
	kube := kubeclient.NewKubeclient(true, kubeclient.Options{
		Kubeconfig: cfg.Cluster().KubeconfigPath,
		Masterurl:  cfg.Cluster().MasterURL,
		Scheme:     scheme,
	})
	components = append(components, kube)

	// queue
	queue := queue.NewWorkqueue()
	components = append(components, queue)

	// projects
	projectsClient := projectsClientV1alpha1.NewProjectClient(kube, scheme, cfg.Cluster().Namespace)
	components = append(components, projectsClient)

	// managednamespace
	managedNamespaceClient := mnsClientV1alpha1.NewManagednsClient(kube, scheme, cfg.Cluster().Namespace)
	components = append(components, managedNamespaceClient)

	// informers
	projInformer := informer.NewProjectInformer(
		projectsClient,
		queue,
		cfg.Cluster().Namespace,
		cfg.Cluster().DefaultResync,
	)
	informers = append(informers, projInformer)
	components = append(components, projInformer)

	mnsInformer := informer.NewManagedNamespaceInformer(
		managedNamespaceClient,
		queue,
		cfg.Cluster().Namespace,
		cfg.Cluster().DefaultResync,
	)
	informers = append(informers, mnsInformer)
	components = append(components, mnsInformer)

	// event
	event := event.NewEvent(kube, scheme, event.Options{Component: cfg.App().Name})
	components = append(components, event)

	// controller
	ctrl := controller.NewController(
		kube,
		informers,
		event,
		queue,
		cfg.Cluster().Workers,
		controller.CustomOptions{
			IsCustom: true,
			Group:    projectTypev1.Group,
			Kind:     projectTypev1.Kind,
			Version:  projectTypev1.Version,
		},
	)
	components = append(components, ctrl) // Needed to get the controller informer synced and ready for manager to finish infrastructure setup

	// Build reconcilers
	reconcilers := buildReconcilers(&reconcilerCfg{
		event:        event,
		projInformer: projInformer,
		mnsInformer:  mnsInformer,
		kube:         kube,
	})

	// Build and start manager
	mgr := manager.NewManager(hs, cfg.Cluster().DefaultResync)

	// Register all manager components
	fmt.Println("==========================")
	fmt.Println("REGISTERING MANAGER COMPONENTS...")
	for _, comp := range components {
		mgr.Register(comp)
	}

	var componentNames []string
	for _, comp := range components {
		componentNames = append(componentNames, comp.Name())
	}
	fmt.Printf("Available Components: %s\n", strings.Join(componentNames, ", "))
	fmt.Println("==========================")

	// Register reconcilers to controller
	fmt.Println("REGISTERING RECONCILERS...")
	for _, rec := range reconcilers {
		ctrl.RegisterReconcilers(rec)
	}

	var reconcilerNames []string
	for _, rec := range reconcilers {
		reconcilerNames = append(reconcilerNames, string(rec.Resource()))
	}
	fmt.Printf("Available Reconcilers: %s\n", strings.Join(reconcilerNames, ", "))
	fmt.Println("==========================")

	return &startupCfg{
		event:      event,
		controller: ctrl,
		kube:       kube,
		manager:    mgr,
	}
}

func buildReconcilers(cfg *reconcilerCfg) (reconcilers []domain.Reconciler) {
	// Create reconcilers
	// Project
	projReconciler := reconciler.NewProjectReconciler(cfg.projInformer, cfg.event)
	reconcilers = append(reconcilers, projReconciler)

	// ManagedNamespace
	managedNsReconciler := reconciler.NewManagedNamespaceReconciler(cfg.kube, cfg.mnsInformer, cfg.event)
	reconcilers = append(reconcilers, managedNsReconciler)

	return reconcilers
}

func buildScheme() (*runtime.Scheme, error) {
	// ── Add scheme ─────────────────────────────────────────────────────────────
	// Register both built-in types and the CRD types.
	// The scheme is needed by the CRD informer to decode API responses
	// into typed Go structs (Example *Project).
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		logger.Error().Err(err).Msg("failed to add client-go scheme")
		return nil, err
	}

	// Add project
	if err := projectTypev1.AddToScheme(scheme); err != nil {
		logger.Error().Err(err).Msg("failed to add Project CRD scheme")
		return nil, err
	}

	// Add managedNs
	if err := mnsTypev1.AddToScheme(scheme); err != nil {
		logger.Error().Err(err).Msg("failed to add ManagedNamespace CRD scheme")
		return nil, err
	}

	return scheme, nil
}