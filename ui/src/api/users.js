// User & RBAC admin queries + mutations (mock-backed).
// Mirrors the Control Plane user endpoints in design.md §8.

import usersData from '../mocks/users.json';
import { mock } from './client.js';

let users = structuredClone(usersData);

export function listUsers() {
  return mock(users);
}

export function setUserRole(username, role) {
  const user = users.find((u) => u.username === username);
  if (user) user.role = role;
  return mock(user ?? null);
}

export function setUserStatus(username, status) {
  const user = users.find((u) => u.username === username);
  if (user) user.status = status;
  return mock(user ?? null);
}
