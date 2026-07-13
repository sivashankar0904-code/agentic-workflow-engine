import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { getDag, listDags, setDagActive } from '../api/dags.js'
import { hasRole } from '../auth/session.js'
import DagGraph from '../components/graph/DagGraph.jsx'
import UploadDagModal from '../components/UploadDagModal.jsx'
import './registry.css'

export default function RegistryPage() {
  const { name } = useParams()
  return name ? <DagDetail name={name} /> : <RegistryTable />
}

// ── Governance table (design.md §4) ──────────────────────────────────────
function RegistryTable() {
  const [dags, setDags] = useState(null)
  const [busy, setBusy] = useState(null)
  const [uploadOpen, setUploadOpen] = useState(false)
  const navigate = useNavigate()
  const canGovern = hasRole('ops') || hasRole('admin')

  const load = () => {
    setDags(null)
    listDags({ activeOnly: false }).then(setDags)
  }
  useEffect(load, [])

  async function toggle(dag) {
    setBusy(dag.name)
    await setDagActive(dag.name, !dag.active)
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
        {canGovern && (
          <button className="btn btn-primary" onClick={() => setUploadOpen(true)}>
            + New DAG
          </button>
        )}
      </div>

      {uploadOpen && (
        <UploadDagModal onClose={() => setUploadOpen(false)} />
      )}

      <div className="card">
        {dags === null ? (
          <div className="loading">
            <div className="spinner" />
          </div>
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>Name</th>
                <th>Status</th>
                <th>Allowed roles</th>
                <th>Owner</th>
                <th style={{ textAlign: 'right' }}>Actions</th>
              </tr>
            </thead>
            <tbody>
              {dags.map((dag) => (
                <tr key={dag.name}>
                  <td>
                    <button
                      className="link-name"
                      onClick={() => navigate(`/registry/${dag.name}`)}
                    >
                      {dag.name}
                    </button>
                    <div className="cell-sub">{dag.description}</div>
                  </td>
                  <td>
                    <span className={`badge ${dag.active ? 'on' : 'off'}`}>
                      <span className={`dot ${dag.active ? 'on' : 'off'}`} />
                      {dag.active ? 'Active' : 'Inactive'}
                    </span>
                  </td>
                  <td>
                    <div className="chip-row">
                      {dag.allowedRoles.map((r) => (
                        <span key={r} className="chip">
                          {r}
                        </span>
                      ))}
                    </div>
                  </td>
                  <td className="muted">{dag.owner}</td>
                  <td>
                    <div className="actions">
                      {canGovern && (
                        <button
                          className={`btn btn-sm ${
                            dag.active ? 'btn-danger' : 'btn-primary'
                          }`}
                          disabled={busy === dag.name}
                          onClick={() => toggle(dag)}
                        >
                          {dag.active ? 'Deactivate' : 'Activate'}
                        </button>
                      )}
                      <button
                        className="btn btn-sm"
                        onClick={() => navigate(`/registry/${dag.name}`)}
                      >
                        Edit
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

// ── DAG detail: YAML editor + graph (design.md §5) ───────────────────────
function DagDetail({ name }) {
  const [dag, setDag] = useState(null)
  const [yaml, setYaml] = useState('')
  const [notFound, setNotFound] = useState(false)
  const navigate = useNavigate()

  useEffect(() => {
    setDag(null)
    setNotFound(false)
    getDag(name).then((d) => {
      if (!d) {
        setNotFound(true)
        return
      }
      setDag(d)
      setYaml(d.yaml)
    })
  }, [name])

  if (notFound) {
    return (
      <div className="card">
        <div className="empty">
          <strong>DAG not found</strong>
          <p className="muted">
            No workflow named “{name}” is visible to you.
          </p>
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
          <h1 style={{ display: 'inline-block' }}>{dag.name}</h1>
          <p>{dag.description}</p>
        </div>
        <div className="actions">
          <span className={`badge ${dag.active ? 'on' : 'off'}`}>
            <span className={`dot ${dag.active ? 'on' : 'off'}`} />
            {dag.active ? 'Active' : 'Inactive'}
          </span>
          <button
            className="btn btn-primary btn-sm"
            onClick={() => navigate(`/run/${dag.name}`)}
          >
            Run ▸
          </button>
        </div>
      </div>

      <div className="editor-split">
        <div className="card editor-pane">
          <div className="card-head">
            <h2>Definition (YAML)</h2>
            <span className="muted mono" style={{ fontSize: 12 }}>
              owner: {dag.owner}
            </span>
          </div>
          <textarea
            className="yaml-editor"
            value={yaml}
            onChange={(e) => setYaml(e.target.value)}
            spellCheck={false}
          />
          <div className="editor-foot">
            <button className="btn btn-sm">Validate</button>
            <button className="btn btn-primary btn-sm" disabled>
              Save
            </button>
            <span className="muted" style={{ fontSize: 12, marginLeft: 'auto' }}>
              Read-only · no backend yet
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
