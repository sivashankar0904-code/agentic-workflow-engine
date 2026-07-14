// Session / identity handling, backed by the Control Plane.
//
// login() authenticates and caches the bearer token plus the caller's identity
// and permission set (from GET /me) in localStorage. RBAC-gated UI reads the
// cached permission list via hasPermission(); pages hide actions the caller
// can't perform, but the Control Plane re-enforces server-side — the UI hiding
// is convenience, not security.

import { request, setToken, getToken } from '../api/client.js'

const SESSION_KEY = 'cp.session' // { username, role, permissions: [] }

function readSession() {
  try {
    return JSON.parse(localStorage.getItem(SESSION_KEY)) || null
  } catch {
    return null
  }
}

function writeSession(session) {
  if (session) localStorage.setItem(SESSION_KEY, JSON.stringify(session))
  else localStorage.removeItem(SESSION_KEY)
}

// login authenticates username/password, stores the token, then fetches the
// full identity + permission set via /me. Throws ApiError on bad credentials.
export async function login(username, password) {
  const { token } = await request('/login', {
    method: 'POST',
    body: { username, password },
  })
  setToken(token)

  const me = await request('/me')
  writeSession({ username: me.username, role: me.role, permissions: me.permissions || [] })
  return me
}

// logout clears the token and cached session.
export function logout() {
  setToken(null)
  writeSession(null)
}

// isAuthenticated reports whether a token and session are present.
export function isAuthenticated() {
  return Boolean(getToken() && readSession())
}

// getSession returns the cached identity, or null if not logged in.
export function getSession() {
  return readSession()
}

// hasPermission checks the cached permission set for the given dotted
// permission name (e.g. "control_plane.dag_registry.dag.create").
export function hasPermission(name) {
  const s = readSession()
  return Boolean(s && s.permissions.includes(name))
}

// hasRole / isAdmin are compatibility helpers over the cached role name, kept
// so existing call sites keep working. New gating should prefer hasPermission.
export function hasRole(role) {
  const s = readSession()
  return Boolean(s && s.role === role)
}

export function isAdmin() {
  return hasRole('admin')
}

// changePassword is the self-service password change (requires the current
// password), surfaced from the account menu.
export function changePassword(currentPassword, newPassword) {
  return request('/me/password', {
    method: 'POST',
    body: { currentPassword, newPassword },
  })
}
