package controlplane

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"orchestrator/internal/dag"
)

// Client pulls DAG definitions from the Control Plane Service. This engine is
// stateless with respect to definitions: it authenticates (future work) and
// reads only the DAGs the Control Plane serves it — it never stores them.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// New returns a Client pointed at the Control Plane's base URL
// (e.g. "http://localhost:9000").
func New(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

// ListActive returns the names of all DAGs currently active on the Control Plane.
func (c *Client) ListActive() ([]string, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/dags?active=true")
	if err != nil {
		return nil, fmt.Errorf("list active dags: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list active dags: unexpected status %d", resp.StatusCode)
	}

	var body struct {
		Dags []string `json:"dags"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("list active dags: %w", err)
	}
	return body.Dags, nil
}

// Get fetches the named DAG's YAML body from the Control Plane and parses it.
func (c *Client) Get(name string) (dag.DAG, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/dags/" + name)
	if err != nil {
		return dag.DAG{}, fmt.Errorf("get dag %q: %w", name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return dag.DAG{}, fmt.Errorf("get dag %q: unexpected status %d", name, resp.StatusCode)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return dag.DAG{}, fmt.Errorf("get dag %q: %w", name, err)
	}
	return dag.FromYAML(raw)
}
