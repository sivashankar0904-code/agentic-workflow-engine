import { useRef, useState } from 'react'
import { uploadDag } from '../api/dags.js'

// "+ New DAG" upload modal. Lets the user name the DAG and drag-and-drop or
// browse for a .yaml/.yml file (or paste YAML directly), then POSTs it to the
// Control Plane. New DAGs start inactive.
export default function UploadDagModal({ onClose, onSaved }) {
  const [name, setName] = useState('')
  const [fileName, setFileName] = useState(null)
  const [text, setText] = useState('')
  const [dragOver, setDragOver] = useState(false)
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)
  const inputRef = useRef(null)

  async function submit() {
    setError('')
    setBusy(true)
    try {
      await uploadDag(name.trim(), text)
      onSaved?.()
      onClose()
    } catch (err) {
      setError(err.message)
    } finally {
      setBusy(false)
    }
  }

  function readFile(file) {
    if (!file) return
    setFileName(file.name)
    const reader = new FileReader()
    reader.onload = () => setText(String(reader.result ?? ''))
    reader.readAsText(file)
  }

  function handleDrop(e) {
    e.preventDefault()
    setDragOver(false)
    readFile(e.dataTransfer.files?.[0])
  }

  return (
    <div className="modal-backdrop" onClick={onClose}>
      <div className="modal" onClick={(e) => e.stopPropagation()}>
        <div className="modal-head">
          <h2>Upload DAG</h2>
          <button className="icon-btn" onClick={onClose} aria-label="Close">
            ✕
          </button>
        </div>

        <div className="modal-body">
          {error && (
            <div className="alert alert-error" role="alert">
              {error}
            </div>
          )}
          <label className="field">
            <span className="field-label">DAG name</span>
            <input
              className="text-input"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="csv-flow"
              autoFocus
            />
          </label>

          <div
            className={`dropzone ${dragOver ? 'over' : ''}`}
            onDragOver={(e) => {
              e.preventDefault()
              setDragOver(true)
            }}
            onDragLeave={() => setDragOver(false)}
            onDrop={handleDrop}
            onClick={() => inputRef.current?.click()}
          >
            <input
              ref={inputRef}
              type="file"
              accept=".yaml,.yml"
              hidden
              onChange={(e) => readFile(e.target.files?.[0])}
            />
            <span className="dropzone-icon">⇪</span>
            {fileName ? (
              <>
                <strong>{fileName}</strong>
                <span className="muted">Click or drop to replace</span>
              </>
            ) : (
              <>
                <strong>Drop a .yaml file here</strong>
                <span className="muted">or click to browse</span>
              </>
            )}
          </div>

          <div className="modal-divider">
            <span>or paste YAML</span>
          </div>

          <textarea
            className="yaml-editor modal-textarea"
            placeholder={'nodes:\n  - name: ingest\n    topic: ingest\n    host: localhost\n    port: 8001\n    entry: true\n    routes: []'}
            value={text}
            onChange={(e) => {
              setFileName(null)
              setText(e.target.value)
            }}
            spellCheck={false}
          />
        </div>

        <div className="modal-foot">
          <span className="muted" style={{ fontSize: 12 }}>
            New DAGs start inactive.
          </span>
          <div className="actions">
            <button className="btn btn-sm" onClick={onClose}>
              Cancel
            </button>
            <button
              className="btn btn-primary btn-sm"
              disabled={busy || !name.trim() || !text.trim()}
              onClick={submit}
            >
              {busy ? 'Uploading…' : 'Upload'}
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
