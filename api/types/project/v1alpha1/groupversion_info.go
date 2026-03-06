// Written without controller-gen
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	Group   = "platform.ialexeze.io"
	Version = "v1alpha1"
	Kind    = "Project"
)

var (
	GroupVersion = schema.GroupVersion{
		Group:   Group,
		Version: Version,
	}

	// API PATH
	APIPath = "/apis"

	// spec.names.plural
	NamePlural = "projects"

	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme
)

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(GroupVersion,
		&Project{},
		&ProjectList{},
	)

	metav1.AddToGroupVersion(scheme, GroupVersion)
	return nil
}
