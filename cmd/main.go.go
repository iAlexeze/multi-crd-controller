package main

import (
	"context"

	"github.com/ialexeze/kubernetes-crd-example/pkg/config/api/types/v1alpha1"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/config"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/logger"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

func main() {
	cfg, err := config.Init()
	if err != nil {
		logger.Fatal().AnErr("failed to load configurations", err)
	}

	// initilaize logger
	logger.Init(cfg)

	// define root context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ── 1. Scheme ─────────────────────────────────────────────────────────────
	// Register both built-in types and the CRD types.
	// The scheme is needed by the CRD informer to decode API responses
	// into typed Go structs (*ManagedNamespace).
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		logger.Fatal().Err(err).Msg("failed to add client-go scheme")
	}
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		logger.Fatal().Err(err).Msg("failed to add CRD scheme")
	}

	// create domain components and build manager
	mgr := buildManager(cfg)

	// Start all manager components
	if err = mgr.Start(ctx); err != nil {
		logger.Fatal().AnErr("manager startup error", err)
	}

	// Keep running until cancelled
	mgr.Wait()
}
