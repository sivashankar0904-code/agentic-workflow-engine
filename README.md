# Go Agentic AI Orchestrator

An event-driven orchestration system where a Go orchestrator consumes messages from Kafka and routes them to downstream services based on a YAML-defined DAG. The routing config is hot-reloadable via the HTTP API without restarting the orchestrator.

---

## Architecture

```
                    Kafka topic: mock-service-1
                              │
                    Go Orchestrator (port 8000)
                    reads dag.yaml routing rules
                              │
        ┌─────────────────────┼─────────────────────┐
        ▼                     ▼                     ▼
  mock-service-2        mock-service-3        mock-service-4
   (Kafka topic)         (Kafka topic)         (Kafka topic)
```

The orchestrator consumes from the `source` topic, evaluates each message against
the routing rules in `dag.yaml`, and republishes matching messages to the target topic.

---

## Project Structure

```
agentic-workflow-engine/
├── main.go                  # Go orchestrator — Kafka consumer + HTTP API
├── go.mod / go.sum
├── dags/dag.yaml            # DAG routing config (hot-reloadable)
├── Dockerfile               # Orchestrator container image
├── docker-compose.yml       # Runs the orchestrator on the shared network
└── internals/               # config, kafka client, and orchestrator internals
```

---

## DAG Config (`dag.yaml`)

Defines services and routing rules. `POST /config` a new YAML to hot-reload routing
without restarting the orchestrator.

```yaml
services:
  - name: mock_service_1
    host: localhost
    port: 8001
    topic: mock-service-1

routing:
  source: mock-service-1       # topic the orchestrator consumes from
  rules:
    - condition:
        field: message         # JSON field to match on
        contains: "[CSV]"
      target: mock-service-2   # Kafka topic to route to

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
- Apache Kafka running on `localhost:9092`

---

## Running

### Locally

```powershell
go run main.go
```

The orchestrator listens on port 8000.

### With Docker Compose

Kafka and Postgres are expected to already run in the external
`local-docker_default` network. DAG YAML files are stored on the local
filesystem (see `DAG_DIR`). Copy `.env.example` to `.env`, then:

```powershell
docker compose up --build
```

---

## Go Orchestrator API

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | Health check |
| `GET` | `/config` | Returns current DAG as JSON |
| `POST` | `/config` | Upload new YAML config, hot-reloads routing immediately |

---

## Tech Stack

| Layer | Technology |
|---|---|
| Orchestrator | Go, [franz-go](https://github.com/twmb/franz-go), gopkg.in/yaml.v3 |
| Message Bus | Apache Kafka |
