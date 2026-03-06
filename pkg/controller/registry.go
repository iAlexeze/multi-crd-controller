// pkg/config/pkg/controller/registry.go
package controller

import (
	"github.com/ialexeze/multi-crd-controller/pkg/config/domain"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/informer"
)

type CRDInfo struct {
	Group   string
	Version string
	Kind    string
	APIPath string // optional, but you already have it
}

type RegistryEntry struct {
	CRD        CRDInfo
	Informer   informer.InformerComponents
	Reconciler domain.Reconciler
}

type ResourceRegistry struct {
	entries map[domain.Resource]RegistryEntry
}

func NewRegistry() *ResourceRegistry {
	return &ResourceRegistry{
		entries: make(map[domain.Resource]RegistryEntry),
	}
}

func (r *ResourceRegistry) Register(
	resource domain.Resource,
	crd CRDInfo,
	inf informer.InformerComponents,
	rec domain.Reconciler,
) {
	r.entries[resource] = RegistryEntry{
		CRD:        crd,
		Informer:   inf,
		Reconciler: rec,
	}
}

func (r *ResourceRegistry) Entries() map[domain.Resource]RegistryEntry {
	return r.entries
}
