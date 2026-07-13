package dag

import "gopkg.in/yaml.v3"

// FromYAML parses raw YAML bytes (as served by the control plane) into a DAG.
func FromYAML(raw []byte) (DAG, error) {
	var d DAG
	if err := yaml.Unmarshal(raw, &d); err != nil {
		return DAG{}, err
	}
	return d, nil
}
