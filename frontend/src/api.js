// This is the ONLY place the frontend talks to the backend.
//
// It used to use fetch() over REST. Now it talks to NATS directly over a
// WebSocket (the browser cannot speak the native NATS TCP protocol):
//   - commands use request-reply  (nc.request) so we still get an answer
//   - live updates use a subscription on todo.event.>
//
// App.jsx did NOT have to change for this swap — that's the payoff of
// keeping every network call in one file.
import { connect, JSONCodec } from 'nats.ws'

// The NATS server's WebSocket port (see nats.conf). Override with
// VITE_NATS_URL in a .env file to point at a deployed broker.
const SERVER = import.meta.env.VITE_NATS_URL ?? 'ws://localhost:8080'

const jc = JSONCodec()

// Connect lazily and reuse one connection for the whole app.
let ncPromise = null
function conn() {
  if (!ncPromise) {
    ncPromise = connect({ servers: SERVER }).catch((e) => {
      ncPromise = null // allow a later retry if the first connect failed
      throw e
    })
  }
  return ncPromise
}

// request sends one command and waits for the service's reply.
async function request(subject, payload) {
  const nc = await conn()
  const msg = await nc.request(subject, jc.encode(payload ?? {}), { timeout: 5000 })
  const reply = jc.decode(msg.data)
  if (!reply || reply.ok !== true) {
    throw new Error((reply && reply.error) || 'request failed')
  }
  return reply.data
}

export function listTodos() {
  return request('todo.cmd.list')
}

export function createTodo(title) {
  return request('todo.cmd.create', { title })
}

export function updateTodo(id, { title, done }) {
  return request('todo.cmd.update', { id, title, done })
}

export function deleteTodo(id) {
  return request('todo.cmd.delete', { id })
}

// onTodoEvent subscribes to todo.event.* and calls handler(subject, event)
// for every change made by ANY client. Returns the subscription so the
// caller can unsubscribe. This is what powers live, cross-tab updates.
export async function onTodoEvent(handler) {
  const nc = await conn()
  const sub = nc.subscribe('todo.event.>')
  ;(async () => {
    for await (const m of sub) {
      try {
        handler(m.subject, jc.decode(m.data))
      } catch {
        // ignore a malformed event
      }
    }
  })()
  return sub
}
