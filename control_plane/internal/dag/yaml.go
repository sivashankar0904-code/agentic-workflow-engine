package dag

import "gopkg.in/yaml.v3"

// FromYAML parses raw YAML bytes into a DAG.
func FromYAML(raw []byte) (DAG, error) {
	var d DAG
	if err := yaml.Unmarshal(raw, &d); err != nil {
		return DAG{}, err
	}
	return d, nil
}

// ToYAML serializes a DAG back to YAML. Used to serve DAGs reassembled from the
// relational graph (nodes/edges) on read.
func ToYAML(d DAG) ([]byte, error) {
	return yaml.Marshal(d)
}
