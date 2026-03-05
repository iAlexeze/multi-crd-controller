package controller

import (
	"context"
	"fmt"

	"github.com/ialexeze/kubernetes-crd-example/pkg/config/api/types/v1alpha1"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/logger"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
)

// runWorker is a long-running function that processes items from the queue
func (c *Controller) runWorker(ctx context.Context) {
	for c.processNextItem(ctx) {
	}
}

// processNextItem processes one item from the queue
func (c *Controller) processNextItem(ctx context.Context) bool {
	// Wait until there's an item or the queue is shut down
	key, shutdown := c.queue.Get()
	if shutdown {
		return false
	}

	// We call Done at the end of this function to mark the item as processed
	defer c.queue.Done(key)

	// Reconcile the item
	err := c.reconcile(ctx, key)
	if err != nil {
		// Handle error: requeue with rate limiting
		logger.Error().Err(err).Str("key", key).Msg("reconciliation failed")
		c.queue.AddRateLimited(key)
		return true
	}

	// Success: forget the item (remove from rate limiting history)
	c.queue.Forget(key)
	return true
}

// reconcile handles the actual business logic for a project
func (c *Controller) reconcile(ctx context.Context, key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return fmt.Errorf("invalid key format: %w", err)
	}

	// Get the object from the store
	obj, exists, err := c.informer.Store().GetByKey(key)
	if err != nil {
		return fmt.Errorf("failed to get object from store: %w", err)
	}

	if !exists {
		// Object was deleted
		logger.Info().Msgf("Project %s/%s has been deleted", namespace, name)
		// Perform any cleanup logic here
		return nil
	}

	// Type assert to project
	project, ok := obj.(*v1alpha1.Project)
	if !ok {
		return fmt.Errorf("expected *v1alpha1.Project, got %T", obj)
	}

	// Your reconciliation logic here
	logger.Info().Msgf("Reconciling project %s/%s (replicas: %d)",
		namespace, name, project.Spec.Replicas)

	// Example: Check if project needs finalizer
	if project.DeletionTimestamp != nil {
		// Handle deletion
		return c.handleDeletion(ctx, project)
	}

	// Normal reconciliation
	return c.reconcileNormal(ctx, project)
}

func (c *Controller) reconcileNormal(ctx context.Context, project *v1alpha1.Project) error {
	// Add your business logic here
	// e.g., ensure dependent resources exist, update status, etc.

	c.events.Recorder().Eventf(
		project,
		corev1.EventTypeNormal,
		"ProjectReconciled",
		"%s project reconciled successfully", project.Name,
	)
	logger.Debug().Msgf("Normal reconciliation for %s", project.Name)
	return nil
}

func (c *Controller) handleDeletion(ctx context.Context, project *v1alpha1.Project) error {
	logger.Info().Msgf("Handling deletion for %s", project.Name)
	// Add cleanup logic here
	// e.g., delete external resources, remove finalizers

	// Emit events
	c.events.Recorder().Eventf(
		project,
		corev1.EventTypeWarning,
		"ProjectDelete",
		"%s project deleted from %s namespace", project.Name, project.Namespace,
	)
	return nil
}
