import { useEffect, useState } from 'react'
import { listDags } from '../api/dags.js'

// Role-level resource allocation: a role × DAG matrix mirroring each DAG's
// allowedRoles[] (design.md §4, architecture.md's dag_registry RBAC table).
// Roles themselves are customizable — add a new role (appends a column) or
// remove one (drops its column and grants). UI only — everything here is
// local state; nothing is persisted (no backend yet).
export default function RolesAccessPanel() {
  const [dags, setDags] = useState(null)
  const [roles, setRoles] = useState([])
  const [grants, setGrants] = useState({}) // { [dagName]: Set<role> }
  const [newRole, setNewRole] = useState('')

  useEffect(() => {
    listDags({ activeOnly: false }).then((rows) => {
      setDags(rows)
      setRoles([...new Set(rows.flatMap((d) => d.allowedRoles))].sort())
      const initial = {}
      for (const dag of rows) initial[dag.name] = new Set(dag.allowedRoles)
      setGrants(initial)
    })
  }, [])

  function toggle(dagName, role) {
    setGrants((prev) => {
      const next = { ...prev, [dagName]: new Set(prev[dagName]) }
      next[dagName].has(role) ? next[dagName].delete(role) : next[dagName].add(role)
      return next
    })
  }

  function addRole() {
    const name = newRole.trim().toLowerCase()
    if (!name || roles.includes(name)) return
    setRoles((prev) => [...prev, name].sort())
    setNewRole('')
  }

  function removeRole(role) {
    setRoles((prev) => prev.filter((r) => r !== role))
    setGrants((prev) => {
      const next = {}
      for (const [dagName, set] of Object.entries(prev)) {
        const copy = new Set(set)
        copy.delete(role)
        next[dagName] = copy
      }
      return next
    })
  }

  if (dags === null) {
    return (
      <div className="card">
        <div className="loading">
          <div className="spinner" />
        </div>
      </div>
    )
  }

  return (
    <div className="card">
      <div className="card-head">
        <h2>Role access matrix</h2>
        <span className="muted" style={{ fontSize: 12 }}>
          {dags.length} DAGs × {roles.length} roles
        </span>
      </div>

      <div className="role-chip-row">
        {roles.map((role) => (
          <span key={role} className="role-chip">
            {role}
            <button
              onClick={() => removeRole(role)}
              aria-label={`Remove role ${role}`}
              title={`Remove role ${role}`}
            >
              ✕
            </button>
          </span>
        ))}
        <span className="role-add">
          <input
            value={newRole}
            onChange={(e) => setNewRole(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && addRole()}
            placeholder="new role…"
          />
          <button className="btn btn-sm" onClick={addRole} disabled={!newRole.trim()}>
            Add
          </button>
        </span>
      </div>

      <div className="card-body" style={{ padding: 0 }}>
        <div className="matrix-scroll">
          <table className="table matrix-table">
            <thead>
              <tr>
                <th className="matrix-corner">DAG</th>
                {roles.map((role) => (
                  <th key={role} className="matrix-role-head">
                    {role}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {dags.map((dag) => (
                <tr key={dag.name}>
                  <td>
                    <div className="cell-strong">{dag.name}</div>
                    <span
                      className={`badge ${dag.active ? 'on' : 'off'}`}
                      style={{ marginTop: 4 }}
                    >
                      <span className={`dot ${dag.active ? 'on' : 'off'}`} />
                      {dag.active ? 'Active' : 'Inactive'}
                    </span>
                  </td>
                  {roles.map((role) => (
                    <td key={role} className="matrix-cell">
                      <label className="matrix-checkbox">
                        <input
                          type="checkbox"
                          checked={grants[dag.name]?.has(role) ?? false}
                          onChange={() => toggle(dag.name, role)}
                        />
                      </label>
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      <div className="editor-foot">
        <span className="muted" style={{ fontSize: 12 }}>
          Checked = role can see &amp; run this DAG. Not saved — no backend yet.
        </span>
        <button
          className="btn btn-primary btn-sm"
          disabled
          style={{ marginLeft: 'auto' }}
        >
          Save changes
        </button>
      </div>
    </div>
  )
}
