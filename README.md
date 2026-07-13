# Agentic Workflow Engine

A control-plane/data-plane split for running workflow DAGs: the **Control
Plane Service** is the single source of truth for DAG definitions, RBAC, and
active/inactive lifecycle; **`service_orchestrator`** (and, separately,
`agentic_orchestrator`) are stateless execution engines that pull the DAGs
they're authorized for and run them. See [docs/architecture.md](docs/architecture.md)
for the full picture.

---

## Architecture

```
        Control Plane Service (Go, port 9000)
                      │
          ┌───────────┴───────────┐
          ▼                       ▼
     HTTP API              Postgres graph store
(/health, /dags)            dag_registry (+ active)
                             ├── nodes  (graph vertices)
                             └── edges  (routes[])
                      │
         authenticated pull (GET /dags?active=true, GET /dags/{name})
                      │
                      ▼
     service_orchestrator (Go, port 8000)
     execution engine — no database of its own
```

DAGs are uploaded and served as YAML by the Control Plane; internally they are
persisted relationally (registry + nodes + edges) so the graph can be queried,
traversed, and validated in SQL. `service_orchestrator` never touches that
storage — it authenticates to the Control Plane, pulls the active DAGs it's
allowed to see, and builds a routing table per flow.

---

## Project Structure

```
agentic-workflow-engine/
├── control_plane/           # DAG registry — the only service with a database
│   ├── cmd/controlplane/    # main.go — entrypoint: opens pg pool, wires deps
│   ├── schemas/             # SQL table definitions (apply manually)
│   │   ├── 01_dag_registry.sql
│   │   ├── 02_nodes.sql
│   │   └── 03_edges.sql
│   ├── mock/dag.yaml        # example DAG (upload payload)
│   ├── internal/
│   │   ├── config/          # env config (DATABASE_URL)
│   │   ├── dag/             # domain: types, YAML transform, Postgres Store (CRUD)
│   │   └── server/          # gin engine, handlers, middleware
│   ├── go.mod / go.sum
│   └── Dockerfile
├── service_orchestrator/    # execution engine — service routing runtime
│   ├── cmd/orchestrator/    # main.go — entrypoint: pulls DAGs, builds registry
│   ├── internal/
│   │   ├── config/          # env config (CONTROL_PLANE_URL)
│   │   ├── dag/             # domain: YAML wire types (read-only client side)
│   │   ├── controlplane/    # HTTP client for the Control Plane's DAG API
│   │   ├── engine/          # Flow (routing table) + Registry (active flows)
│   │   └── server/          # gin engine, handlers, middleware
│   ├── go.mod / go.sum
│   └── Dockerfile
└── docker-compose.yml       # postgres + control_plane + orchestrator + ui
```

---

## The DAG Contract

```yaml
nodes:
  - name: ingest
    topic: ingest
    host: localhost
    port: 8001
    entry: true
    routes:
      - when: {field: message, op: contains, value: "[CSV]"}
        to: csv
      - when: {field: message, op: contains, value: "[PDF]"}
        to: pdf

  - name: csv
    topic: csv
    host: localhost
    port: 8002
    routes:
      - to: archive

  - name: pdf
    topic: pdf
    host: localhost
    port: 8003
    routes:
      - to: archive

  - name: archive
    topic: archive
    host: localhost
    port: 8004
    routes: []
```

One top-level `nodes[]` list; each node declares its own outgoing `routes[]`.
See [docs/architecture.md](docs/architecture.md) for the full schema
invariants (exactly one `entry`, terminal nodes, `tools[]`, named `op`s).

---

## Running

### With Docker Compose

```powershell
docker compose up --build
```

Starts Postgres (schema applied on first init), the Control Plane on `:9000`,
`service_orchestrator` on `:8000`, and the UI on `:3000`.

### Locally

```powershell
# Apply schemas once (registry -> nodes -> edges order matters):
psql "$env:DATABASE_URL" -f control_plane/schemas/01_dag_registry.sql
psql "$env:DATABASE_URL" -f control_plane/schemas/02_nodes.sql
psql "$env:DATABASE_URL" -f control_plane/schemas/03_edges.sql

cd control_plane
go run ./cmd/controlplane   # listens on :9000, needs DATABASE_URL

cd ../service_orchestrator
go run ./cmd/orchestrator   # listens on :8000, needs CONTROL_PLANE_URL
```

---

## API — Control Plane (`:9000`)

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | Health check |
| `GET` | `/dags` | List all stored DAG names (`?active=true` to filter) |
| `GET` | `/dags/{name}` | Returns the named DAG as YAML |
| `POST` | `/dags/{name}` | Upload YAML; creates or replaces the named DAG (inactive by default) |
| `DELETE` | `/dags/{name}` | Delete the named DAG (nodes/edges cascade) |
| `POST` | `/dags/{name}/activate` | Mark the DAG active — served to execution engines |
| `POST` | `/dags/{name}/deactivate` | Mark the DAG inactive — retained, not served |

## API — `service_orchestrator` (`:8000`)

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | Health check |
| `GET` | `/flows` | Names of flows currently built from the Control Plane's active DAGs |

---

## Tech Stack

| Layer | Technology |
|---|---|
| Control Plane | Go, [gin](https://github.com/gin-gonic/gin), gopkg.in/yaml.v3, [pgx](https://github.com/jackc/pgx) |
| Execution engine | Go, gin, gopkg.in/yaml.v3 (no database) |
| DAG storage | PostgreSQL, owned exclusively by the Control Plane |
