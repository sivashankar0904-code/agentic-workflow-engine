// Role & permission admin queries + mutations, backed by the Control Plane.
// Roles are data (not a fixed enum); permissions are the
// service_name.feature_group.feature.action catalog registered per service.

import { request } from './client.js'

// GET /roles -> [{ id, name, permissions: [dottedName] }]
export async function listRoles() {
  const { roles } = await request('/roles')
  return roles || []
}

// GET /permissions -> [{ id, serviceKey, featureGroup, feature, action, name }]
export async function listPermissions() {
  const { permissions } = await request('/permissions')
  return permissions || []
}

// POST /roles
export function createRole(name) {
  return request('/roles', { method: 'POST', body: { name } })
}

// DELETE /roles/:name
export function deleteRole(name) {
  return request(`/roles/${encodeURIComponent(name)}`, { method: 'DELETE' })
}

// PUT /roles/:name/permissions — replace the role's permission set with exact
// leaf permission IDs (the matrix's per-checkbox save).
export function setRolePermissions(name, permissionIds) {
  return request(`/roles/${encodeURIComponent(name)}/permissions`, {
    method: 'PUT',
    body: { permissionIds },
  })
}

// PUT /roles/:name/permission-groups — bulk-grant an entire feature_group (or,
// with feature set, one feature) in a single call.
export function setRolePermissionGroup(name, { serviceId, featureGroup, feature }) {
  return request(`/roles/${encodeURIComponent(name)}/permission-groups`, {
    method: 'PUT',
    body: { serviceId, featureGroup, feature },
  })
}
