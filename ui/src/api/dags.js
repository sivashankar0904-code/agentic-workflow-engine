// DAG registry queries + lifecycle mutations, backed by the Control Plane.
// DAGs are addressed by their stable id (returned by the list); upload is
// name-addressed. See architecture.md / design.md §4-5.

import { load as yamlLoad } from 'js-yaml'
import { request } from './client.js'
import { dagToYaml } from './yaml.js'

// GET /dags?active=<bool> -> [{ id, name, active }]
export async function listDags({ activeOnly = false } = {}) {
  const { dags } = await request(`/dags?active=${activeOnly}`)
  return dags || []
}

// GET /dags/:id -> topology (parsed from the returned YAML) plus the raw YAML
// body and the id/name for the detail view. Returns null on 404.
export async function getDag(id) {
  let text
  try {
    text = await request(`/dags/${id}`, { raw: true })
  } catch (err) {
    if (err.status === 404) return null
    throw err
  }
  const parsed = yamlLoad(text) || {}
  const nodes = parsed.nodes || []
  return { id, name: String(id), nodes, yaml: text }
}

// POST /dags/name/:name -> create/replace a DAG from YAML, returns { id, name }.
export function uploadDag(name, yamlBody) {
  return request(`/dags/name/${encodeURIComponent(name)}`, {
    method: 'POST',
    raw: true,
    body: yamlBody,
  })
}

// POST activate/deactivate by id — flips the lifecycle flag, no YAML edit.
export function setDagActive(id, active) {
  return request(`/dags/${id}/${active ? 'activate' : 'deactivate'}`, {
    method: 'POST',
  })
}

// DELETE /dags/:id
export function deleteDag(id) {
  return request(`/dags/${id}`, { method: 'DELETE' })
}

// PATCH /dags/:id/roles -> replace the DAG's allowed roles (by role id).
export function setDagRoles(id, roleIds) {
  return request(`/dags/${id}/roles`, {
    method: 'PATCH',
    body: { roleIds },
  })
}

// Re-exported so the editor can serialize an edited node object back to YAML
// (used when the graph view needs a YAML string for a locally-built DAG).
export { dagToYaml }
