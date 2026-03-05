package main

import (
	"context"

	"github.com/ialexeze/kubernetes-crd-example/pkg/config/domain"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/config"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/controller"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/health"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/informer"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/kubeclient"
	"k8s.io/apimachinery/pkg/runtime"

	clientV1alpha1 "github.com/ialexeze/kubernetes-crd-example/pkg/config/clientset/v1alpha1"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/leader"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/manager"
)

func buildManager(cfg *config.Config, scheme *runtime.Scheme) *manager.Manager {
	// create domain components
	var components []domain.Component

	// health server
	hs := health.NewHealthServer("projects", cfg.Health().Port)
	components = append(components, hs)

	// kube client
	kube := kubeclient.NewKubeclient(true, kubeclient.Options{
		Kubeconfig: cfg.Cluster().KubeconfigPath,
		Masterurl:  cfg.Cluster().MasterURL,
		Scheme:     scheme,
	})
	components = append(components, kube)

	// projects
	projects := clientV1alpha1.NewProjectClient(kube, scheme, cfg.Cluster().Namespace)
	components = append(components, projects)

	// informer
	informer := informer.NewInformer(projects, cfg.Cluster().DefaultResync)
	components = append(components, informer)

	// controller
	ctrl := controller.NewController(informer, cfg.Cluster().Workers)
	components = append(components, ctrl)

	// leader election
	leader := leader.NewLeaderElection(
		kube,
		func(ctx context.Context) { ctrl.Run(ctx) }, // controller run
		leader.Options{
			Namespace:     cfg.Cluster().Namespace,
			LeaseDuration: cfg.Leader().LeaseDuration,
			RenewDeadline: cfg.Leader().RenewDeadline,
			RetryPeriod:   cfg.Leader().RetryPeriod,
		})
	components = append(components, leader)

	// Build and start manager
	mgr := manager.NewManager(hs, cfg.Cluster().DefaultResync)

	// Register all manager components
	for _, comp := range components {
		mgr.Register(comp)
	}

	return mgr
}
