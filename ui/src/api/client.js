// Mock API client.
//
// This mirrors the shape of the real client described in design.md §9-10: a thin
// wrapper that would attach an auth header and hit the Control Plane (:9000) and
// the execution ingress. For now there is NO backend integration — every call
// resolves from the bundled mock JSON in ../mocks after a small simulated delay,
// so pages exercise real loading states. Swapping to real HTTP later means
// changing only this file (and the three api/*.js modules that call `mock()`).

const LATENCY_MS = 220;

// Return a deep clone so callers can mutate optimistic copies without touching
// the imported JSON module (which is shared across the app).
function clone(value) {
  return structuredClone(value);
}

export function mock(data) {
  return new Promise((resolve) => {
    setTimeout(() => resolve(clone(data)), LATENCY_MS);
  });
}

// Placeholder for the future real transport. Base URLs live here so pages never
// hard-code them.
export const BASE_URLS = {
  controlPlane: 'http://localhost:9000',
  ingress: 'http://localhost:9000/ingress',
};
