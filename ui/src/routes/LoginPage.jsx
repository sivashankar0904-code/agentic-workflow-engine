import { useState } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import { login } from '../auth/session.js'

// Login screen. On success, redirects to wherever the user was headed (or
// /registry). The Control Plane owns auth; this form just posts credentials.
export default function LoginPage() {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()

  const from = location.state?.from || '/registry'

  async function submit(e) {
    e.preventDefault()
    setError('')
    setBusy(true)
    try {
      await login(username.trim(), password)
      navigate(from, { replace: true })
    } catch (err) {
      setError(err.status === 401 ? 'Invalid username or password.' : err.message)
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="login-shell">
      <form className="card login-card" onSubmit={submit}>
        <div className="brand" style={{ marginBottom: 8 }}>
          <span className="brand-mark">⛓</span>
          <span>Agentic Workflow Engine</span>
        </div>
        <h1 style={{ fontSize: 18, margin: '0 0 12px' }}>Sign in</h1>

        {error && (
          <div className="alert alert-error" role="alert">
            {error}
          </div>
        )}

        <label className="field">
          <span className="field-label">Username</span>
          <input
            className="text-input"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            autoFocus
            autoComplete="username"
          />
        </label>

        <label className="field">
          <span className="field-label">Password</span>
          <input
            className="text-input"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            autoComplete="current-password"
          />
        </label>

        <button
          className="btn btn-primary"
          type="submit"
          disabled={busy || !username.trim() || !password}
          style={{ width: '100%', marginTop: 8 }}
        >
          {busy ? 'Signing in…' : 'Sign in'}
        </button>
      </form>
    </div>
  )
}
