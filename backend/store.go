package main

import (
	"errors"
	"sort"
	"sync"
	"time"
)

// ErrNotFound is returned when a todo with the given id does not exist.
var ErrNotFound = errors.New("todo not found")

// Store is an in-memory, thread-safe collection of todos.
//
// This is the "DB" in our architecture: client -> TODO service -> store.
// It is deliberately small and hidden behind a handful of methods
// (List / Create / Update / Delete). When you later want real
// persistence, you can replace this file with a SQLite or Postgres
// implementation that has the SAME methods, and the HTTP handlers
// won't need to change at all.
type Store struct {
	mu     sync.Mutex
	todos  map[int64]Todo
	nextID int64
}

// NewStore creates an empty store.
func NewStore() *Store {
	return &Store{
		todos:  make(map[int64]Todo),
		nextID: 1,
	}
}

// List returns every todo, newest first.
func (s *Store) List() []Todo {
	s.mu.Lock()
	defer s.mu.Unlock()

	list := make([]Todo, 0, len(s.todos))
	for _, t := range s.todos {
		list = append(list, t)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].ID > list[j].ID })
	return list
}

// Create adds a new todo and returns it.
func (s *Store) Create(title string) Todo {
	s.mu.Lock()
	defer s.mu.Unlock()

	t := Todo{
		ID:        s.nextID,
		Title:     title,
		Done:      false,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	s.todos[t.ID] = t
	s.nextID++
	return t
}

// Update changes the title and done state of an existing todo.
func (s *Store) Update(id int64, title string, done bool) (Todo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, ok := s.todos[id]
	if !ok {
		return Todo{}, ErrNotFound
	}
	t.Title = title
	t.Done = done
	s.todos[id] = t
	return t, nil
}

// Delete removes a todo by id.
func (s *Store) Delete(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.todos[id]; !ok {
		return ErrNotFound
	}
	delete(s.todos, id)
	return nil
}
