# Go Agentic AI Orchestrator

A Go service that stores and serves a YAML-defined DAG (services + routing rules)
over an HTTP API. DAG files live on the local filesystem and can be uploaded /
hot-swapped via the API without restarting the service.

---

## Architecture

```
                    Go Orchestrator (port 8000)
                              │
                    ┌─────────┴─────────┐
                    ▼                   ▼
              HTTP API            DAG YAML store
          (/health, /config)     (local filesystem)
```

The service exposes the current DAG as JSON and accepts new YAML via `POST /config`,
persisting each upload to `DAG_DIR` and making it the active config in memory.

---

## Project Structure

```
agentic-workflow-engine/
├── main.go                  # HTTP API entrypoint
├── go.mod / go.sum
├── dags/dag.yaml            # DAG config (hot-reloadable, stored locally)
├── Dockerfile               # Container image
├── docker-compose.yml       # Runs the service with a mounted dags/ volume
└── internals/               # config, dagconfig store, and HTTP server
```

---

## DAG Config (`dag.yaml`)

Defines services and routing rules. `POST /config` a new YAML to hot-swap the
active config without restarting the service.

```yaml
services:
  - name: mock_service_1
    host: localhost
    port: 8001
    topic: mock-service-1

routing:
  source: mock-service-1       # source identifier
  rules:
    - condition:
        field: message         # JSON field to match on
        contains: "[CSV]"
      target: mock-service-2   # target to route to

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

---

## Running

### Locally

```powershell
go run main.go
```

The service listens on port 8000. DAG files are read from / written to `DAG_DIR`
(default `./dags`).

### With Docker Compose

DAG YAML files are stored on the local filesystem and mounted into the container
(see `DAG_DIR`). Copy `.env.example` to `.env`, then:

```powershell
docker compose up --build
```

---

## API

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | Health check |
| `GET` | `/config` | Returns current DAG as JSON |
| `POST` | `/config` | Upload new YAML config; stores it and makes it active. Optional `?name=<file>` to name the stored file |

---

## Tech Stack

| Layer | Technology |
|---|---|
| Service | Go, [gin](https://github.com/gin-gonic/gin), gopkg.in/yaml.v3 |
| DAG storage | Local filesystem |
