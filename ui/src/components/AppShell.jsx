import { NavLink, Outlet } from 'react-router-dom'
import Sidebar from './Sidebar.jsx'
import { useTheme } from './useTheme.js'
import { getSession, isAdmin } from '../auth/session.js'

export default function AppShell() {
  const [theme, toggleTheme] = useTheme()
  const session = getSession()
  const admin = isAdmin()

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
          <button className="account" title={`Signed in as ${session.username}`}>
            <span className="avatar">{session.username.slice(0, 1)}</span>
            <span>{session.username}</span>
            <span className="muted" aria-hidden>
              ▾
            </span>
          </button>
        </div>
      </header>

      <Sidebar />

      <main className="main">
        <Outlet />
      </main>
    </div>
  )
}
