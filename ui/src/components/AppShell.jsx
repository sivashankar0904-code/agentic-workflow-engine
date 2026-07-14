import { useState } from 'react'
import { NavLink, Outlet, useNavigate } from 'react-router-dom'
import Sidebar from './Sidebar.jsx'
import { useTheme } from './useTheme.js'
import ChangePasswordModal from './ChangePasswordModal.jsx'
import { getSession, isAdmin, logout } from '../auth/session.js'

export default function AppShell() {
  const [theme, toggleTheme] = useTheme()
  const [menuOpen, setMenuOpen] = useState(false)
  const [pwOpen, setPwOpen] = useState(false)
  const navigate = useNavigate()
  const session = getSession()
  const admin = isAdmin()

  function signOut() {
    logout()
    navigate('/login', { replace: true })
  }

  return (
    <div className="shell">
      <header className="header">
        <div className="brand">
          <span className="brand-mark">⛓</span>
          <span>Agentic Workflow Engine</span>
        </div>

        <nav className="header-nav">
          <NavLink to="/registry" className="nav-link">
            Registry
          </NavLink>
          <NavLink to="/run" className="nav-link">
            Run
          </NavLink>
          {admin && (
            <NavLink to="/users" className="nav-link">
              Users
            </NavLink>
          )}
        </nav>

        <div className="header-spacer" />

        <div className="header-tools">
          <button
            className="icon-btn"
            onClick={toggleTheme}
            title={`Switch to ${theme === 'dark' ? 'light' : 'dark'} theme`}
            aria-label="Toggle theme"
          >
            {theme === 'dark' ? '☀' : '☾'}
          </button>

          <div className="account-wrap">
            <button
              className="account"
              onClick={() => setMenuOpen((o) => !o)}
              title={`Signed in as ${session?.username}`}
            >
              <span className="avatar">{session?.username?.slice(0, 1)?.toUpperCase()}</span>
              <span>{session?.username}</span>
              <span className="muted" aria-hidden>
                ▾
              </span>
            </button>
            {menuOpen && (
              <>
                <div className="menu-backdrop" onClick={() => setMenuOpen(false)} />
                <div className="account-menu" role="menu">
                  <div className="account-menu-head">
                    <div className="cell-strong">{session?.username}</div>
                    <div className="cell-sub">{session?.role}</div>
                  </div>
                  <button
                    className="account-menu-item"
                    onClick={() => {
                      setMenuOpen(false)
                      setPwOpen(true)
                    }}
                  >
                    Change password
                  </button>
                  <button className="account-menu-item danger" onClick={signOut}>
                    Log out
                  </button>
                </div>
              </>
            )}
          </div>
        </div>
      </header>

      {pwOpen && <ChangePasswordModal onClose={() => setPwOpen(false)} />}

      <Sidebar />

      <main className="main">
        <Outlet />
      </main>
    </div>
  )
}
