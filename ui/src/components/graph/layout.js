// Topological-rank layout for the DAG graph (design.md §5, §12a).
//
// The schema is a graph of nodes with per-node routes[] (multi-stage,
// branch-and-rejoin), so we lay out by topological rank — NOT two fixed columns.
// Each node's rank is the longest path from an entry node; nodes at the same rank
// stack vertically. Edges are returned with their source/target geometry and the
// route label (`field op "value"`).

const NODE_W = 176;
const NODE_H = 62;
const RANK_GAP = 88; // horizontal gap between ranks
const ROW_GAP = 26; // vertical gap within a rank
const PAD = 24;

export function layoutDag(dag) {
  const nodes = dag.nodes;
  const byName = new Map(nodes.map((n) => [n.name, n]));

  // Longest-path rank via memoized DFS (graph is acyclic by contract).
  const rankCache = new Map();
  function rankOf(name, seen = new Set()) {
    if (rankCache.has(name)) return rankCache.get(name);
    if (seen.has(name)) return 0; // defensive against unexpected cycles
    seen.add(name);
    const node = byName.get(name);
    const incoming = nodes.filter((n) =>
      (n.routes || []).some((r) => r.to === name)
    );
    const rank = incoming.length
      ? Math.max(...incoming.map((n) => rankOf(n.name, new Set(seen)) + 1))
      : 0;
    rankCache.set(name, rank);
    return rank;
  }

  const ranks = new Map(); // rank -> node names
  for (const node of nodes) {
    const r = node.entry ? 0 : rankOf(node.name);
    if (!ranks.has(r)) ranks.set(r, []);
    ranks.get(r).push(node.name);
  }

  // Assign coordinates.
  const pos = new Map();
  const sortedRanks = [...ranks.keys()].sort((a, b) => a - b);
  let maxCol = 0;
  for (const r of sortedRanks) {
    const col = ranks.get(r);
    col.forEach((name, i) => {
      pos.set(name, {
        x: PAD + r * (NODE_W + RANK_GAP),
        y: PAD + i * (NODE_H + ROW_GAP),
      });
    });
    maxCol = Math.max(maxCol, col.length);
  }

  const laidNodes = nodes.map((n) => ({
    ...n,
    x: pos.get(n.name).x,
    y: pos.get(n.name).y,
    w: NODE_W,
    h: NODE_H,
    terminal: !n.routes || n.routes.length === 0,
  }));

  // Build edges with endpoint geometry and labels.
  const edges = [];
  for (const node of nodes) {
    const src = pos.get(node.name);
    for (const route of node.routes || []) {
      const dst = pos.get(route.to);
      if (!dst) continue;
      edges.push({
        from: node.name,
        to: route.to,
        x1: src.x + NODE_W,
        y1: src.y + NODE_H / 2,
        x2: dst.x,
        y2: dst.y + NODE_H / 2,
        label: route.when
          ? `${route.when.field} ${opSymbol(route.when.op)} ${JSON.stringify(
              route.when.value
            )}`
          : null,
      });
    }
  }

  const width =
    PAD * 2 + (sortedRanks.length ? sortedRanks[sortedRanks.length - 1] : 0) *
      (NODE_W + RANK_GAP) +
    NODE_W;
  const height = PAD * 2 + maxCol * (NODE_H + ROW_GAP) - ROW_GAP;

  return { nodes: laidNodes, edges, width, height, NODE_W, NODE_H };
}

function opSymbol(op) {
  switch (op) {
    case 'equals':
      return '==';
    case 'contains':
      return '∋';
    case 'notEquals':
      return '!=';
    default:
      return op;
  }
}

// Name-prefix cluster (ft_* / ag_*) for subtle tinting — a rendering convention,
// not a schema construct (design.md §5a).
export function clusterOf(name) {
  if (name.startsWith('ft_')) return 'ft';
  if (name.startsWith('ag_')) return 'ag';
  return null;
}
