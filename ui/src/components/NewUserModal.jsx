import { useState } from 'react'

const ROLES = ['admin', 'editor', 'ops', 'viewer']

// "+ New user" modal — UI only, no submit logic/persistence yet.
export default function NewUserModal({ onClose }) {
  const [username, setUsername] = useState('')
  const [email, setEmail] = useState('')
  const [role, setRole] = useState('viewer')

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
            <span className="field-label">Role</span>
            <select
              className="role-select"
              value={role}
              onChange={(e) => setRole(e.target.value)}
              style={{ width: '100%' }}
            >
              {ROLES.map((r) => (
                <option key={r} value={r}>
                  {r}
                </option>
              ))}
            </select>
          </label>
        </div>

        <div className="modal-foot">
          <span className="muted" style={{ fontSize: 12 }}>
            No backend yet — creation isn't wired up.
          </span>
          <div className="actions">
            <button className="btn btn-sm" onClick={onClose}>
              Cancel
            </button>
            <button
              className="btn btn-primary btn-sm"
              disabled={!username.trim() || !email.trim()}
            >
              Create user
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
