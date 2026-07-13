// Run console: submit a workflow message + poll its progress (mock-backed).
// Mirrors design.md §6 — post to the ingress, then observe the message moving
// node-to-node. Here we replay a canned path from mocks/runs.json.

import runsData from '../mocks/runs.json';
import { mock } from './client.js';

// The suggested sample message to prefill for a given DAG.
export function sampleFor(dagName) {
  return runsData[dagName]?.sampleMessage ?? '';
}

// The ordered path a submitted message takes through the DAG's nodes, plus the
// terminal result. A real ingress would stream this; we return the plan and let
// the page animate it.
export function submitMessage(dagName /*, message */) {
  const run = runsData[dagName];
  if (!run) return mock({ path: [], result: { status: 'unroutable', output: null } });
  return mock({ path: run.path, result: run.result });
}
