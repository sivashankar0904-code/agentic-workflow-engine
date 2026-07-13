import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { listDags } from '../api/dags.js'

// Registry list with an Active/All toggle (design.md §3, §4). Selecting a DAG
// navigates to its registry route; the toggle is just a filter over the same
// fetch (the real app maps it to GET /dags?active=true).
export default function Sidebar() {
  const [activeOnly, setActiveOnly] = useState(true)
  const [dags, setDags] = useState(null)
  const navigate = useNavigate()
  const { name: selected } = useParams()

  useEffect(() => {
    let alive = true
    setDags(null)
    listDags({ activeOnly }).then((rows) => {
      if (alive) setDags(rows)
    })
    return () => {
      alive = false
    }
  }, [activeOnly])

  const activeCount = dags?.filter((d) => d.active).length ?? 0

  return (
    <aside className="sidebar">
      <div className="sidebar-title">DAG Registry</div>

      <div className="toggle" role="tablist">
        <button
          className={activeOnly ? 'on' : ''}
          onClick={() => setActiveOnly(true)}
        >
          Active{dags && activeOnly ? ` (${dags.length})` : ''}
        </button>
        <button
          className={!activeOnly ? 'on' : ''}
          onClick={() => setActiveOnly(false)}
        >
          All{dags && !activeOnly ? ` (${dags.length})` : ''}
        </button>
      </div>

      {dags === null ? (
        <div className="loading" style={{ minHeight: 120 }}>
          <div className="spinner" />
        </div>
      ) : (
        <div className="dag-list">
          {dags.map((dag) => (
            <div
              key={dag.name}
              className={`dag-item ${selected === dag.name ? 'active' : ''}`}
              onClick={() => navigate(`/registry/${dag.name}`)}
            >
              <span className="name">{dag.name}</span>
              <span
                className={`dot ${dag.active ? 'on' : 'off'}`}
                title={dag.active ? 'active' : 'inactive'}
              />
            </div>
          ))}
        </div>
      )}

      {dags && activeOnly && (
        <p className="muted" style={{ fontSize: 12, padding: '10px 8px 0' }}>
          {activeCount} served to engines.
        </p>
      )}
    </aside>
  )
}
