package main

import (
	"encoding/json"
	"errors"
	"log"
	"strings"

	"github.com/nats-io/nats.go"
)

// Service handles incoming NATS *command* messages and replies to them.
// It replaces the old HTTP handlers: instead of (w http.ResponseWriter,
// r *http.Request) each handler now takes a *nats.Msg and answers with
// msg.Respond(...).
type Service struct {
	store *Store
	nc    *nats.Conn
}

// Reply is the envelope every command reply uses. NATS has no HTTP status
// codes, so we carry success/failure in the payload ourselves.
type Reply struct {
	OK    bool   `json:"ok"`
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

func (s *Service) respond(msg *nats.Msg, r Reply) {
	payload, err := json.Marshal(r)
	if err != nil {
		log.Printf("marshal reply: %v", err)
		return
	}
	if err := msg.Respond(payload); err != nil {
		log.Printf("respond on %s: %v", msg.Subject, err)
	}
}

func (s *Service) ok(msg *nats.Msg, data any)   { s.respond(msg, Reply{OK: true, Data: data}) }
func (s *Service) fail(msg *nats.Msg, e string) { s.respond(msg, Reply{OK: false, Error: e}) }

// todo.cmd.list  ->  reply with every todo
func (s *Service) handleList(msg *nats.Msg) {
	s.ok(msg, s.store.List())
}

// todo.cmd.create  {title}  ->  create, broadcast event, reply with new todo
func (s *Service) handleCreate(msg *nats.Msg) {
	var body struct {
		Title string `json:"title"`
	}
	if err := json.Unmarshal(msg.Data, &body); err != nil {
		s.fail(msg, "invalid JSON body")
		return
	}
	title := strings.TrimSpace(body.Title)
	if title == "" {
		s.fail(msg, "title is required")
		return
	}

	todo := s.store.Create(title)
	publishEvent(s.nc, "created", todo)
	s.ok(msg, todo)
}

// todo.cmd.update  {id,title,done}  ->  update, broadcast event, reply
func (s *Service) handleUpdate(msg *nats.Msg) {
	var body struct {
		ID    int64  `json:"id"`
		Title string `json:"title"`
		Done  bool   `json:"done"`
	}
	if err := json.Unmarshal(msg.Data, &body); err != nil {
		s.fail(msg, "invalid JSON body")
		return
	}
	title := strings.TrimSpace(body.Title)
	if title == "" {
		s.fail(msg, "title is required")
		return
	}

	todo, err := s.store.Update(body.ID, title, body.Done)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			s.fail(msg, "todo not found")
			return
		}
		s.fail(msg, err.Error())
		return
	}

	publishEvent(s.nc, "updated", todo)
	s.ok(msg, todo)
}

// todo.cmd.delete  {id}  ->  delete, broadcast event, reply
func (s *Service) handleDelete(msg *nats.Msg) {
	var body struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(msg.Data, &body); err != nil {
		s.fail(msg, "invalid JSON body")
		return
	}

	todo, err := s.store.Delete(body.ID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			s.fail(msg, "todo not found")
			return
		}
		s.fail(msg, err.Error())
		return
	}

	publishEvent(s.nc, "deleted", todo)
	s.ok(msg, todo)
}
