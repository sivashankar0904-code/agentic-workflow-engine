// Session / identity handling (mock-backed).
//
// design.md §3, §8: the current identity drives RBAC-gated UI (hiding actions the
// caller can't perform) and the admin route guard. With no backend yet, we read a
// fixed identity from mocks/session.json. Real auth would populate this from a
// token/session against the Control Plane.

import sessionData from '../mocks/session.json';

export function getSession() {
  return sessionData;
}

export function hasRole(role) {
  return sessionData.roles.includes(role);
}

export function isAdmin() {
  return hasRole('admin');
}
