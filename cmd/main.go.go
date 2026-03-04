package main

import (
	"context"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/config"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/logger"
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

	// create domain components and build manager
	mgr := buildManager(cfg)

	// Start all manager components
	if err = mgr.Start(ctx); err != nil {
		logger.Fatal().AnErr("manager startup error", err)
	}

	// Keep running until cancelled
	mgr.Wait()
}
