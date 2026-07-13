import { useMemo } from 'react'
import { clusterOf, layoutDag } from './layout.js'
import ToolChips from './ToolChips.jsx'
import './graph.css'

// SVG DAG renderer (design.md §5). Consumes the standardized nodes[]/routes[]
// contract, lays out topological ranks, and draws rounded-rect nodes with labeled
// edges, START/END markers, tool chips, and subtle ft_*/ag_* cluster tinting.
//
// Reused by the Run console (§6): pass `activePath` (node names, in order) and
// `activeIndex` to highlight the path a message takes.
export default function DagGraph({ dag, activePath = null, activeIndex = -1 }) {
  const { nodes, edges, width, height, NODE_H } = useMemo(
    () => layoutDag(dag),
    [dag]
  )

  const activeSet = new Set(
    activePath ? activePath.slice(0, activeIndex + 1) : []
  )
  const isEdgeActive = (e) =>
    activeSet.has(e.from) && activeSet.has(e.to)

  return (
    <div className="graph-scroll">
      <svg
        className="dag-graph"
        viewBox={`0 0 ${width} ${height}`}
        width={width}
        height={height}
        role="img"
        aria-label={`Graph for ${dag.name}`}
      >
        <defs>
          <marker
            id="arrow"
            viewBox="0 0 10 10"
            refX="9"
            refY="5"
            markerWidth="7"
            markerHeight="7"
            orient="auto-start-reverse"
          >
            <path d="M0,0 L10,5 L0,10 z" className="arrow-head" />
          </marker>
          <marker
            id="arrow-active"
            viewBox="0 0 10 10"
            refX="9"
            refY="5"
            markerWidth="7"
            markerHeight="7"
            orient="auto-start-reverse"
          >
            <path d="M0,0 L10,5 L0,10 z" className="arrow-head-active" />
          </marker>
        </defs>

        {/* Edges */}
        {edges.map((e, i) => {
          const midX = (e.x1 + e.x2) / 2
          const active = isEdgeActive(e)
          const path = `M ${e.x1} ${e.y1} C ${midX} ${e.y1}, ${midX} ${e.y2}, ${
            e.x2
          } ${e.y2}`
          return (
            <g key={i} className={active ? 'edge active' : 'edge'}>
              <path
                d={path}
                markerEnd={`url(#${active ? 'arrow-active' : 'arrow'})`}
              />
              {e.label && (
                <text
                  x={midX}
                  y={(e.y1 + e.y2) / 2 - 6}
                  textAnchor="middle"
                  className="edge-label"
                >
                  {e.label}
                </text>
              )}
            </g>
          )
        })}

        {/* Nodes */}
        {nodes.map((n) => {
          const cluster = clusterOf(n.name)
          const active = activeSet.has(n.name)
          const isCurrent =
            activePath && activePath[activeIndex] === n.name
          const hasTools = Array.isArray(n.tools)
          return (
            <g
              key={n.name}
              className={`node ${cluster ? `cluster-${cluster}` : ''} ${
                active ? 'active' : ''
              } ${isCurrent ? 'current' : ''}`}
              transform={`translate(${n.x}, ${n.y})`}
            >
              {n.entry && (
                <text x={-6} y={n.h / 2 + 4} textAnchor="end" className="marker">
                  START ▸
                </text>
              )}
              <rect
                width={n.w}
                height={n.h}
                rx="10"
                className="node-box"
              />
              <text x={12} y={22} className="node-name">
                {n.name}
              </text>
              <text x={12} y={39} className="node-meta">
                {n.host}:{n.port}
              </text>
              <text x={12} y={53} className="node-topic">
                ⊙ {n.topic}
              </text>
              {hasTools && (
                <ToolChips
                  tools={n.tools}
                  x={12}
                  y={n.h - 20}
                  width={n.w - 20}
                />
              )}
              {n.terminal && (
                <text
                  x={n.w + 6}
                  y={n.h / 2 + 4}
                  textAnchor="start"
                  className="marker"
                >
                  ▪ END
                </text>
              )}
            </g>
          )
        })}
      </svg>
    </div>
  )
}
