import { Navigate, Route, Routes, useLocation } from 'react-router-dom'
import AppShell from './components/AppShell.jsx'
import RegistryPage from './routes/RegistryPage.jsx'
import RunPage from './routes/RunPage.jsx'
import UsersPage from './routes/UsersPage.jsx'
import LoginPage from './routes/LoginPage.jsx'
import { isAdmin, isAuthenticated } from './auth/session.js'

// Require a logged-in session; unauthenticated users go to /login, preserving
// where they were headed so login can send them back.
function RequireAuth({ children }) {
  const location = useLocation()
  if (!isAuthenticated()) {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />
  }
  return children
}

// Route guard: admin-only pages redirect non-admins back to the registry.
function RequireAdmin({ children }) {
  return isAdmin() ? children : <Navigate to="/registry" replace />
}

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route
        element={
          <RequireAuth>
            <AppShell />
          </RequireAuth>
        }
      >
        <Route index element={<Navigate to="/registry" replace />} />
        <Route path="registry" element={<RegistryPage />} />
        <Route path="registry/:id" element={<RegistryPage />} />
        <Route path="run" element={<RunPage />} />
        <Route path="run/:name" element={<RunPage />} />
        <Route
          path="users"
          element={
            <RequireAdmin>
              <UsersPage />
            </RequireAdmin>
          }
        />
        <Route path="*" element={<Navigate to="/registry" replace />} />
      </Route>
    </Routes>
  )
}
