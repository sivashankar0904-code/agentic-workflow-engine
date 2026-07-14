import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { getDag, listDags, setDagActive } from '../api/dags.js'
import { hasPermission } from '../auth/session.js'
import DagGraph from '../components/graph/DagGraph.jsx'
import UploadDagModal from '../components/UploadDagModal.jsx'
import './registry.css'

const PERM_PATCH = 'control_plane.dag_registry.dag.patch'
const PERM_CREATE = 'control_plane.dag_registry.dag.create'

export default function RegistryPage() {
  const { id } = useParams()
  return id ? <DagDetail id={id} /> : <RegistryTable />
}

// ── Governance table (design.md §4) ──────────────────────────────────────
function RegistryTable() {
  const [dags, setDags] = useState(null)
  const [busy, setBusy] = useState(null)
  const [uploadOpen, setUploadOpen] = useState(false)
  const navigate = useNavigate()
  const canCreate = hasPermission(PERM_CREATE)
  const canPatch = hasPermission(PERM_PATCH)

  const load = () => {
    setDags(null)
    listDags({ activeOnly: false }).then(setDags)
  }
  useEffect(load, [])

  async function toggle(dag) {
    setBusy(dag.id)
    await setDagActive(dag.id, !dag.active)
    setBusy(null)
    load()
  }

  return (
    <>
      <div className="page-head">
        <div>
          <h1>DAG Registry</h1>
          <p>
            Every stored workflow. Active flows are served to the execution
            engines; inactive ones are retained but not run.
          </p>
        </div>
        {canCreate && (
          <button className="btn btn-primary" onClick={() => setUploadOpen(true)}>
            + New DAG
          </button>
        )}
      </div>

      {uploadOpen && <UploadDagModal onClose={() => setUploadOpen(false)} onSaved={load} />}

      <div className="card">
        {dags === null ? (
          <div className="loading">
            <div className="spinner" />
          </div>
        ) : dags.length === 0 ? (
          <div className="empty">
            <strong>No DAGs yet</strong>
            <p className="muted">Upload a workflow definition to get started.</p>
          </div>
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>Name</th>
                <th>Status</th>
                <th style={{ textAlign: 'right' }}>Actions</th>
              </tr>
            </thead>
            <tbody>
              {dags.map((dag) => (
                <tr key={dag.id}>
                  <td>
                    <button
                      className="link-name"
                      onClick={() => navigate(`/registry/${dag.id}`)}
                    >
                      {dag.name}
                    </button>
                  </td>
                  <td>
                    <span className={`badge ${dag.active ? 'on' : 'off'}`}>
                      <span className={`dot ${dag.active ? 'on' : 'off'}`} />
                      {dag.active ? 'Active' : 'Inactive'}
                    </span>
                  </td>
                  <td>
                    <div className="actions">
                      {canPatch && (
                        <button
                          className={`btn btn-sm ${
                            dag.active ? 'btn-danger' : 'btn-primary'
                          }`}
                          disabled={busy === dag.id}
                          onClick={() => toggle(dag)}
                        >
                          {dag.active ? 'Deactivate' : 'Activate'}
                        </button>
                      )}
                      <button
                        className="btn btn-sm"
                        onClick={() => navigate(`/registry/${dag.id}`)}
                      >
                        View
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </>
  )
}

// ── DAG detail: YAML + graph (design.md §5) ──────────────────────────────
function DagDetail({ id }) {
  const [dag, setDag] = useState(null)
  const [yaml, setYaml] = useState('')
  const [notFound, setNotFound] = useState(false)
  const navigate = useNavigate()

  useEffect(() => {
    setDag(null)
    setNotFound(false)
    getDag(id).then((d) => {
      if (!d) {
        setNotFound(true)
        return
      }
      setDag(d)
      setYaml(d.yaml)
    })
  }, [id])

  if (notFound) {
    return (
      <div className="card">
        <div className="empty">
          <strong>DAG not found</strong>
          <p className="muted">No workflow with this id is visible to you.</p>
          <button className="btn" onClick={() => navigate('/registry')}>
            Back to registry
          </button>
        </div>
      </div>
    )
  }

  if (dag === null) {
    return (
      <div className="card">
        <div className="loading">
          <div className="spinner" />
        </div>
      </div>
    )
  }

  return (
    <>
      <div className="page-head">
        <div>
          <button className="crumb" onClick={() => navigate('/registry')}>
            Registry
          </button>
          <span className="crumb-sep">/</span>
          <h1 style={{ display: 'inline-block' }}>DAG #{dag.id}</h1>
        </div>
      </div>

      <div className="editor-split">
        <div className="card editor-pane">
          <div className="card-head">
            <h2>Definition (YAML)</h2>
          </div>
          <textarea
            className="yaml-editor"
            value={yaml}
            onChange={(e) => setYaml(e.target.value)}
            spellCheck={false}
          />
          <div className="editor-foot">
            <span className="muted" style={{ fontSize: 12, marginLeft: 'auto' }}>
              Read-only view
            </span>
          </div>
        </div>

        <div className="card graph-pane">
          <div className="card-head">
            <h2>Graph</h2>
            <span className="muted" style={{ fontSize: 12 }}>
              {dag.nodes.length} nodes
            </span>
          </div>
          <div className="graph-host">
            <DagGraph dag={dag} />
          </div>
        </div>
      </div>
    </>
  )
}
