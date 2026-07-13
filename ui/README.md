# UI

React + Vite front door for the Agentic Workflow Engine — the UI tier described
in [`docs/design.md`](../docs/design.md). It has **no backend integration yet**:
every screen is driven by mock JSON in [`src/mocks/`](src/mocks/), shaped to match
the Control Plane / ingress responses the real app will call.

## Getting started

```bash
npm install
npm run dev
```

Open the URL printed in the terminal (default http://localhost:5173).

## What's here

- **Registry** (`/registry`) — governance table of all stored DAGs (active +
  inactive), with activate/deactivate. Selecting a DAG opens the YAML definition
  next to a live **graph** rendered from the `nodes[]`/`routes[]` schema
  (topological-rank layout, labeled edges, START/END markers, tool chips).
- **Run** (`/run`) — submit a workflow message to an active DAG and watch it flow
  node-to-node over the same graph, then see the result payload.
- **Users** (`/users`, admin-only) — RBAC console: roles, enable/disable, with the
  last-active-admin guard.

Light/dark theme toggle in the header; layout is responsive.

## Mock data

All data lives in [`src/mocks/`](src/mocks/) as JSON (nothing is hard-coded in
components):

| File | Contents |
|---|---|
| `dags.json` | The 4 DAGs from the architecture's RBAC table, including the two-sub-pipeline worked example (`support-triage`). |
| `users.json` | RBAC user list. |
| `session.json` | The current signed-in identity + roles. |
| `runs.json` | Sample message + node path + result per DAG. |

The API layer ([`src/api/`](src/api/)) reads these behind the same function
signatures the real HTTP client will expose, so wiring a backend later means
changing only [`src/api/client.js`](src/api/client.js) and the three `api/*.js`
modules — no page changes.

## File layout

Follows [`docs/design.md`](../docs/design.md) §10:

```
src/
  api/         client, dags, users, run, yaml serializer
  auth/        session + role helpers
  components/  AppShell, Sidebar, graph/ (DagGraph, layout, ToolChips)
  routes/      RegistryPage, RunPage, UsersPage
  mocks/       JSON fixtures
```

## Scripts

- `npm run dev` — start the dev server
- `npm run build` — build for production into `dist/`
- `npm run preview` — preview the production build locally

## Docker

Built and served via nginx from the repo-root `docker-compose.yml` (`ui` service,
port 3000). See [`Dockerfile`](Dockerfile) and [`nginx.conf`](nginx.conf).
