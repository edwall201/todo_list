package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/nats-io/nats.go"
)

// CloudEvent is a minimal implementation of the CloudEvents v1.0 JSON
// format (https://cloudevents.io). Publishing this shape means any
// CloudEvents-aware tool can consume our messages later, but we keep it
// hand-written so there are no extra libraries to learn right now.
type CloudEvent struct {
	SpecVersion     string          `json:"specversion"`
	ID              string          `json:"id"`
	Source          string          `json:"source"`
	Type            string          `json:"type"`
	Time            string          `json:"time"`
	DataContentType string          `json:"datacontenttype"`
	Data            json.RawMessage `json:"data"`
}

// EventPublisher publishes todo.* events to NATS. If NATS is not
// reachable it quietly degrades to a no-op, so the REST API keeps
// working with or without a message broker running.
type EventPublisher struct {
	nc     *nats.Conn
	source string
}

// NewEventPublisher connects to NATS (NATS_URL env var, default
// nats://127.0.0.1:4222). A failed connection is logged but NOT fatal.
func NewEventPublisher() *EventPublisher {
	url := os.Getenv("NATS_URL")
	if url == "" {
		url = nats.DefaultURL
	}

	nc, err := nats.Connect(url,
		nats.Name("todo-service"),
		nats.Timeout(2*time.Second),
	)
	if err != nil {
		log.Printf("NATS not connected at %s (%v) -- events disabled, REST still works", url, err)
		return &EventPublisher{}
	}

	log.Printf("connected to NATS at %s -- publishing todo.* events", url)
	return &EventPublisher{nc: nc, source: "/todo-service"}
}

// Publish sends one CloudEvent describing what happened to a todo.
// action is one of "created", "updated", "deleted". The NATS subject
// becomes "todo.<action>", so a subscriber can match them all with
// the wildcard "todo.>".
func (p *EventPublisher) Publish(action string, todo Todo) {
	if p.nc == nil {
		return // NATS not connected; nothing to publish
	}

	data, err := json.Marshal(todo)
	if err != nil {
		log.Printf("event: marshal todo: %v", err)
		return
	}

	evt := CloudEvent{
		SpecVersion:     "1.0",
		ID:              randID(),
		Source:          p.source,
		Type:            "com.example.todo." + action,
		Time:            time.Now().UTC().Format(time.RFC3339),
		DataContentType: "application/json",
		Data:            data,
	}

	payload, err := json.Marshal(evt)
	if err != nil {
		log.Printf("event: marshal envelope: %v", err)
		return
	}

	subject := "todo." + action
	if err := p.nc.Publish(subject, payload); err != nil {
		log.Printf("event: publish %s: %v", subject, err)
	}
}

// Close flushes any buffered messages, then closes the connection.
func (p *EventPublisher) Close() {
	if p.nc != nil {
		_ = p.nc.Drain()
	}
}

// randID returns a short random hex string for the CloudEvent id.
func randID() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
