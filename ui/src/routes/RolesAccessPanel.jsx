import { useEffect, useState } from 'react'
import {
  listRoles,
  listPermissions,
  createRole,
  deleteRole,
  setRolePermissions,
} from '../api/roles.js'

// Role × permission matrix (design.md §8), backed by the Control Plane.
// Rows are permissions grouped service → feature_group → feature; columns are
// roles. Ticking a cell replaces that role's exact permission set. Roles are
// customizable — add or remove one. All changes persist immediately; the
// server re-enforces.
export default function RolesAccessPanel() {
  const [roles, setRoles] = useState(null)
  const [perms, setPerms] = useState([])
  const [newRole, setNewRole] = useState('')
  const [error, setError] = useState('')
  const [saving, setSaving] = useState(false)

  const load = () => {
    setError('')
    Promise.all([listRoles(), listPermissions()])
      .then(([r, p]) => {
        setRoles(r)
        setPerms(p)
      })
      .catch((err) => setError(err.message))
  }
  useEffect(load, [])

  // Toggle one permission for one role, then persist the role's full new set.
  async function toggle(role, permId) {
    const current = new Set(role.permissionIds())
    if (current.has(permId)) current.delete(permId)
    else current.add(permId)
    setSaving(true)
    setError('')
    try {
      await setRolePermissions(role.name, [...current])
      load()
    } catch (err) {
      setError(err.message)
    } finally {
      setSaving(false)
    }
  }

  async function addRole() {
    const name = newRole.trim().toLowerCase()
    if (!name) return
    setError('')
    try {
      await createRole(name)
      setNewRole('')
      load()
    } catch (err) {
      setError(err.message)
    }
  }

  async function removeRole(name) {
    setError('')
    try {
      await deleteRole(name)
      load()
    } catch (err) {
      setError(err.message)
    }
  }

  if (roles === null) {
    return (
      <div className="card">
        <div className="loading">
          <div className="spinner" />
        </div>
      </div>
    )
  }

  // Decorate each role with a permission-id lookup derived from its dotted
  // permission names against the permission catalog.
  const permIdByName = new Map(perms.map((p) => [p.name, p.id]))
  const decorated = roles.map((r) => ({
    ...r,
    permissionIds: () => (r.permissions || []).map((n) => permIdByName.get(n)).filter(Boolean),
    hasPerm: (permName) => (r.permissions || []).includes(permName),
  }))

  return (
    <div className="card">
      <div className="card-head">
        <h2>Role access matrix</h2>
        <span className="muted" style={{ fontSize: 12 }}>
          {perms.length} permissions × {roles.length} roles
          {saving ? ' · saving…' : ''}
        </span>
      </div>

      {error && (
        <div className="alert alert-error" role="alert" style={{ margin: 12 }}>
          {error}
        </div>
      )}

      <div className="role-chip-row">
        {decorated.map((role) => (
          <span key={role.name} className="role-chip">
            {role.name}
            {role.name !== 'admin' && (
              <button
                onClick={() => removeRole(role.name)}
                aria-label={`Remove role ${role.name}`}
                title={`Remove role ${role.name}`}
              >
                ✕
              </button>
            )}
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
                <th className="matrix-corner">Permission</th>
                {decorated.map((role) => (
                  <th key={role.name} className="matrix-role-head">
                    {role.name}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {perms.map((perm) => (
                <tr key={perm.id}>
                  <td>
                    <div className="cell-strong mono" style={{ fontSize: 12 }}>
                      {perm.featureGroup}.{perm.feature}.{perm.action}
                    </div>
                    <div className="cell-sub">{perm.serviceKey}</div>
                  </td>
                  {decorated.map((role) => {
                    const admin = role.name === 'admin'
                    return (
                      <td key={role.name} className="matrix-cell">
                        <label className="matrix-checkbox">
                          <input
                            type="checkbox"
                            checked={role.hasPerm(perm.name)}
                            disabled={admin || saving}
                            title={admin ? 'admin always has every permission' : ''}
                            onChange={() => toggle(role, perm.id)}
                          />
                        </label>
                      </td>
                    )
                  })}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      <div className="editor-foot">
        <span className="muted" style={{ fontSize: 12 }}>
          Checked = the role holds that permission. Changes save immediately.
        </span>
      </div>
    </div>
  )
}
