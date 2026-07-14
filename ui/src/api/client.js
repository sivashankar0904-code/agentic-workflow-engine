// API client for the Control Plane (:9000).
//
// A thin fetch wrapper that attaches the bearer token (stored by auth/session.js
// after login) to every request and unwraps JSON / surfaces errors. The DAG
// run console still uses the mock() helper below, since the execution ingress
// it targets doesn't exist yet.

const TOKEN_KEY = 'cp.token'

// Base URLs live here so pages never hard-code them. In dev the UI (Vite,
// :5173) calls the Control Plane directly; its permissive CORS allows it.
export const BASE_URLS = {
  controlPlane: 'http://localhost:9000',
  ingress: 'http://localhost:9000/ingress',
}

export function getToken() {
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(token) {
  if (token) localStorage.setItem(TOKEN_KEY, token)
  else localStorage.removeItem(TOKEN_KEY)
}

// ApiError carries the HTTP status plus the server's error message, so callers
// can branch on status (e.g. 401 -> redirect to login).
export class ApiError extends Error {
  constructor(status, message) {
    super(message)
    this.status = status
  }
}

// request performs an authenticated call to the Control Plane. `path` is
// relative to the control-plane base URL (e.g. "/dags"). Options:
//   - method, body (auto-JSON-encoded unless raw), headers
//   - raw: true  -> return the response text instead of parsing JSON
//     (used for GET /dags/:id, which returns YAML)
export async function request(path, { method = 'GET', body, raw = false, headers = {} } = {}) {
  const opts = { method, headers: { ...headers } }
  const token = getToken()
  if (token) opts.headers['Authorization'] = `Bearer ${token}`

  if (body !== undefined) {
    if (raw) {
      opts.body = body
      opts.headers['Content-Type'] = opts.headers['Content-Type'] || 'application/x-yaml'
    } else {
      opts.body = JSON.stringify(body)
      opts.headers['Content-Type'] = 'application/json'
    }
  }

  const res = await fetch(`${BASE_URLS.controlPlane}${path}`, opts)

  if (!res.ok) {
    let message = res.statusText
    try {
      const data = await res.json()
      message = data.error || message
    } catch {
      // non-JSON error body; keep statusText
    }
    throw new ApiError(res.status, message)
  }

  if (raw) return res.text()
  if (res.status === 204) return null
  const text = await res.text()
  return text ? JSON.parse(text) : null
}

// mock resolves from bundled JSON after a small delay — retained only for the
// run console (api/run.js), whose backend ingress isn't built yet.
const LATENCY_MS = 220

export function mock(data) {
  return new Promise((resolve) => {
    setTimeout(() => resolve(structuredClone(data)), LATENCY_MS)
  })
}
