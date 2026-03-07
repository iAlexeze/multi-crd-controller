package domain

import (
	"context"
	"time"
	// "k8s.io/client-go/tools/cache"
)

type Component interface {

	// Start() starts the compnent
	Start(context.Context) error

	// Shutdown() shuts down the component gracefully
	Shutdown(context.Context)

	// Name() returns the name of the component
	Name() string
}

// To implement, replace 'componentName' with the appropriate component
// var _ domain.Component = (*componentName)(nil)
// func (c *componentName) Start(ctx context.Context) error {}
// func (c *componentName) Shutdown(ctx context.Context) {}
// func (c *componentName) Name() string {}

type Reconciler interface {
	// Reconcile handles the actual business logic for a resource
	Reconcile(ctx context.Context, key string) error

	// Resource() returns the resource name or kind for the reconciler
	Resource() Resource
}

// Define resource types
// Add more resources as needed for uniformity
type Resource string

const (
	DefaultResync                             = 30 * time.Second
	DefaultNamespace                          = "default"
	ProjectResource                  Resource = "Project"
	ProjectInformerResource          Resource = "ProjectInformer"
	ManagedNamespaceResource         Resource = "ManagedNamespace"
	ManagedNamespaceInformerResource Resource = "ManagedNamespaceInformer"
)
