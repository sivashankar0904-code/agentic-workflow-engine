package dagconfig

import (
	"context"
	"log"
	"sync"

	"gopkg.in/yaml.v3"

	"orchestrator/internals/s3bucket"
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

// ── Store: thread-safe, MinIO-backed holder for the DAG ───────────────────────
//
// The DAG lives as an object in the bucket. The bucket holds only DAG files,
// keyed by filename; the store tracks which key is currently active.

type Store struct {
	bucket *s3bucket.Client

	mu  sync.RWMutex
	key string // active DAG object key
	dag DAG
}

// NewStore loads the DAG at activeKey from the bucket.
func NewStore(ctx context.Context, bucket *s3bucket.Client, activeKey string) (*Store, error) {
	s := &Store{bucket: bucket, key: activeKey}
	if err := s.Load(ctx); err != nil {
		return nil, err
	}
	return s, nil
}

// Load (re)reads the active DAG object from the bucket into the store.
func (s *Store) Load(ctx context.Context) error {
	s.mu.RLock()
	key := s.key
	s.mu.RUnlock()

	data, err := s.bucket.Get(ctx, key)
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
	log.Printf("DAG loaded from s3://%s/%s: source=%s rules=%d",
		s.bucket.Bucket(), key, d.Routing.Source, len(d.Routing.Rules))
	return nil
}

// Get returns a snapshot of the current DAG.
func (s *Store) Get() DAG {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.dag
}

// ActiveKey returns the object key of the DAG currently in use.
func (s *Store) ActiveKey() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.key
}

// Replace validates raw YAML, uploads it to the bucket under name, and makes it
// the active DAG atomically.
func (s *Store) Replace(ctx context.Context, name string, raw []byte) (DAG, error) {
	var d DAG
	if err := yaml.Unmarshal(raw, &d); err != nil {
		return DAG{}, err
	}

	if err := s.bucket.Put(ctx, name, raw, "application/x-yaml"); err != nil {
		return DAG{}, err
	}

	s.mu.Lock()
	s.key = name
	s.dag = d
	s.mu.Unlock()
	log.Printf("DAG uploaded to s3://%s/%s and activated: source=%s rules=%d",
		s.bucket.Bucket(), name, d.Routing.Source, len(d.Routing.Rules))
	return d, nil
}
