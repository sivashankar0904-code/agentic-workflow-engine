# Go Agentic AI Orchestrator

A Go service that stores workflow DAGs (services + routing rules) in Postgres as a
first-class graph, and serves them over an HTTP API. DAGs are uploaded and served
as YAML; internally they are persisted relationally (registry + nodes + edges) so
the graph can be queried, traversed, and validated in SQL. Multiple named DAGs
coexist.

---

## Architecture

```
                    Go Orchestrator (port 8000)
                              │
                    ┌─────────┴─────────┐
                    ▼                   ▼
              HTTP API            Postgres graph store
       (/health, /config, /dags)   dag_registry
                                    ├── nodes  (services / vertices)
                                    └── edges  (routing rules / edges)
```

YAML is only the wire format. On `POST`, the YAML is parsed and written into the
graph tables in a transaction; on `GET`, the graph is reassembled from nodes/edges
and rendered back to YAML.

---

## Project Structure

```
agentic-workflow-engine/
├── service_orchestrator/
│   ├── cmd/orchestrator/    # main.go — entrypoint: opens pg pool, wires deps
│   ├── schemas/             # SQL table definitions (apply manually)
│   │   ├── 01_dag_registry.sql
│   │   ├── 02_nodes.sql
│   │   └── 03_edges.sql
│   ├── mock/dag.yaml        # example DAG (upload payload)
│   ├── internal/
│   │   ├── config/          # env config (DATABASE_URL)
│   │   ├── dag/              # domain: types, YAML transform, Postgres Store (CRUD)
│   │   └── server/           # gin engine, handlers, middleware
│   ├── go.mod
│   ├── go.sum
│   └── Dockerfile
└── docker-compose.yml       # postgres + orchestrator
```

---

## Graph Model

| Table | Role |
|---|---|
| `dag_registry` | one row per named DAG |
| `nodes` | services — graph vertices, unique by `(dag_id, name)` |
| `edges` | routing rules — graph edges, referencing nodes by id (FK) |

Because edges reference nodes by foreign key, the graph is enforced by the DB and
queryable — e.g. reachability from a source node via a recursive CTE, cycle
detection, orphan detection.

The routing `source` and each rule `target` are topic strings; they are resolved
to nodes by matching `nodes.topic`.

---

## DAG Payload (`dag.yaml`)

```yaml
services:
  - name: mock_service_1
    host: localhost
    port: 8001
    topic: mock-service-1

routing:
  source: mock-service-1       # source node (by topic)
  rules:
    - condition:
        field: message         # JSON field to match on
        contains: "[CSV]"
      target: mock-service-2   # target node (by topic)

    - condition:
        field: message
        contains: "[Excel]"
      target: mock-service-3

    - condition:
        field: message
        contains: "[PDF]"
      target: mock-service-4
```

---

## Prerequisites

- Go 1.21+
- PostgreSQL (schema files in `service_orchestrator/schemas/` applied manually, in numeric order)

---

## Running

### Locally

```powershell
# Apply schemas once (registry -> nodes -> edges order matters):
psql "$env:DATABASE_URL" -f service_orchestrator/schemas/01_dag_registry.sql
psql "$env:DATABASE_URL" -f service_orchestrator/schemas/02_nodes.sql
psql "$env:DATABASE_URL" -f service_orchestrator/schemas/03_edges.sql

cd service_orchestrator
go run ./cmd/orchestrator
```

The service listens on port 8000. Connection string comes from `DATABASE_URL`
(default `postgres://postgres:postgres@localhost:5432/orchestrator?sslmode=disable`).

### With Docker Compose

Compose starts Postgres (applying `schemas/` on first init) and the orchestrator:

```powershell
docker compose up --build
```

---

## API

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | Health check |
| `GET` | `/dags` | List all stored DAG names |
| `GET` | `/config?name=<dag>` | Returns the named DAG as YAML |
| `POST` | `/config?name=<dag>` | Upload YAML; creates or replaces the named DAG |
| `DELETE` | `/config?name=<dag>` | Delete the named DAG (nodes/edges cascade) |

---

## Tech Stack

| Layer | Technology |
|---|---|
| Service | Go, [gin](https://github.com/gin-gonic/gin), gopkg.in/yaml.v3 |
| DAG storage | PostgreSQL via [pgx](https://github.com/jackc/pgx) (graph: registry + nodes + edges) |
