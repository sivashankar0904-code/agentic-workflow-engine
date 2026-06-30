import { useState, useEffect, useRef } from "react";

const SVC1_URL      = "http://localhost:8001/api/v1/chat/";
const ORCHESTRATOR  = "http://localhost:8000";

const SERVICES = [
  { label: "mock_service_2", sub: "CSV",   url: "http://localhost:8002/api/v1/messages/" },
  { label: "mock_service_3", sub: "Excel", url: "http://localhost:8003/api/v1/messages/" },
  { label: "mock_service_4", sub: "PDF",   url: "http://localhost:8004/api/v1/messages/" },
];

// ── Hooks ─────────────────────────────────────────────────────────────────────

function useMessages(url) {
  const [messages, setMessages] = useState([]);
  const [status, setStatus]     = useState("connecting");
  useEffect(() => {
    let cancelled = false;
    async function poll() {
      try {
        const res = await fetch(url);
        if (!res.ok) throw new Error();
        const data = await res.json();
        if (!cancelled) { setMessages(data.messages ?? []); setStatus("online"); }
      } catch { if (!cancelled) setStatus("offline"); }
    }
    poll();
    const id = setInterval(poll, 3000);
    return () => { cancelled = true; clearInterval(id); };
  }, [url]);
  return { messages, status };
}

function useDAG() {
  const [config, setConfig] = useState(null);
  const load = async () => {
    try {
      const res = await fetch(`${ORCHESTRATOR}/config`);
      if (res.ok) setConfig(await res.json());
    } catch {}
  };
  useEffect(() => { load(); }, []);
  return { config, reload: load };
}

// ── Components ────────────────────────────────────────────────────────────────

function ServicePanel({ label, sub, url }) {
  const { messages, status } = useMessages(url);
  return (
    <div style={s.panel}>
      <div style={s.panelHeader}>
        <span style={s.panelTitle}>{label}</span>
        <span style={{ ...s.badge, background: status === "online" ? "#22c55e" : status === "offline" ? "#ef4444" : "#f59e0b" }}>
          {status}
        </span>
      </div>
      <div style={s.panelSub}>{sub} messages</div>
      <div style={s.msgList}>
        {messages.length === 0
          ? <span style={s.empty}>No messages yet</span>
          : [...messages].reverse().map((m, i) => (
              <div key={i} style={s.msg}>{m.message ?? JSON.stringify(m)}</div>
            ))}
      </div>
    </div>
  );
}

