package engine

import (
	"fmt"
	"log"
	"sync"

	"orchestrator/internal/controlplane"
)

// Registry holds one compiled Flow per active DAG this engine is authorized
// to run. It has no database of its own — Refresh is the only way flows enter
// or leave the registry, driven entirely by what the Control Plane currently
// serves as active.
type Registry struct {
	cp *controlplane.Client

	mu    sync.RWMutex
	flows map[string]*Flow
}

// NewRegistry returns an empty Registry that will pull DAGs via cp.
func NewRegistry(cp *controlplane.Client) *Registry {
	return &Registry{cp: cp, flows: make(map[string]*Flow)}
}

// Refresh re-lists the Control Plane's active DAGs, builds any that are new or
// changed, and drops any no longer active. Inactive or unauthorized flows are
// never built.
func (r *Registry) Refresh() error {
	names, err := r.cp.ListActive()
	if err != nil {
		return fmt.Errorf("refresh: %w", err)
	}

	built := make(map[string]*Flow, len(names))
	for _, name := range names {
		d, err := r.cp.Get(name)
		if err != nil {
			log.Printf("refresh: skipping %q: %v", name, err)
			continue
		}
		flow, err := Build(name, d)
		if err != nil {
			log.Printf("refresh: skipping %q: %v", name, err)
			continue
		}
		built[name] = flow
	}

	r.mu.Lock()
	r.flows = built
	r.mu.Unlock()
	return nil
}

// Get returns the live Flow for name, or false if it isn't currently active
// (or wasn't authorized/buildable).
func (r *Registry) Get(name string) (*Flow, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	f, ok := r.flows[name]
	return f, ok
}

// Names returns the names of all currently built flows.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.flows))
	for name := range r.flows {
		names = append(names, name)
	}
	return names
}
