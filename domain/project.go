package domain

import (
	"context"

	projectv1alpha1 "github.com/ialexeze/multi-crd-controller/pkg/config/api/types/project/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
)

type ProjectInterface interface {
	List(ctx context.Context, opts metav1.ListOptions) (*projectv1alpha1.ProjectList, error)
	Get(ctx context.Context, name string, options metav1.GetOptions) (*projectv1alpha1.Project, error)
	Create(ctx context.Context, project *projectv1alpha1.Project) (*projectv1alpha1.Project, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Namespace() string
	// ...
}

type ProjectsV1Alpha1nterface interface {
	Projects(namespace string) ProjectInterface
	Namespace() string
	RestClient() rest.Interface
}
