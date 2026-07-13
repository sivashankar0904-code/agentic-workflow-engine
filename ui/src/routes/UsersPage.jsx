import { useEffect, useState } from 'react'
import { listUsers, setUserRole, setUserStatus } from '../api/users.js'
import NewUserModal from '../components/NewUserModal.jsx'
import ManageAccessModal from '../components/ManageAccessModal.jsx'
import RolesAccessPanel from './RolesAccessPanel.jsx'

const ROLES = ['admin', 'editor', 'ops', 'viewer']

// Users & RBAC admin console (design.md §8). Admin-gated at the route level.
// Two tabs: per-user list (roles, status, individual DAG access) and the
// role-level resource-allocation matrix (which roles can see which DAGs).
export default function UsersPage() {
  const [tab, setTab] = useState('users') // 'users' | 'roles'
  const [users, setUsers] = useState(null)
  const [newUserOpen, setNewUserOpen] = useState(false)
  const [accessUser, setAccessUser] = useState(null)

  const load = () => {
    setUsers(null)
    listUsers().then(setUsers)
  }
  useEffect(load, [])

  const activeAdmins = (users ?? []).filter(
    (u) => u.role === 'admin' && u.status === 'active'
  ).length

  async function changeRole(user, role) {
    await setUserRole(user.username, role)
    load()
  }
  async function toggleStatus(user) {
    const next = user.status === 'active' ? 'disabled' : 'active'
    await setUserStatus(user.username, next)
    load()
  }

  // The server refuses to demote/disable the last active admin; mirror that guard.
  const isLastAdmin = (user) =>
    user.role === 'admin' && user.status === 'active' && activeAdmins <= 1

  return (
    <>
      <div className="page-head">
        <div>
          <h1>Users &amp; Access</h1>
          <p>
            RBAC is enforced by the Control Plane. This console manages roles,
            status, and DAG access; the server is the real gate.
          </p>
        </div>
        {tab === 'users' && (
          <button className="btn btn-primary" onClick={() => setNewUserOpen(true)}>
            + New user
          </button>
        )}
      </div>

      {newUserOpen && (
        <NewUserModal onClose={() => setNewUserOpen(false)} />
      )}
      {accessUser && (
        <ManageAccessModal user={accessUser} onClose={() => setAccessUser(null)} />
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
                <th>DAG access</th>
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
                        title={
                          locked ? 'Cannot demote the last active admin' : ''
                        }
                      >
                        {ROLES.map((r) => (
                          <option key={r} value={r}>
                            {r}
                          </option>
                        ))}
                      </select>
                    </td>
                    <td>
                      <span
                        className={`badge ${
                          user.status === 'active' ? 'on' : 'off'
                        }`}
                      >
                        <span
                          className={`dot ${
                            user.status === 'active' ? 'on' : 'off'
                          }`}
                        />
                        {user.status === 'active' ? 'Active' : 'Disabled'}
                      </span>
                    </td>
                    <td>
                      {user.assignedDags.length === 0 ? (
                        <span className="muted" style={{ fontSize: 12.5 }}>
                          none
                        </span>
                      ) : (
                        <div className="dag-chips">
                          {user.assignedDags.map((name) => (
                            <span key={name} className="chip">
                              {name}
                            </span>
                          ))}
                        </div>
                      )}
                    </td>
                    <td>
                      <div className="actions">
                        <button
                          className="btn btn-sm"
                          onClick={() => setAccessUser(user)}
                        >
                          Manage access
                        </button>
                        <button className="btn btn-sm">Reset pw</button>
                        <button
                          className={`btn btn-sm ${
                            user.status === 'active' ? 'btn-danger' : ''
                          }`}
                          disabled={locked}
                          onClick={() => toggleStatus(user)}
                          title={
                            locked
                              ? 'Cannot disable the last active admin'
                              : ''
                          }
                        >
                          {user.status === 'active' ? 'Disable' : 'Enable'}
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
