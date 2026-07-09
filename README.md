# Go Agentic AI Orchestrator

A event-driven orchestration system where a Go orchestrator consumes messages from Kafka, routes them to downstream Python microservices based on a YAML-defined DAG, and a React UI provides real-time observability and live config management.

---

## Architecture

```
React UI (port 5173)
    │
    ├── POST /api/v1/chat/  ──►  mock_service_1 (port 8001)
    │                                   │
    │                          publishes to Kafka
    │                          topic: mock-service-1
    │                                   │
    │                          Go Orchestrator (port 8000)
    │                          reads dag.yaml routing rules
    │                                   │
    │              ┌────────────────────┼────────────────────┐
    │              ▼                    ▼                    ▼
    │      mock-service-2       mock-service-3       mock-service-4
    │      mock_service_2       mock_service_3       mock_service_4
    │      (port 8002)          (port 8003)          (port 8004)
    │
    └── GET /api/v1/messages/  polls all services every 3s
```

---

## Project Structure

```
Go-Agentic-AI-Orchestrator/
├── main.go                  # Go orchestrator — Kafka consumer + HTTP API
├── go.mod / go.sum
├── dag.yaml                 # DAG routing config (hot-reloadable)
├── kafka_setup.py           # One-time Kafka topic creation script
│
├── mock_service_1/          # FastAPI — receives messages from UI, publishes to Kafka
├── mock_service_2/          # FastAPI — consumes mock-service-2 topic (CSV)
├── mock_service_3/          # FastAPI — consumes mock-service-3 topic (Excel)
├── mock_service_4/          # FastAPI — consumes mock-service-4 topic (PDF)
│
└── ui-react/                # React + Vite frontend
```

Each mock service follows this structure:
```
mock_service_N/
├── app/
│   ├── main.py              # FastAPI app, router registration, lifespan
│   ├── api/
│   │   ├── deps.py          # get_db() dependency
│   │   └── v1/
│   │       ├── items.py     # resource route handlers
│   │       ├── chat.py      # POST /api/v1/chat/ (service 1 only)
│   │       └── messages.py  # GET /api/v1/messages/ (services 2-4)
│   ├── core/
│   │   ├── config.py        # pydantic-settings (env vars)
│   │   ├── consumer.py      # Kafka consumer thread (services 2-4)
│   │   ├── kafka.py         # Kafka producer (service 1 only)
│   │   └── rbac.py          # ROLES, has_permission(), require_permission()
│   ├── db/
│   │   └── session.py       # SQLAlchemy session factory
│   ├── models/
│   │   └── schemas.py       # Pydantic request/response models
│   └── services/
│       └── items.py         # business logic stubs
├── tests/
├── main.py                  # uvicorn entry point
├── .env
└── requirements.txt
```

---

## DAG Config (`dag.yaml`)

Defines services and routing rules. Upload a new YAML from the React UI to hot-reload routing without restarting the orchestrator.

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

## Kafka Topics

| Topic            | Producer          | Consumer              |
|------------------|-------------------|-----------------------|
| mock-service-1   | mock_service_1    | Go Orchestrator       |
| mock-service-2   | Go Orchestrator   | mock_service_2        |
| mock-service-3   | Go Orchestrator   | mock_service_3        |
| mock-service-4   | Go Orchestrator   | mock_service_4        |
| orchestrator     | —                 | reserved              |

Topic settings: retention 1 day, 1 partition, replication factor 1.

---

## Prerequisites

- Go 1.21+
- Python 3.11+ with [uv](https://github.com/astral-sh/uv)
- Node.js 18+
- Apache Kafka running on `localhost:9092`

---

## Setup

### 1. Create Kafka Topics

Create these topics in Kafka UI (or CLI) with 1 partition, replication factor 1, retention 1 day:

```
mock-service-1
mock-service-2
mock-service-3
mock-service-4
orchestrator
```

### 2. Mock Services

```powershell
# Repeat for mock_service_2, mock_service_3, mock_service_4
cd mock_service_1
uv init --no-readme
uv add -r requirements.txt
```

### 3. React UI

```powershell
cd ui-react
npm install
```

---

## Running

Start each component in its own terminal:

```powershell
# mock_service_1 — port 8001
cd mock_service_1
uv run python main.py

# mock_service_2 — port 8002
cd mock_service_2
uv run python main.py

# mock_service_3 — port 8003
cd mock_service_3
uv run python main.py

# mock_service_4 — port 8004
cd mock_service_4
uv run python main.py

# Go Orchestrator — port 8000
go run main.go

# React UI — port 5173
cd ui-react
npm run dev
```

Open [http://localhost:5173](http://localhost:5173).

---

## React UI

| Feature | Description |
|---|---|
| **Send Message** | Select file type (CSV / Excel / PDF), type a message, hit Send or Enter |
| **Service Inbox** | 3-column view showing messages received by each downstream service, auto-refreshes every 3s |
| **DAG Config** | Visualises the current routing rules and registered services |
| **Upload YAML** | Upload a new `dag.yaml` to hot-reload routing rules without restarting the orchestrator |

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
| Mock Services | Python, FastAPI, kafka-python, pydantic-settings, SQLAlchemy |
| Message Bus | Apache Kafka |
| Frontend | React, Vite |
