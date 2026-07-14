import { useState } from 'react'
import { changePassword } from '../auth/session.js'

// Self-service password change (design.md §3 account menu). Requires the
// current password; the Control Plane verifies it server-side.
export default function ChangePasswordModal({ onClose }) {
  const [current, setCurrent] = useState('')
  const [next, setNext] = useState('')
  const [confirm, setConfirm] = useState('')
  const [error, setError] = useState('')
  const [done, setDone] = useState(false)
  const [busy, setBusy] = useState(false)

  const mismatch = next !== '' && confirm !== '' && next !== confirm

  async function submit() {
    setError('')
    setBusy(true)
    try {
      await changePassword(current, next)
      setDone(true)
    } catch (err) {
      setError(err.status === 401 ? 'Current password is incorrect.' : err.message)
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="modal-backdrop" onClick={onClose}>
      <div className="modal" onClick={(e) => e.stopPropagation()}>
        <div className="modal-head">
          <h2>Change password</h2>
          <button className="icon-btn" onClick={onClose} aria-label="Close">
            ✕
          </button>
        </div>

        <div className="modal-body">
          {done ? (
            <p className="muted" style={{ marginTop: 0 }}>
              Your password has been changed.
            </p>
          ) : (
            <>
              {error && (
                <div className="alert alert-error" role="alert">
                  {error}
                </div>
              )}
              <label className="field">
                <span className="field-label">Current password</span>
                <input
                  className="text-input"
                  type="password"
                  value={current}
                  onChange={(e) => setCurrent(e.target.value)}
                  autoFocus
                  autoComplete="current-password"
                />
              </label>
              <label className="field">
                <span className="field-label">New password</span>
                <input
                  className="text-input"
                  type="password"
                  value={next}
                  onChange={(e) => setNext(e.target.value)}
                  autoComplete="new-password"
                />
              </label>
              <label className="field">
                <span className="field-label">Confirm new password</span>
                <input
                  className="text-input"
                  type="password"
                  value={confirm}
                  onChange={(e) => setConfirm(e.target.value)}
                  autoComplete="new-password"
                />
              </label>
              {mismatch && (
                <span className="muted" style={{ fontSize: 12.5, color: 'var(--danger, #c33)' }}>
                  New passwords don’t match.
                </span>
              )}
            </>
          )}
        </div>

        <div className="modal-foot">
          <div className="actions" style={{ marginLeft: 'auto' }}>
            <button className="btn btn-sm" onClick={onClose}>
              {done ? 'Close' : 'Cancel'}
            </button>
            {!done && (
              <button
                className="btn btn-primary btn-sm"
                disabled={busy || !current || !next || mismatch}
                onClick={submit}
              >
                {busy ? 'Saving…' : 'Change password'}
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
