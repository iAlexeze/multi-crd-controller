package main

import (
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/domain"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/config"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/health"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/kubeclient"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/manager"
)

func buildManager(cfg *config.Config) *manager.Manager {
	// create domain components
	var components []domain.Component

	// health server
	hs := health.NewHealthServer("projects", cfg.Health().Port)
	components = append(components, hs)

	// kube client
	kube := kubeclient.NewKubeclient(true, kubeclient.Options{
		Kubeconfig: cfg.Cluster().KubeconfigPath,
		Masterurl:  cfg.Cluster().MasterURL,
	})
	components = append(components, kube)

	// Build and start manager
	mgr := manager.NewManager(cfg.Cluster().DefaultResync)

	// Register all manager components
	for _, comp := range components {
		mgr.Register(comp)
	}

	return mgr
}
