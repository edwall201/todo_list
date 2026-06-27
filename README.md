# TODO List — React + Go (RESTful API)

A simple TODO list app. **Phase 1 (this repo): a classic RESTful API.**
Later phases will swap the transport to **NATS / event-driven / CloudEvents**
so you can feel the difference between request/response (REST) and
publish/subscribe (messaging).

```
┌──────────────┐   HTTP/JSON (REST)   ┌──────────────┐   in-memory   ┌──────────┐
│  React app   │ ───────────────────► │ TODO service │ ────────────► │  store   │
│ (frontend)   │ ◄─────────────────── │   (Go)       │ ◄──────────── │ ("DB")   │
└──────────────┘                      └──────────────┘               └──────────┘
   :5173                                  :8080
```

The store is currently **in-memory**, so todos reset when the Go server
restarts. It lives behind four methods (`List / Create / Update / Delete`)
in `backend/store.go`, so you can later replace it with SQLite/Postgres
without changing any handler.

## Features

- **Add** a task (`POST /api/todos`)
- **Modify** a task's title (`PUT /api/todos/{id}`) — double-click the text or the ✎ button
- **Finish** a task / toggle done (`PUT /api/todos/{id}`) — click the square box
- **Delete** a task (`DELETE /api/todos/{id}`) — the ✕ button

## REST API

| Method | Path              | Body                          | Purpose            |
|--------|-------------------|-------------------------------|--------------------|
| GET    | `/api/todos`      | —                             | List all todos     |
| POST   | `/api/todos`      | `{ "title": "..." }`          | Create a todo      |
| PUT    | `/api/todos/{id}` | `{ "title": "...", "done": false }` | Update title/done  |
| DELETE | `/api/todos/{id}` | —                             | Delete a todo      |
| GET    | `/api/health`     | —                             | Health check       |

## Requirements

- **Go** 1.22 or newer (uses the standard-library method router; no external deps)
- **Node.js** 18 or newer + npm

## Run it

Open **two terminals**.

### 1) Backend (Go)

```bash
cd backend
go run .
# -> TODO service listening on http://localhost:8080
```

Quick test without the frontend:

```bash
curl localhost:8080/api/todos
curl -X POST localhost:8080/api/todos -H 'Content-Type: application/json' -d '{"title":"Buy milk"}'
```

### 2) Frontend (React + Vite)

```bash
cd frontend
npm install      # first time only
npm run dev
# -> open http://localhost:5173
```

The Vite dev server proxies `/api/*` to `http://localhost:8080`, so you
don't have to deal with CORS or ports during development.

## Project layout

```
todo_list/
├── backend/
│   ├── go.mod
│   ├── main.go        # server setup, routes, CORS, sample data
│   ├── handlers.go    # one function per REST endpoint
│   ├── store.go       # in-memory "DB" (swap for SQL later)
│   └── models.go      # the Todo struct
└── frontend/
    ├── package.json
    ├── vite.config.js # dev server + /api proxy
    ├── index.html
    └── src/
        ├── main.jsx   # React entry point
        ├── App.jsx    # the whole UI (add / edit / finish / delete)
        ├── api.js     # the ONLY file that calls the backend
        └── index.css  # notepad styling
```

## Next step: REST vs NATS

When you move to NATS, the **store and the React UI barely change** — what
changes is the line *between* them:

- **REST (now):** the frontend asks the service a question and waits for the
  answer. One caller, one responder, synchronous.
- **NATS (next):** the service *publishes* an event like `todo.created`, and
  any number of subscribers react to it. No one waits. This is how you get
  things like live updates, audit logs, or notifications "for free."

Because every network call is isolated in `frontend/src/api.js` and every
data operation is isolated in `backend/store.go`, those are the two files
you'll focus on when introducing messaging.
