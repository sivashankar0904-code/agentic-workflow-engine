package dagconfig

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// ── YAML schema ───────────────────────────────────────────────────────────────

type Condition struct {
	Field    string `yaml:"field" json:"field"`
	Contains string `yaml:"contains" json:"contains"`
}

type Rule struct {
	Condition Condition `yaml:"condition" json:"condition"`
	Target    string    `yaml:"target" json:"target"`
}

type Routing struct {
	Source string `yaml:"source" json:"source"`
	Rules  []Rule `yaml:"rules" json:"rules"`
}

type Service struct {
	Name  string `yaml:"name" json:"name"`
	Host  string `yaml:"host" json:"host"`
	Port  int    `yaml:"port" json:"port"`
	Topic string `yaml:"topic" json:"topic"`
}

type DAG struct {
	Services []Service `yaml:"services" json:"services"`
	Routing  Routing   `yaml:"routing" json:"routing"`
}

// ── Store: thread-safe, filesystem-backed holder for the DAG ──────────────────
//
// The DAG lives as a YAML file in a local directory. The directory holds only
// DAG files, keyed by filename; the store tracks which file is currently active.

type Store struct {
	dir string // directory holding DAG YAML files

	mu   sync.RWMutex
	name string // active DAG file name
	dag  DAG
}

// NewStore returns a store bound to dir, creating dir if needed. No DAG is
// loaded at startup; the store holds an empty DAG until one is uploaded via
// Replace. With no active DAG there is simply no workflow.
func NewStore(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &Store{dir: dir}, nil
}

// Active reports whether a DAG is currently loaded.
func (s *Store) Active() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.name != ""
}

// Get returns a snapshot of the current DAG.
func (s *Store) Get() DAG {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.dag
}

// ActiveKey returns the file name of the DAG currently in use.
func (s *Store) ActiveKey() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.name
}

// Replace validates raw YAML, writes it to the directory under name, and makes
// it the active DAG atomically.
func (s *Store) Replace(ctx context.Context, name string, raw []byte) (DAG, error) {
	var d DAG
	if err := yaml.Unmarshal(raw, &d); err != nil {
		return DAG{}, err
	}

	path := filepath.Join(s.dir, name)
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		return DAG{}, err
	}

	s.mu.Lock()
	s.name = name
	s.dag = d
	s.mu.Unlock()
	log.Printf("DAG written to %s and activated: source=%s rules=%d",
		path, d.Routing.Source, len(d.Routing.Rules))
	return d, nil
}
