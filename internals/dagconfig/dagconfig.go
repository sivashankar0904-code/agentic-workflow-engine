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

// NewStore loads the DAG named activeName from dir, creating dir if needed.
func NewStore(ctx context.Context, dir, activeName string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	s := &Store{dir: dir, name: activeName}
	if err := s.Load(ctx); err != nil {
		return nil, err
	}
	return s, nil
}

// Load (re)reads the active DAG file into the store.
func (s *Store) Load(ctx context.Context) error {
	s.mu.RLock()
	name := s.name
	s.mu.RUnlock()

	path := filepath.Join(s.dir, name)
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var d DAG
	if err := yaml.Unmarshal(data, &d); err != nil {
		return err
	}

	s.mu.Lock()
	s.dag = d
	s.mu.Unlock()
	log.Printf("DAG loaded from %s: source=%s rules=%d",
		path, d.Routing.Source, len(d.Routing.Rules))
	return nil
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
