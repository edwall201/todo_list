// This file is the ONLY place the frontend talks to the backend.
// Keeping all fetch() calls here means that when you later switch from
// REST to NATS/event-driven, you only rewrite this one file.

// Empty base => use the Vite dev proxy ("/api" -> localhost:8080).
// To point at a deployed backend, set VITE_API_URL in a .env file.
const BASE = import.meta.env.VITE_API_URL ?? ''

async function handle(res) {
  if (!res.ok) {
    let message = `Request failed (${res.status})`
    try {
      const data = await res.json()
      if (data.error) message = data.error
    } catch {
      // body wasn't JSON; keep the default message
    }
    throw new Error(message)
  }
  if (res.status === 204) return null // No Content (e.g. after DELETE)
  return res.json()
}

export function listTodos() {
  return fetch(`${BASE}/api/todos`).then(handle)
}

export function createTodo(title) {
  return fetch(`${BASE}/api/todos`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ title }),
  }).then(handle)
}

export function updateTodo(id, { title, done }) {
  return fetch(`${BASE}/api/todos/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ title, done }),
  }).then(handle)
}

export function deleteTodo(id) {
  return fetch(`${BASE}/api/todos/${id}`, { method: 'DELETE' }).then(handle)
}
