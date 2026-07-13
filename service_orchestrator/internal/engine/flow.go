package engine

import (
	"fmt"
	"strings"

	"orchestrator/internal/dag"
)

// Route is a resolved edge: an optional predicate and the node it leads to.
type Route struct {
	When *dag.When
	To   *dag.Node
}

// Flow is the routing table built from one DAG: for each node, its outgoing
// routes with `to` already resolved to the target node.
type Flow struct {
	Name      string
	Nodes     map[string]*dag.Node
	Routes    map[string][]Route
	EntryNode string
}

// Build compiles a DAG into a Flow routing table, resolving each route's `to`
// (a node name) against the DAG's own nodes.
func Build(name string, d dag.DAG) (*Flow, error) {
	f := &Flow{
		Name:   name,
		Nodes:  make(map[string]*dag.Node, len(d.Nodes)),
		Routes: make(map[string][]Route, len(d.Nodes)),
	}

	for i := range d.Nodes {
		n := &d.Nodes[i]
		f.Nodes[n.Name] = n
		if n.Entry {
			if f.EntryNode != "" {
				return nil, fmt.Errorf("dag %q: multiple entry nodes (%q, %q)", name, f.EntryNode, n.Name)
			}
			f.EntryNode = n.Name
		}
	}

	for i := range d.Nodes {
		n := &d.Nodes[i]
		for _, r := range n.Routes {
			target, ok := f.Nodes[r.To]
			if !ok {
				return nil, fmt.Errorf("dag %q: node %q routes to unknown node %q", name, n.Name, r.To)
			}
			f.Routes[n.Name] = append(f.Routes[n.Name], Route{When: r.When, To: target})
		}
	}

	return f, nil
}

// Next evaluates node's routes against message, returning the first matching
// target (an unconditional route, i.e. When == nil, always matches). Returns
// nil if no route matches — the message terminates at node.
func (f *Flow) Next(nodeName string, message map[string]any) *dag.Node {
	for _, route := range f.Routes[nodeName] {
		if route.When == nil || matches(route.When, message) {
			return route.To
		}
	}
	return nil
}

func matches(w *dag.When, message map[string]any) bool {
	v, ok := message[w.Field]
	if !ok {
		return false
	}
	s, ok := v.(string)
	if !ok {
		return false
	}
	switch w.Op {
	case "equals":
		return s == w.Value
	case "contains":
		return strings.Contains(s, w.Value)
	default:
		return false
	}
}
