package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

// API holds the dependencies the HTTP handlers need: the store ("DB")
// and the event publisher (NATS).
type API struct {
	store *Store
	pub   *EventPublisher
}

// --- small helpers -------------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// --- handlers (one per REST endpoint) ------------------------------------
//
// Each mutating handler does its REST work first, then publishes a NATS
// event. The HTTP response does NOT depend on the event: publishing is
// fire-and-forget, so a slow or missing broker never blocks the user.

// GET /api/todos  ->  list every todo
func (a *API) listTodos(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.store.List())
}

// POST /api/todos  ->  create a todo from {"title": "..."}
func (a *API) createTodo(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	title := strings.TrimSpace(body.Title)
	if title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	todo := a.store.Create(title)
	a.pub.Publish("created", todo)
	writeJSON(w, http.StatusCreated, todo)
}

// PUT /api/todos/{id}  ->  update title + done (covers "modify" and "finish")
func (a *API) updateTodo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var body struct {
		Title string `json:"title"`
		Done  bool   `json:"done"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	title := strings.TrimSpace(body.Title)
	if title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	todo, err := a.store.Update(id, title, body.Done)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "todo not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	a.pub.Publish("updated", todo)
	writeJSON(w, http.StatusOK, todo)
}

// DELETE /api/todos/{id}  ->  remove a todo
func (a *API) deleteTodo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	todo, err := a.store.Delete(id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "todo not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	a.pub.Publish("deleted", todo)
	w.WriteHeader(http.StatusNoContent)
}
