// User & RBAC admin queries + mutations, backed by the Control Plane
// (design.md §8). The server returns isActive: boolean and role: string.

import { request } from './client.js'

// GET /users -> [{ username, role, isActive, email }]
export async function listUsers() {
  const { users } = await request('/users')
  return users || []
}

// POST /users — onboard a new user with an admin-chosen temporary password.
export function createUser({ username, password, role, email }) {
  return request('/users', {
    method: 'POST',
    body: { username, password, role, email },
  })
}

// PATCH /users/:username/role
export function setUserRole(username, role) {
  return request(`/users/${encodeURIComponent(username)}/role`, {
    method: 'PATCH',
    body: { role },
  })
}

// PATCH /users/:username/active
export function setUserActive(username, isActive) {
  return request(`/users/${encodeURIComponent(username)}/active`, {
    method: 'PATCH',
    body: { isActive },
  })
}
