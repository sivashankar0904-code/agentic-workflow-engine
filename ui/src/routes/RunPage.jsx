import { useEffect, useRef, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { getDag, listDags } from '../api/dags.js'
import { sampleFor, submitMessage } from '../api/run.js'
import DagGraph from '../components/graph/DagGraph.jsx'
import './run.css'

export default function RunPage() {
  const { name } = useParams()
  const navigate = useNavigate()
  const [active, setActive] = useState(null)

  // Only active DAGs can be run (design.md §6 — engines build the active set).
  useEffect(() => {
    listDags({ activeOnly: true }).then((rows) => {
      setActive(rows)
      if (!name && rows.length) navigate(`/run/${rows[0].name}`, { replace: true })
    })
  }, [name, navigate])

  return (
    <>
      <div className="page-head">
        <div>
          <h1>Run Console</h1>
          <p>
            Submit a workflow message to the execution ingress and watch it flow
            through the graph.
          </p>
        </div>
      </div>

      {active === null ? (
        <div className="card">
          <div className="loading">
            <div className="spinner" />
          </div>
        </div>
      ) : (
        <div className="run-picker">
          {active.map((d) => (
            <button
              key={d.name}
              className={`run-tab ${name === d.name ? 'on' : ''}`}
              onClick={() => navigate(`/run/${d.name}`)}
            >
              {d.name}
            </button>
          ))}
        </div>
      )}

      {name && <RunConsole name={name} />}
    </>
  )
}

const STEP_MS = 750

function RunConsole({ name }) {
  const [dag, setDag] = useState(null)
  const [message, setMessage] = useState('')
  const [run, setRun] = useState(null) // { path, result }
  const [stepIndex, setStepIndex] = useState(-1)
  const [status, setStatus] = useState('idle') // idle | running | done
  const timer = useRef(null)

  useEffect(() => {
    setDag(null)
    setRun(null)
    setStepIndex(-1)
    setStatus('idle')
    getDag(name).then((d) => {
      setDag(d)
      setMessage(sampleFor(name))
    })
    return () => clearInterval(timer.current)
  }, [name])

  async function send() {
    clearInterval(timer.current)
    setStatus('running')
    setStepIndex(-1)
    const result = await submitMessage(name, message)
    setRun(result)

    // Animate the message advancing node-to-node.
    let i = 0
    setStepIndex(0)
    timer.current = setInterval(() => {
      i += 1
      if (i >= result.path.length) {
        clearInterval(timer.current)
        setStatus('done')
        return
      }
      setStepIndex(i)
    }, STEP_MS)
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
    <div className="run-grid">
      <div className="card">
        <div className="card-head">
          <h2>Message</h2>
          <span className="muted" style={{ fontSize: 12 }}>
            entry: {dag.nodes.find((n) => n.entry)?.name}
          </span>
        </div>
        <div className="card-body">
          <div className="msg-row">
            <input
              className="msg-input"
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              placeholder="Workflow message…"
            />
            <button
              className="btn btn-primary"
              onClick={send}
              disabled={status === 'running' || !message.trim()}
            >
              {status === 'running' ? 'Sending…' : 'Send'}
            </button>
          </div>

          <div className="flow-status">
            <StatusPill status={status} />
            {run && (
              <div className="flow-steps">
                {run.path.map((node, i) => (
                  <span key={node} className="flow-step">
                    <span
                      className={`dot ${
                        i <= stepIndex
                          ? i === stepIndex && status === 'running'
                            ? 'running'
                            : 'on'
                          : 'off'
                      }`}
                    />
                    <span className={i <= stepIndex ? '' : 'muted'}>{node}</span>
                    {i < run.path.length - 1 && (
                      <span className="flow-arrow">→</span>
                    )}
                  </span>
                ))}
              </div>
            )}
          </div>

          {status === 'done' && run && (
            <pre className="result">
              {JSON.stringify(run.result.output, null, 2)}
            </pre>
          )}
        </div>
      </div>

      <div className="card graph-pane run-graph">
        <div className="card-head">
          <h2>Flow</h2>
          <span className="muted" style={{ fontSize: 12 }}>
            live path overlay
          </span>
        </div>
        <div className="graph-host">
          <DagGraph
            dag={dag}
            activePath={run?.path ?? null}
            activeIndex={stepIndex}
          />
        </div>
      </div>
    </div>
  )
}

function StatusPill({ status }) {
  if (status === 'idle')
    return <span className="badge off">Ready to send</span>
  if (status === 'running')
    return (
      <span className="badge on" style={{ background: 'var(--surface-2)' }}>
        <span className="dot running" />
        Running
      </span>
    )
  return (
    <span className="badge on">
      <span className="dot on" />
      Completed
    </span>
  )
}
