package v1alpha1

import (
	"context"

	"github.com/ialexeze/multi-crd-controller/pkg/config/api/types/project/v1alpha1"
	"github.com/ialexeze/multi-crd-controller/pkg/config/domain"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/logger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

// Projects implements the project interface
func (p *projectClient) Projects(namespace string) domain.ProjectInterface {
	return &projectClient{
		name:           "projects",
		restClient:     p.restClient,
		namespace:      namespace,
		scheme:         p.scheme,
		parameterCodec: p.parameterCodec,
	}
}

// API Functions
func (p *projectClient) List(ctx context.Context, opts metav1.ListOptions) (*v1alpha1.ProjectList, error) {
	if p.restClient == nil {
		logger.Fatal().Msg("restClient is nil - check client initialization")
	}

	result := v1alpha1.ProjectList{}
	logger.Debug().Msgf("(BEFORE) projects: %v", len(result.Items))
	err := p.restClient.
		Get().
		Namespace(p.Namespace()).
		Resource(p.Name()).
		VersionedParams(&opts, p.parameterCodec).
		Do(ctx).
		Into(&result)

	logger.Debug().Msgf("(AFTER) projects: %v", len(result.Items))

	return &result, err
}

func (p *projectClient) Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1alpha1.Project, error) {
	result := v1alpha1.Project{}
	err := p.restClient.
		Get().
		Namespace(p.Namespace()).
		Resource(p.Name()).
		Name(name).
		VersionedParams(&opts, p.parameterCodec).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (p *projectClient) Create(ctx context.Context, project *v1alpha1.Project) (*v1alpha1.Project, error) {
	result := v1alpha1.Project{}
	err := p.restClient.
		Post().
		Namespace(p.Namespace()).
		Resource(p.Name()).
		Body(project).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (p *projectClient) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return p.restClient.
		Get().
		Namespace(p.Namespace()).
		Resource(p.Name()).
		VersionedParams(&opts, p.parameterCodec).
		Watch(ctx)
}
