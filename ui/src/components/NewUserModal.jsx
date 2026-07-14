import { useState } from 'react'
import { createUser } from '../api/users.js'

// "+ New user" modal — admin onboarding. The admin sets a temporary initial
// password (shared out-of-band); the user changes it via the account menu on
// first login.
export default function NewUserModal({ roles = [], onClose, onCreated }) {
  const [username, setUsername] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [role, setRole] = useState(roles.includes('viewer') ? 'viewer' : roles[0] || '')
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  async function submit() {
    setError('')
    setBusy(true)
    try {
      await createUser({ username: username.trim(), email: email.trim(), password, role })
      onCreated?.()
      onClose()
    } catch (err) {
      setError(err.message)
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="modal-backdrop" onClick={onClose}>
      <div className="modal" onClick={(e) => e.stopPropagation()}>
        <div className="modal-head">
          <h2>New user</h2>
          <button className="icon-btn" onClick={onClose} aria-label="Close">
            ✕
          </button>
        </div>

        <div className="modal-body">
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
              placeholder="jsmith"
              autoFocus
            />
          </label>

          <label className="field">
            <span className="field-label">Email</span>
            <input
              className="text-input"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="jsmith@example.com"
            />
          </label>

          <label className="field">
            <span className="field-label">Temporary password</span>
            <input
              className="text-input"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              autoComplete="new-password"
            />
          </label>

          <label className="field">
            <span className="field-label">Role</span>
            <select
              className="role-select"
              value={role}
              onChange={(e) => setRole(e.target.value)}
              style={{ width: '100%' }}
            >
              {roles.map((r) => (
                <option key={r} value={r}>
                  {r}
                </option>
              ))}
            </select>
          </label>
        </div>

        <div className="modal-foot">
          <span className="muted" style={{ fontSize: 12 }}>
            The user changes this password on first login.
          </span>
          <div className="actions">
            <button className="btn btn-sm" onClick={onClose}>
              Cancel
            </button>
            <button
              className="btn btn-primary btn-sm"
              disabled={busy || !username.trim() || !password || !role}
              onClick={submit}
            >
              {busy ? 'Creating…' : 'Create user'}
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
