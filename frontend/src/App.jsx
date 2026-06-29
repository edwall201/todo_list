import { useEffect, useState } from 'react'
import * as api from './api'

// Decorative row of sparkles at the top, like the reference image.
function Sparkles() {
  return (
    <svg
      className="sparkles"
      viewBox="0 0 240 40"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      aria-hidden="true"
    >
      <circle cx="20" cy="20" r="3" fill="currentColor" />
      <line x1="24" y1="20" x2="54" y2="20" stroke="currentColor" strokeWidth="1" />
      <path d="M70 5 L74 16 L85 20 L74 24 L70 35 L66 24 L55 20 L66 16 Z" fill="currentColor" />
      <path d="M120 0 L125 15 L140 20 L125 25 L120 40 L115 25 L100 20 L115 15 Z" fill="currentColor" />
      <path d="M170 6 L173 17 L184 20 L173 23 L170 34 L167 23 L156 20 L167 17 Z" fill="currentColor" />
      <line x1="186" y1="20" x2="216" y2="20" stroke="currentColor" strokeWidth="1" />
      <circle cx="220" cy="20" r="3" fill="currentColor" />
    </svg>
  )
}

// A small checkmark drawn as SVG (shown inside a finished task's box).
function CheckMark() {
  return (
    <svg viewBox="0 0 24 24" width="14" height="14" aria-hidden="true">
      <path
        d="M5 12l5 5L20 7"
        stroke="currentColor"
        strokeWidth="3"
        fill="none"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  )
}

export default function App() {
  const [todos, setTodos] = useState([])
  const [newTitle, setNewTitle] = useState('')
  const [editingId, setEditingId] = useState(null)
  const [editingText, setEditingText] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)

  // Load the list once when the page first renders.
  useEffect(() => {
    refresh()
  }, [])

  // Live updates: when ANY client changes a todo, the service broadcasts
  // a NATS event on todo.event.*. We just re-fetch so every open tab
  // stays in sync — this is the payoff of going fully event-driven.
  useEffect(() => {
    let sub
    api
      .onTodoEvent(() => refresh())
      .then((s) => {
        sub = s
      })
      .catch(() => {})
    return () => {
      if (sub) sub.unsubscribe()
    }
  }, [])

  async function refresh() {
    try {
      setLoading(true)
      setTodos(await api.listTodos())
      setError('')
    } catch (e) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }

  // ADD
  async function handleAdd(e) {
    e.preventDefault()
    const title = newTitle.trim()
    if (!title) return
    try {
      const todo = await api.createTodo(title)
      setTodos([todo, ...todos])
      setNewTitle('')
    } catch (e) {
      setError(e.message)
    }
  }

  // FINISH (toggle the done flag)
  async function toggleDone(todo) {
    try {
      const updated = await api.updateTodo(todo.id, { title: todo.title, done: !todo.done })
      setTodos(todos.map((t) => (t.id === todo.id ? updated : t)))
    } catch (e) {
      setError(e.message)
    }
  }

  // MODIFY (edit the title)
  function startEdit(todo) {
    setEditingId(todo.id)
    setEditingText(todo.title)
  }

  async function saveEdit(todo) {
    const title = editingText.trim()
    if (!title) {
      setEditingId(null)
      return
    }
    try {
      const updated = await api.updateTodo(todo.id, { title, done: todo.done })
      setTodos(todos.map((t) => (t.id === todo.id ? updated : t)))
    } catch (e) {
      setError(e.message)
    } finally {
      setEditingId(null)
    }
  }

  // DELETE
  async function handleDelete(id) {
    try {
      await api.deleteTodo(id)
      setTodos(todos.filter((t) => t.id !== id))
    } catch (e) {
      setError(e.message)
    }
  }

  // Pad with blank lines so it always looks like the lined notepad.
  const minRows = 8
  const emptyRows = Math.max(0, minRows - todos.length)

  return (
    <div className="page">
      <div className="paper">
        <Sparkles />
        <h1 className="title">TO DO LIST</h1>

        <form className="add-row" onSubmit={handleAdd}>
          <input
            className="add-input"
            placeholder="Add a task and press Enter…"
            value={newTitle}
            onChange={(e) => setNewTitle(e.target.value)}
          />
          <button type="submit" className="add-btn" aria-label="Add task">
            +
          </button>
        </form>

        {error && <div className="error">{error}</div>}
        {loading && <div className="hint">Loading…</div>}

        <ul className="list">
          {todos.map((todo) => (
            <li key={todo.id} className={`item ${todo.done ? 'done' : ''}`}>
              <button
                className={`check ${todo.done ? 'checked' : ''}`}
                onClick={() => toggleDone(todo)}
                aria-label={todo.done ? 'Mark as not done' : 'Mark as done'}
              >
                {todo.done && <CheckMark />}
              </button>

              {editingId === todo.id ? (
                <input
                  className="edit-input"
                  autoFocus
                  value={editingText}
                  onChange={(e) => setEditingText(e.target.value)}
                  onBlur={() => saveEdit(todo)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') saveEdit(todo)
                    if (e.key === 'Escape') setEditingId(null)
                  }}
                />
              ) : (
                <span className="text" onDoubleClick={() => startEdit(todo)}>
                  {todo.title}
                </span>
              )}

              <div className="actions">
                <button className="icon-btn" onClick={() => startEdit(todo)} aria-label="Edit task">
                  ✎
                </button>
                <button
                  className="icon-btn"
                  onClick={() => handleDelete(todo.id)}
                  aria-label="Delete task"
                >
                  ✕
                </button>
              </div>
            </li>
          ))}

          {Array.from({ length: emptyRows }).map((_, i) => (
            <li key={`empty-${i}`} className="item empty">
              <span className="check" />
              <span className="text" />
            </li>
          ))}
        </ul>
      </div>
    </div>
  )
}
