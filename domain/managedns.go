package domain

import (
	"context"

	managednsv1alpha "github.com/ialexeze/multi-crd-controller/pkg/config/api/types/managedNamespace/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
)

type ManagedNamespaceInterface interface {
	List(ctx context.Context, opts metav1.ListOptions) (*managednsv1alpha.ManagedNamespaceList, error)
	Get(ctx context.Context, name string, options metav1.GetOptions) (*managednsv1alpha.ManagedNamespace, error)
	Create(ctx context.Context, project *managednsv1alpha.ManagedNamespace) (*managednsv1alpha.ManagedNamespace, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Namespace() string
	// ...
}

type ManagedNamespaceV1Alpha1nterface interface {
	ManagedNamespaces(namespace string) ManagedNamespaceInterface
	Namespace() string
	RestClient() rest.Interface
}
