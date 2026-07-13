import { useEffect, useState } from 'react'
import { listDags } from '../api/dags.js'

// Per-user resource allocation — which DAGs a user can access. UI only: the
// checklist starts from the user's current assignedDags but changes are local
// state, nothing is persisted (no backend yet).
export default function ManageAccessModal({ user, onClose }) {
  const [dags, setDags] = useState(null)
  const [selected, setSelected] = useState(() => new Set(user.assignedDags))

  useEffect(() => {
    listDags({ activeOnly: false }).then(setDags)
  }, [])

  function toggle(name) {
    setSelected((prev) => {
      const next = new Set(prev)
      next.has(name) ? next.delete(name) : next.add(name)
      return next
    })
  }

  return (
    <div className="modal-backdrop" onClick={onClose}>
      <div className="modal" onClick={(e) => e.stopPropagation()}>
        <div className="modal-head">
          <h2>Manage access — {user.username}</h2>
          <button className="icon-btn" onClick={onClose} aria-label="Close">
            ✕
          </button>
        </div>

        <div className="modal-body">
          <p className="muted" style={{ fontSize: 12.5, marginTop: 0 }}>
            Select which DAGs this user can view and run. Actual enforcement
            happens server-side; this list is a convenience mirror.
          </p>

          {dags === null ? (
            <div className="loading" style={{ minHeight: 140 }}>
              <div className="spinner" />
            </div>
          ) : (
            <div className="access-list">
              {dags.map((dag) => (
                <label key={dag.name} className="access-row">
                  <input
                    type="checkbox"
                    checked={selected.has(dag.name)}
                    onChange={() => toggle(dag.name)}
                  />
                  <span className="access-name">{dag.name}</span>
                  <span className={`badge ${dag.active ? 'on' : 'off'}`}>
                    <span className={`dot ${dag.active ? 'on' : 'off'}`} />
                    {dag.active ? 'Active' : 'Inactive'}
                  </span>
                </label>
              ))}
            </div>
          )}
        </div>

        <div className="modal-foot">
          <span className="muted" style={{ fontSize: 12 }}>
            {selected.size} of {dags?.length ?? 0} assigned · not saved
          </span>
          <div className="actions">
            <button className="btn btn-sm" onClick={onClose}>
              Cancel
            </button>
            <button className="btn btn-primary btn-sm" disabled>
              Save access
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
