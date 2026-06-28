package main

import (
	"log"
	"net/http"
)

func main() {
	// 1. The "DB": an in-memory store.
	store := NewStore()

	// A couple of sample todos so the UI isn't empty on first run.
	store.Create("Welcome! Click the box to finish a task")
	store.Create("Double-click a task (or the pencil) to edit it")
	store.Create("Press the + to add your own")

	// 2. The event publisher. Connects to NATS and publishes a
	//    CloudEvent after every change. If NATS isn't running this is a
	//    no-op, so the REST API still works on its own.
	pub := NewEventPublisher()
	defer pub.Close()

	// 3. The TODO service: HTTP handlers that talk to the store and
	//    publish events.
	api := &API{store: store, pub: pub}

	// 4. Routes. Go 1.22+ lets the standard library router match by
	//    HTTP method and capture path values like {id}.
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/todos", api.listTodos)
	mux.HandleFunc("POST /api/todos", api.createTodo)
	mux.HandleFunc("PUT /api/todos/{id}", api.updateTodo)
	mux.HandleFunc("DELETE /api/todos/{id}", api.deleteTodo)
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	addr := ":8080"
	log.Printf("TODO service listening on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, withCORS(mux)); err != nil {
		log.Fatal(err)
	}
}

// withCORS wraps the router so the React dev server (a different origin,
// http://localhost:5173) is allowed to call this API from the browser.
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Browsers send a preflight OPTIONS request before PUT/DELETE.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
