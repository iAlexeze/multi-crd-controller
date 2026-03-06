// File: api/v1alpha1/groupversion_info.go
package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

type ManagedNamespaceConditionType string

const (
	Active      ManagedNamespaceConditionType = "Active"
	Ready       ManagedNamespaceConditionType = "Ready"
	Failed      ManagedNamespaceConditionType = "Failed"
	Error       ManagedNamespaceConditionType = "Error"
	Warning     ManagedNamespaceConditionType = "Warning"
	Progressing ManagedNamespaceConditionType = "Progressing"
	Reconciled  ManagedNamespaceConditionType = "Reconciled"

	Group   = "platform.ialexeze.io"
	Version = "v1alpha1"
	Kind    = "ManagedNamespace"
)

var (
	// GroupVersion is the group and version of your API
	GroupVersion = schema.GroupVersion{
		Group:   Group,
		Version: Version,
	}

	// API PATH
	APIPath = "/apis"

	// spec.names.plural
	NamePlural = "managednamespaces"

	// SchemeBuilder is used to add Go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds all types in this package to the scheme
	AddToScheme = SchemeBuilder.AddToScheme
)

// register known types
func init() {
	SchemeBuilder.Register(&ManagedNamespace{}, &ManagedNamespaceList{})
}
