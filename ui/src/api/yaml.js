// Minimal DAG-to-YAML serializer.
//
// The Control Plane stores DAGs relationally and reassembles them to YAML on read
// (architecture.md). Here we do the same reassembly client-side from the mock
// node objects, producing the topology-only body (name / active / RBAC are
// control-plane-side and deliberately omitted, per the schema docs). Scope is
// intentionally narrow: it only needs to render the shapes in our mock data.

function whenToInline(when) {
  return `{field: ${when.field}, op: ${when.op}, value: ${JSON.stringify(when.value)}}`;
}

export function dagToYaml(dag) {
  const lines = ['nodes:'];

  for (const node of dag.nodes) {
    lines.push(`  - name: ${node.name}`);
    lines.push(`    topic: ${node.topic}`);
    lines.push(`    host: ${node.host}`);
    lines.push(`    port: ${node.port}`);
    if (node.entry) lines.push('    entry: true');
    if (Array.isArray(node.tools)) {
      lines.push(`    tools: [${node.tools.join(', ')}]`);
    }

    if (!node.routes || node.routes.length === 0) {
      lines.push('    routes: []');
    } else {
      lines.push('    routes:');
      for (const route of node.routes) {
        if (route.when) {
          lines.push(`      - when: ${whenToInline(route.when)}`);
          lines.push(`        to: ${route.to}`);
        } else {
          lines.push(`      - to: ${route.to}`);
        }
      }
    }
    lines.push('');
  }

  return lines.join('\n').trimEnd() + '\n';
}
