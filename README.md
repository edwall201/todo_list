# TODO List — React + Go over NATS (event-driven)

A simple TODO list app, built in phases to compare transports.

- **Phase 1 — REST.** Browser ⇄ service over HTTP request/response.
- **Phase 2 — REST + events.** Same REST, but the service also published NATS events.
- **Phase 3 — all NATS (this version).** REST/HTTP is **gone**. The browser talks to the backend **only over NATS** (via WebSocket). Commands use **request-reply**; changes are broadcast as **events**, which also drive **live updates across browser tabs**.

```
                            request-reply (todo.cmd.*)
┌────────────┐  WebSocket  ┌──────────────────────┐   TCP   ┌──────────────┐
│ React app  │ ──────────► │     NATS server      │ ◄─────► │ todo-service │
│ (nats.ws)  │ ◄────────── │  :4222 tcp / :8080 ws│         │    (Go)      │
└────────────┘   events    └──────────────────────┘         └──────────────┘
   :5173      (todo.event.*)        ▲                              │ publish
                                    │ subscribe todo.event.>       │ todo.event.*
                                    │                              ▼
                            ┌──────────────────┐            (broadcast to all)
                            │  todo-subscriber │
                            └──────────────────┘
```

The store is still **in-memory**, so todos reset when the Go service restarts.

## How the browser talks to NATS

A browser can't speak the native NATS TCP protocol, so:

1. The NATS server is run with **WebSocket** enabled (see `nats.conf`, port 8080).
2. The frontend uses **`nats.ws`** to connect over `ws://localhost:8080`.
3. Two messaging patterns are used:
   - **Commands → request-reply.** The browser `request()`s a subject and waits for the reply. This is what replaces REST.
   - **Events → publish/subscribe.** The service broadcasts every change; the browser (and the standalone subscriber) listen and react.

## Subjects

**Commands (request-reply):** the browser sends, the service replies.

| Subject           | Body                          | Reply (`data`)      |
|-------------------|-------------------------------|---------------------|
| `todo.cmd.list`   | `{}`                          | array of todos      |
| `todo.cmd.create` | `{ "title": "..." }`          | the new todo        |
| `todo.cmd.update` | `{ "id", "title", "done" }`   | the updated todo    |
| `todo.cmd.delete` | `{ "id": 1 }`                 | the deleted todo    |

Because NATS has no HTTP status codes, every reply uses an envelope:

```json
{ "ok": true,  "data": { ... } }
{ "ok": false, "error": "title is required" }
```

**Events (publish/subscribe):** the service announces, anyone can listen. Match all with `todo.event.>`.

| Subject               | Published when…                |
|-----------------------|--------------------------------|
| `todo.event.created`  | a todo is created              |
| `todo.event.updated`  | a todo is modified or finished |
| `todo.event.deleted`  | a todo is deleted              |

Each event is a CloudEvents v1.0 JSON envelope (`specversion`, `id`, `source`, `type`, `time`, `data`).

## Requirements

- **Go** 1.25 or newer
- **Node.js** 18 or newer + npm
- A **NATS server** — now **required** (it's the only transport), and it must have WebSocket enabled. The included `nats.conf` does that.

## Run it

### 1) NATS server (with WebSocket)

```bash
# Docker (mounts the included config)
docker run --rm -p 4222:4222 -p 8080:8080 -p 8222:8222 \
  -v "$PWD/nats.conf:/nats.conf" nats:latest -c /nats.conf

# or, with a locally installed nats-server
nats-server -c nats.conf
```

You should see `Listening for websocket clients on ws://0.0.0.0:8080`.

### 2) Backend — the TODO service

```bash
cd backend
go run .
# -> todo-service connected to NATS at nats://127.0.0.1:4222
# -> listening for commands on todo.cmd.*  /  broadcasting events on todo.event.*
```

### 3) The event subscriber (a second terminal, optional)

```bash
cd backend
go run ./subscriber
# -> todo-subscriber listening for "todo.event.>"
```

### 4) Frontend — React + Vite (a third terminal)

```bash
cd frontend
npm install      # first time only (adds nats.ws)
npm run dev
# -> open http://localhost:5173
```

**Try the payoff:** open the app in **two browser tabs**. Add or finish a task in
one — the other updates instantly, because both subscribe to `todo.event.>`.

## Project layout

```
todo_list/
├── nats.conf               # NATS server config (enables WebSocket on 8080)
├── backend/
│   ├── go.mod
│   ├── main.go             # connect NATS, subscribe to todo.cmd.*, block
│   ├── handlers.go         # Service: NATS command handlers (request-reply)
│   ├── store.go            # in-memory "DB"
│   ├── models.go           # the Todo struct
│   ├── events.go           # publishEvent: CloudEvent on todo.event.*
│   └── subscriber/
│       └── main.go         # standalone logger of todo.event.>
└── frontend/
    ├── package.json        # depends on nats.ws
    ├── vite.config.js      # no /api proxy any more
    ├── index.html
    └── src/
        ├── main.jsx
        ├── App.jsx         # UI; subscribes to events for live updates
        ├── api.js          # NATS client (request-reply + subscribe)
        └── index.css
```

## What it took to drop REST (the honest trade-offs)

Going all-NATS for a browser app is very doable, but you pay for what REST gave free:

- **You must run a broker** — there's no "just hit the URL". No NATS, no app.
- **The browser needs WebSocket** + a NATS client library (`nats.ws`), so the JS bundle is larger.
- **You re-invent request/response** — commands need request-reply, and since there
  are no HTTP status codes you carry success/failure in your own `{ok,error}` envelope.

What you gain:

- **Live updates for free** — broadcasting `todo.event.*` means every tab/client stays in sync.
- **Easy fan-out** — add another subscriber (analytics, notifications) without touching the service.

Note how little moved: on the frontend **only `src/api.js` changed**; `App.jsx` just
added one subscription. On the backend the handlers went from `(w, r)` to `(*nats.Msg)`,
but the store and the Todo model are untouched.
