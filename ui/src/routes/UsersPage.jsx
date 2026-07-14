import { useEffect, useState } from 'react'
import { listUsers, setUserRole, setUserActive } from '../api/users.js'
import { listRoles } from '../api/roles.js'
import NewUserModal from '../components/NewUserModal.jsx'
import RolesAccessPanel from './RolesAccessPanel.jsx'

// Users & RBAC admin console (design.md §8). Admin-gated at the route level.
// Two tabs: per-user list (role, active state) and the role/permission
// resource-allocation matrix. All enforcement is server-side; this console is
// a convenience over the Control Plane.
export default function UsersPage() {
  const [tab, setTab] = useState('users') // 'users' | 'roles'
  const [users, setUsers] = useState(null)
  const [roles, setRoles] = useState([])
  const [newUserOpen, setNewUserOpen] = useState(false)
  const [error, setError] = useState('')

  const load = () => {
    setUsers(null)
    setError('')
    Promise.all([listUsers(), listRoles()])
      .then(([u, r]) => {
        setUsers(u)
        setRoles(r.map((role) => role.name))
      })
      .catch((err) => setError(err.message))
  }
  useEffect(load, [])

  const activeAdmins = (users ?? []).filter(
    (u) => u.role === 'admin' && u.isActive
  ).length

  async function changeRole(user, role) {
    try {
      await setUserRole(user.username, role)
      load()
    } catch (err) {
      setError(err.message)
    }
  }
  async function toggleActive(user) {
    try {
      await setUserActive(user.username, !user.isActive)
      load()
    } catch (err) {
      setError(err.message)
    }
  }

  // The server refuses to demote/disable the last active admin; mirror that guard.
  const isLastAdmin = (user) =>
    user.role === 'admin' && user.isActive && activeAdmins <= 1

  return (
    <>
      <div className="page-head">
        <div>
          <h1>Users &amp; Access</h1>
          <p>
            RBAC is enforced by the Control Plane. This console manages roles,
            active state, and role permissions; the server is the real gate.
          </p>
        </div>
        {tab === 'users' && (
          <button className="btn btn-primary" onClick={() => setNewUserOpen(true)}>
            + New user
          </button>
        )}
      </div>

      {error && (
        <div className="alert alert-error" role="alert">
          {error}
        </div>
      )}

      {newUserOpen && (
        <NewUserModal
          roles={roles}
          onClose={() => setNewUserOpen(false)}
          onCreated={load}
        />
      )}

      <div className="tab-row">
        <button
          className={`tab-btn ${tab === 'users' ? 'on' : ''}`}
          onClick={() => setTab('users')}
        >
          Users
        </button>
        <button
          className={`tab-btn ${tab === 'roles' ? 'on' : ''}`}
          onClick={() => setTab('roles')}
        >
          Roles &amp; Access
        </button>
      </div>

      {tab === 'roles' ? (
        <RolesAccessPanel />
      ) : (
        <div className="card">
          {users === null ? (
            <div className="loading">
              <div className="spinner" />
            </div>
          ) : (
            <table className="table">
              <thead>
                <tr>
                  <th>User</th>
                  <th>Role</th>
                  <th>Status</th>
                  <th style={{ textAlign: 'right' }}>Actions</th>
                </tr>
              </thead>
              <tbody>
                {users.map((user) => {
                  const locked = isLastAdmin(user)
                  return (
                    <tr key={user.username}>
                      <td>
                        <div className="cell-strong">{user.username}</div>
                        <div className="cell-sub">{user.email}</div>
                      </td>
                      <td>
                        <select
                          className="role-select"
                          value={user.role}
                          disabled={locked}
                          onChange={(e) => changeRole(user, e.target.value)}
                          title={locked ? 'Cannot demote the last active admin' : ''}
                        >
                          {roles.map((r) => (
                            <option key={r} value={r}>
                              {r}
                            </option>
                          ))}
                        </select>
                      </td>
                      <td>
                        <span className={`badge ${user.isActive ? 'on' : 'off'}`}>
                          <span className={`dot ${user.isActive ? 'on' : 'off'}`} />
                          {user.isActive ? 'Active' : 'Disabled'}
                        </span>
                      </td>
                      <td>
                        <div className="actions">
                          <button
                            className={`btn btn-sm ${user.isActive ? 'btn-danger' : ''}`}
                            disabled={locked}
                            onClick={() => toggleActive(user)}
                            title={
                              locked ? 'Cannot disable the last active admin' : ''
                            }
                          >
                            {user.isActive ? 'Disable' : 'Enable'}
                          </button>
                        </div>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          )}
        </div>
      )}
    </>
  )
}
