package dag

// ── YAML wire schema ──────────────────────────────────────────────────────────
//
// This is the format DAGs are uploaded and served in. Internally the graph is
// stored relationally (dag_registry / nodes / edges); see repo.go. yaml.go
// converts between this schema and the persisted graph.

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