function DAGViewer({ config }) {
  if (!config) return <div style={s.empty}>No DAG loaded — upload a YAML config.</div>;
  return (
    <div>
      <div style={{ marginBottom: 12 }}>
        <strong>Source:</strong> <code style={s.code}>{config.routing?.source}</code>
      </div>
      <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
        {(config.routing?.rules ?? []).map((rule, i) => (
          <div key={i} style={s.ruleRow}>
            <div style={s.ruleBox}>
              <span style={s.ruleLabel}>if</span>
              <code style={s.code}>{rule.condition?.field} contains "{rule.condition?.contains}"</code>
            </div>
            <span style={s.arrow}>→</span>
            <div style={{ ...s.ruleBox, background: "#eff6ff", borderColor: "#93c5fd" }}>
              <code style={s.code}>{rule.target}</code>
            </div>
          </div>
        ))}
      </div>
      <div style={{ marginTop: 16 }}>
        <strong>Services:</strong>
        <div style={{ display: "flex", gap: 8, flexWrap: "wrap", marginTop: 8 }}>
          {(config.services ?? []).map(svc => (
            <div key={svc.name} style={s.svcChip}>
              {svc.name} <span style={s.chipSub}>{svc.host}:{svc.port}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

function YAMLUpload({ onUploaded }) {
  const fileRef = useRef();
  const [status, setStatus] = useState(null);

  async function upload(e) {
    const file = e.target.files[0];
    if (!file) return;
    setStatus("uploading");
    try {
      const text = await file.text();
      const res = await fetch(`${ORCHESTRATOR}/config`, {
        method: "POST",
        headers: { "Content-Type": "application/x-yaml" },
        body: text,
      });
      if (!res.ok) throw new Error(await res.text());
      setStatus("ok");
      onUploaded();
    } catch (err) {
      setStatus("error: " + err.message);
    }
    fileRef.current.value = "";
  }

  return (
    <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
      <label style={s.uploadBtn}>
        Upload YAML
        <input ref={fileRef} type="file" accept=".yaml,.yml" style={{ display: "none" }} onChange={upload} />
      </label>
      {status && (
        <span style={{ fontSize: 13, color: status === "ok" ? "#16a34a" : status === "uploading" ? "#ca8a04" : "#dc2626" }}>
          {status === "ok" ? "✅ Config reloaded" : status}
        </span>
      )}
    </div>
  );
}

// ── App ───────────────────────────────────────────────────────────────────────

export default function App() {
  const [fileType, setFileType] = useState("CSV");
  const [message,  setMessage]  = useState("");
  const [reply,    setReply]    = useState(null);
  const [error,    setError]    = useState(null);
  const [sending,  setSending]  = useState(false);
  const [tab,      setTab]      = useState("inbox");
  const { config, reload }      = useDAG();

  async function send() {
    if (!message.trim()) return;
    setSending(true); setReply(null); setError(null);
    try {
      const res = await fetch(SVC1_URL, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ message: `[${fileType}] ${message}` }),
      });
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      setReply((await res.json()).reply);
      setMessage("");
    } catch (e) {
      setError(e.message);
    } finally {
      setSending(false);
    }
  }

  return (
    <div style={s.root}>
      <h1 style={s.title}>Agentic AI Orchestrator</h1>

      {/* Send */}
      <div style={s.sendBox}>
        <select value={fileType} onChange={e => setFileType(e.target.value)} style={s.select}>
          {["CSV", "Excel", "PDF"].map(t => <option key={t}>{t}</option>)}
        </select>
        <input style={s.input} placeholder="Type a message..." value={message}
          onChange={e => setMessage(e.target.value)} onKeyDown={e => e.key === "Enter" && send()} />
        <button style={s.btn} onClick={send} disabled={sending}>{sending ? "Sending…" : "Send"}</button>
      </div>
      {reply && <div style={s.reply}>✅ {reply}</div>}
      {error && <div style={s.err}>❌ {error}</div>}

      {/* Tabs */}
      <div style={s.tabs}>
        {["inbox", "dag"].map(t => (
          <button key={t} onClick={() => setTab(t)}
            style={{ ...s.tab, ...(tab === t ? s.tabActive : {}) }}>
            {t === "inbox" ? "Service Inbox" : "DAG Config"}
          </button>
        ))}
        <div style={{ marginLeft: "auto" }}>
          <YAMLUpload onUploaded={reload} />
        </div>
      </div>

      {tab === "inbox" && (
        <div style={s.grid}>
          {SERVICES.map(sv => <ServicePanel key={sv.label} {...sv} />)}
        </div>
      )}

      {tab === "dag" && (
        <div style={s.dagBox}>
          <DAGViewer config={config} />
        </div>
      )}
    </div>
  );
}

// ── Styles ────────────────────────────────────────────────────────────────────

const s = {
  root:      { fontFamily: "system-ui,sans-serif", maxWidth: 1100, margin: "0 auto", padding: 24 },
  title:     { fontSize: 24, fontWeight: 700, marginBottom: 20 },
  sendBox:   { display: "flex", gap: 8, marginBottom: 12 },
  select:    { padding: "8px 12px", borderRadius: 6, border: "1px solid #d1d5db", fontSize: 14 },
  input:     { flex: 1, padding: "8px 12px", borderRadius: 6, border: "1px solid #d1d5db", fontSize: 14 },
  btn:       { padding: "8px 20px", borderRadius: 6, background: "#2563eb", color: "#fff", border: "none", cursor: "pointer", fontSize: 14 },
  reply:     { padding: "8px 12px", background: "#f0fdf4", border: "1px solid #86efac", borderRadius: 6, marginBottom: 12, fontSize: 14 },
  err:       { padding: "8px 12px", background: "#fef2f2", border: "1px solid #fca5a5", borderRadius: 6, marginBottom: 12, fontSize: 14 },
  tabs:      { display: "flex", alignItems: "center", gap: 4, borderBottom: "1px solid #e5e7eb", marginBottom: 20, paddingBottom: 0 },
  tab:       { padding: "8px 16px", border: "none", background: "none", cursor: "pointer", fontSize: 14, color: "#6b7280", borderBottom: "2px solid transparent" },
  tabActive: { color: "#2563eb", borderBottom: "2px solid #2563eb", fontWeight: 600 },
  grid:      { display: "grid", gridTemplateColumns: "repeat(3,1fr)", gap: 16 },
  panel:     { border: "1px solid #e5e7eb", borderRadius: 8, padding: 16, background: "#f9fafb" },
  panelHeader: { display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 4 },
  panelTitle:  { fontWeight: 600, fontSize: 14 },
  panelSub:    { fontSize: 12, color: "#6b7280", marginBottom: 12 },
  badge:     { fontSize: 11, color: "#fff", padding: "2px 8px", borderRadius: 99 },
  msgList:   { display: "flex", flexDirection: "column", gap: 6, maxHeight: 300, overflowY: "auto" },
  msg:       { padding: "6px 10px", background: "#fff", border: "1px solid #e5e7eb", borderRadius: 6, fontSize: 13 },
  empty:     { fontSize: 13, color: "#9ca3af" },
  dagBox:    { border: "1px solid #e5e7eb", borderRadius: 8, padding: 24, background: "#f9fafb" },
  ruleRow:   { display: "flex", alignItems: "center", gap: 12 },
  ruleBox:   { padding: "8px 14px", borderRadius: 6, border: "1px solid #d1d5db", background: "#fff", fontSize: 13 },
  ruleLabel: { fontSize: 11, color: "#6b7280", marginRight: 8 },
  arrow:     { fontSize: 18, color: "#6b7280" },
  code:      { fontFamily: "monospace", fontSize: 13 },
  svcChip:   { padding: "4px 10px", borderRadius: 99, background: "#eff6ff", border: "1px solid #93c5fd", fontSize: 13 },
  chipSub:   { color: "#6b7280", fontSize: 11, marginLeft: 6 },
  uploadBtn: { padding: "7px 14px", borderRadius: 6, background: "#f3f4f6", border: "1px solid #d1d5db", fontSize: 13, cursor: "pointer" },
};
