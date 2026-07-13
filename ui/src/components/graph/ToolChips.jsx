// Renders an agent node's tools[] as chips inside the SVG node (design.md §7).
// Uses foreignObject so chips reuse the same CSS as the rest of the app.
export default function ToolChips({ tools, x, y, width }) {
  if (!tools || tools.length === 0) return null
  return (
    <foreignObject x={x} y={y} width={width} height={22}>
      <div
        className="chip-row"
        style={{ overflow: 'hidden', flexWrap: 'nowrap' }}
      >
        {tools.map((t) => (
          <span key={t} className="chip" style={{ fontSize: 10 }}>
            {t}
          </span>
        ))}
      </div>
    </foreignObject>
  )
}
