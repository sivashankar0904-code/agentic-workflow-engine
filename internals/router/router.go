package router

import (
	"fmt"
	"strings"

	"orchestrator/internals/dagconfig"
)

// Match returns the target topic for a message given the DAG's routing rules,
// or "" if no rule matches.
func Match(dag dagconfig.DAG, msg map[string]any) string {
	for _, rule := range dag.Routing.Rules {
		val, ok := msg[rule.Condition.Field]
		if !ok {
			continue
		}
		if strings.Contains(
			strings.ToLower(fmt.Sprintf("%v", val)),
			strings.ToLower(rule.Condition.Contains),
		) {
			return rule.Target
		}
	}
	return ""
}
