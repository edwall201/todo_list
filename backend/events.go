package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

// CloudEvent is a minimal implementation of the CloudEvents v1.0 JSON
// format (https://cloudevents.io). Publishing this shape means any
// CloudEvents-aware tool can consume our messages, but we keep it
// hand-written so there are no extra libraries to learn.
type CloudEvent struct {
	SpecVersion     string          `json:"specversion"`
	ID              string          `json:"id"`
	Source          string          `json:"source"`
	Type            string          `json:"type"`
	Time            string          `json:"time"`
	DataContentType string          `json:"datacontenttype"`
	Data            json.RawMessage `json:"data"`
}

// publishEvent wraps a todo in a CloudEvent and broadcasts it on
// todo.event.<action> for any interested subscriber: the logger, other
// services, or live browser tabs. It's fire-and-forget — the command
// reply does not depend on it.
func publishEvent(nc *nats.Conn, action string, todo Todo) {
	data, err := json.Marshal(todo)
	if err != nil {
		log.Printf("event: marshal todo: %v", err)
		return
	}

	evt := CloudEvent{
		SpecVersion:     "1.0",
		ID:              randID(),
		Source:          "/todo-service",
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

	subject := "todo.event." + action
	if err := nc.Publish(subject, payload); err != nil {
		log.Printf("event: publish %s: %v", subject, err)
	}
}

// randID returns a short random hex string for the CloudEvent id.
func randID() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
