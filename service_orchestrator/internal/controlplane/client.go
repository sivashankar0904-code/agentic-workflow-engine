package controlplane

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"orchestrator/internal/dag"
)

// Client pulls DAG definitions from the Control Plane Service. This engine is
// stateless with respect to definitions: it authenticates with a service key
// and reads only the DAGs the Control Plane serves it — it never stores them.
type Client struct {
	baseURL    string
	serviceKey string
	httpClient *http.Client
}

// New returns a Client pointed at the Control Plane's base URL
// (e.g. "http://localhost:9000"), authenticating with serviceKey via the
// X-Service-Key header.
func New(baseURL, serviceKey string) *Client {
	return &Client{
		baseURL:    baseURL,
		serviceKey: serviceKey,
		httpClient: &http.Client{},
	}
}

// get issues an authenticated GET to the Control Plane, attaching the
// service key.
func (c *Client) get(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Service-Key", c.serviceKey)
	return c.httpClient.Do(req)
}

// Flow identifies one active DAG the Control Plane serves: the stable id
// the engine fetches it by, plus its human name (the engine's routing key).
type Flow struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// ListActive returns the active DAGs the Control Plane serves this engine,
// as {id, name} pairs. The engine addresses each DAG's definition by id.
func (c *Client) ListActive() ([]Flow, error) {
	resp, err := c.get(c.baseURL + "/dags?active=true")
	if err != nil {
		return nil, fmt.Errorf("list active dags: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list active dags: unexpected status %d", resp.StatusCode)
	}

	var body struct {
		Dags []Flow `json:"dags"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("list active dags: %w", err)
	}
	return body.Dags, nil
}

// Get fetches the DAG with the given id from the Control Plane and parses
// its YAML body.
func (c *Client) Get(id int64) (dag.DAG, error) {
	url := fmt.Sprintf("%s/dags/%d", c.baseURL, id)
	resp, err := c.get(url)
	if err != nil {
		return dag.DAG{}, fmt.Errorf("get dag %d: %w", id, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return dag.DAG{}, fmt.Errorf("get dag %d: unexpected status %d", id, resp.StatusCode)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return dag.DAG{}, fmt.Errorf("get dag %d: %w", id, err)
	}
	return dag.FromYAML(raw)
}
