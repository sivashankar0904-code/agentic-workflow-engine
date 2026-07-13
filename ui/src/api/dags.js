// DAG registry queries + lifecycle mutations (mock-backed).
// Mirrors the Control Plane endpoints in design.md §9: list (active-filtered),
// get by name, activate/deactivate.

import dagsData from '../mocks/dags.json';
import { mock } from './client.js';
import { dagToYaml } from './yaml.js';

// In-memory working copy so activate/deactivate mutations persist for the session
// (the real source of truth is the Control Plane's Postgres).
let registry = structuredClone(dagsData);

function summary(dag) {
  return {
    name: dag.name,
    active: dag.active,
    owner: dag.owner,
    allowedRoles: dag.allowedRoles,
    description: dag.description,
    updatedAt: dag.updatedAt,
    nodeCount: dag.nodes.length,
  };
}

// GET /dags?active=<bool>
export function listDags({ activeOnly = false } = {}) {
  const rows = registry
    .filter((d) => (activeOnly ? d.active : true))
    .map(summary);
  return mock(rows);
}

// GET /dags/{name} — full topology plus the reassembled YAML body.
export function getDag(name) {
  const dag = registry.find((d) => d.name === name);
  if (!dag) return mock(null);
  return mock({ ...dag, yaml: dagToYaml(dag) });
}

// POST activate / deactivate — flips the lifecycle flag, no YAML edit.
export function setDagActive(name, active) {
  const dag = registry.find((d) => d.name === name);
  if (dag) dag.active = active;
  return mock(dag ? summary(dag) : null);
}
