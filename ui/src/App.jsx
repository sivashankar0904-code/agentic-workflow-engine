import { Navigate, Route, Routes } from 'react-router-dom'
import AppShell from './components/AppShell.jsx'
import RegistryPage from './routes/RegistryPage.jsx'
import RunPage from './routes/RunPage.jsx'
import UsersPage from './routes/UsersPage.jsx'
import { isAdmin } from './auth/session.js'

// Route guard: admin-only pages redirect non-admins back to the registry.
function RequireAdmin({ children }) {
  return isAdmin() ? children : <Navigate to="/registry" replace />
}

export default function App() {
  return (
    <Routes>
      <Route element={<AppShell />}>
        <Route index element={<Navigate to="/registry" replace />} />
        <Route path="registry" element={<RegistryPage />} />
        <Route path="registry/:name" element={<RegistryPage />} />
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
