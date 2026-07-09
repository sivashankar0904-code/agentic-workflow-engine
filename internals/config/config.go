package config

import (
	"log"
	"os"
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

// ── Store: thread-safe, hot-reloadable holder for the DAG ─────────────────────

type Store struct {
	path string
	mu   sync.RWMutex
	dag  DAG
}

// NewStore loads the DAG from path and returns a store guarding it.
func NewStore(path string) (*Store, error) {
	s := &Store{path: path}
	if err := s.Load(); err != nil {
		return nil, err
	}
	return s, nil
}

// Load (re)reads the DAG from disk into the store.
func (s *Store) Load() error {
	data, err := os.ReadFile(s.path)
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
	log.Printf("DAG loaded: source=%s rules=%d", d.Routing.Source, len(d.Routing.Rules))
	return nil
}

// Get returns a copy-safe snapshot of the current DAG.
func (s *Store) Get() DAG {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.dag
}

// Replace validates raw YAML, persists it to disk, and swaps it in atomically.
func (s *Store) Replace(raw []byte) (DAG, error) {
	var d DAG
	if err := yaml.Unmarshal(raw, &d); err != nil {
		return DAG{}, err
	}
	if err := os.WriteFile(s.path, raw, 0644); err != nil {
		return DAG{}, err
	}
	s.mu.Lock()
	s.dag = d
	s.mu.Unlock()
	log.Printf("DAG hot-reloaded: source=%s rules=%d", d.Routing.Source, len(d.Routing.Rules))
	return d, nil
}
