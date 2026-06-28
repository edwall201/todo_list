# TODO List — React + Go (REST + NATS)

A simple TODO list app, built in phases to compare transports.

- **Phase 1 — RESTful API.** Request/response over HTTP. The browser asks, the service answers.
- **Phase 2 — NATS events (this version).** The same REST API still serves the browser, but now the service *also* publishes a [CloudEvent](https://cloudevents.io) to NATS after every change. A separate subscriber reacts to those events. Nobody waits.

```
                                            publish todo.*        ┌──────────────────┐
┌────────────┐   HTTP/JSON (REST)   ┌──────────────┐   event   ┌─►│ todo-subscriber  │ logs events
│  React app │ ───────────────────► │ TODO service │ ─────────►│  │  (separate proc) │
│ (frontend) │ ◄─────────────────── │   (Go)       │           │  └──────────────────┘
└────────────┘                      └──────────────┘           │
   :5173                              :8080  │ publish          │ subscribe "todo.>"
                                            ▼                   │
                                     ┌───────────────────────────────┐
                                     │          NATS server          │
                                     │            :4222              │
                                     └───────────────────────────────┘
```

The store is **in-memory**, so todos reset when the Go service restarts.
NATS publishing is **fire-and-forget**: if the broker is down, the REST API
still works — events are just skipped.

## Features

- **Add** a task (`POST /api/todos`)
- **Modify** a task's title (`PUT /api/todos/{id}`) — double-click the text or the ✎ button
- **Finish** a task / toggle done (`PUT /api/todos/{id}`) — click the square box
- **Delete** a task (`DELETE /api/todos/{id}`) — the ✕ button

Each of the last three also publishes a NATS event.

## REST API

| Method | Path              | Body                                | Purpose            |
|--------|-------------------|-------------------------------------|--------------------|
| GET    | `/api/todos`      | —                                   | List all todos     |
| POST   | `/api/todos`      | `{ "title": "..." }`                | Create a todo      |
| PUT    | `/api/todos/{id}` | `{ "title": "...", "done": false }` | Update title/done  |
| DELETE | `/api/todos/{id}` | —                                   | Delete a todo      |
| GET    | `/api/health`     | —                                   | Health check       |

## Events (NATS)

| Subject        | Published when…       |
|----------------|-----------------------|
| `todo.created` | a todo is created     |
| `todo.updated` | a todo is modified or finished |
| `todo.deleted` | a todo is deleted     |

Subscribe to all of them with the wildcard `todo.>`.

Each message is a CloudEvents v1.0 JSON envelope:

```json
{
  "specversion": "1.0",
  "id": "f3a9c1e2b4d5...",
  "source": "/todo-service",
  "type": "com.example.todo.created",
  "time": "2026-06-28T11:44:21Z",
  "datacontenttype": "application/json",
  "data": { "id": 4, "title": "Learn NATS", "done": false, "created_at": "2026-06-28T11:44:21Z" }
}
```

## Requirements

- **Go** 1.25 or newer (the `nats.go` client requires it)
- **Node.js** 18 or newer + npm
- A **NATS server** (only needed to see events; the app runs without it)

## Run it

### 1) Start a NATS server

Pick whichever is easiest:

```bash
# Docker (most portable)
docker run --rm -p 4222:4222 nats:latest

# macOS / Homebrew
brew install nats-server && nats-server

# Go (installs the binary to ~/go/bin)
go install github.com/nats-io/nats-server/v2@latest && ~/go/bin/nats-server
```

It listens on `nats://127.0.0.1:4222` by default. To point the app elsewhere,
set `NATS_URL` for both the service and the subscriber.

### 2) Backend — the TODO service

```bash
cd backend
go run .
# -> connected to NATS at nats://127.0.0.1:4222 -- publishing todo.* events
# -> TODO service listening on http://localhost:8080
```

### 3) The event subscriber (a second terminal)

```bash
cd backend
go run ./subscriber
# -> todo-subscriber listening on nats://127.0.0.1:4222 for "todo.>"
```

### 4) Frontend — React + Vite (a third terminal)

```bash
cd frontend
npm install      # first time only
npm run dev
# -> open http://localhost:5173
```

Now add / finish / edit / delete tasks in the browser and watch the events
appear in the subscriber's terminal in real time. Or test with curl:

```bash
curl -X POST localhost:8080/api/todos -H 'Content-Type: application/json' -d '{"title":"Buy milk"}'
```

> Note: a fresh checkout has no `go.sum`. The first `go run`/`go build`
> downloads dependencies automatically; you can also run `go mod tidy` to
> generate it explicitly.

## Project layout

```
todo_list/
├── backend/
│   ├── go.mod
│   ├── main.go            # server setup, routes, CORS, sample data
│   ├── handlers.go        # one function per REST endpoint (+ publish)
│   ├── store.go           # in-memory "DB"
│   ├── models.go          # the Todo struct
│   ├── events.go          # NATS publisher + CloudEvents envelope
│   └── subscriber/
│       └── main.go        # standalone program that logs todo.* events
└── frontend/
    ├── package.json
    ├── vite.config.js     # dev server + /api proxy
    ├── index.html
    └── src/
        ├── main.jsx
        ├── App.jsx        # the whole UI (add / edit / finish / delete)
        ├── api.js         # the ONLY file that calls the backend
        └── index.css      # notepad styling
```

## REST vs NATS — what actually changed

The browser still talks to the service over REST. What's new is a *second*,
one-way channel:

- **REST** is a conversation: the caller waits for a reply, and exactly one
  service handles each request.
- **NATS** is a broadcast: the service announces "a todo was created" and
  walks away. Zero, one, or many subscribers can react — and you can add or
  restart subscribers without touching the publisher.

That decoupling is the payoff. The `todo-subscriber` shares no code and no
memory with the service; only the subject name (`todo.created`) connects them.

## Possible next steps

- **JetStream** — turn on NATS persistence so events are stored and can be
  replayed (a subscriber that starts later still sees past events).
- **Live UI updates** — expose NATS over WebSocket and subscribe from React,
  so a change in one browser tab instantly updates every other tab.
