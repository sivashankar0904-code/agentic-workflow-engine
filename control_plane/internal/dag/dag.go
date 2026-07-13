package dag

// ── YAML wire schema ──────────────────────────────────────────────────────────
//
// This is the standardized DAG contract (see docs/architecture.md): one
// top-level nodes[] list, where each node declares its own outgoing routes[].
// name/active/RBAC are control-plane-side and never appear in this body.

type When struct {
	Field string `yaml:"field" json:"field"`
	Op    string `yaml:"op" json:"op"`
	Value string `yaml:"value" json:"value"`
}

type Route struct {
	When *When  `yaml:"when,omitempty" json:"when,omitempty"`
	To   string `yaml:"to" json:"to"`
}

type Node struct {
	Name   string   `yaml:"name" json:"name"`
	Topic  string   `yaml:"topic" json:"topic"`
	Host   string   `yaml:"host" json:"host"`
	Port   int      `yaml:"port" json:"port"`
	Entry  bool     `yaml:"entry,omitempty" json:"entry,omitempty"`
	Tools  []string `yaml:"tools,omitempty" json:"tools,omitempty"`
	Routes []Route  `yaml:"routes" json:"routes"`
}

type DAG struct {
	Nodes []Node `yaml:"nodes" json:"nodes"`
}
