// Command todo-subscriber is a standalone program that listens for the
// todo.* CloudEvents published by the TODO service and prints them.
//
// It shares NO code and NO memory with the service -- the only thing
// connecting them is the NATS subject. That decoupling is the whole
// point of event-driven messaging: you can add, remove, or restart
// subscribers without the publisher ever knowing.
//
// Run it with:  go run ./subscriber
package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/nats.go"
)

// event mirrors the CloudEvents fields we want to print. We ignore the
// rest of the envelope.
type event struct {
	Type string          `json:"type"`
	Time string          `json:"time"`
	Data json.RawMessage `json:"data"`
}

func main() {
	url := os.Getenv("NATS_URL")
	if url == "" {
		url = nats.DefaultURL
	}

	nc, err := nats.Connect(url, nats.Name("todo-subscriber"))
	if err != nil {
		log.Fatalf("could not connect to NATS at %s: %v", url, err)
	}
	defer nc.Drain()

	// "todo.>" is a wildcard that matches todo.created, todo.updated,
	// todo.deleted, and anything else under "todo.".
	_, err = nc.Subscribe("todo.>", func(m *nats.Msg) {
		var e event
		if err := json.Unmarshal(m.Data, &e); err != nil {
			log.Printf("[%s] %s", m.Subject, string(m.Data))
			return
		}
		log.Printf("event  subject=%-13s type=%-26s data=%s",
			m.Subject, e.Type, string(e.Data))
	})
	if err != nil {
		log.Fatalf("subscribe: %v", err)
	}

	log.Printf("todo-subscriber listening on %s for \"todo.>\"  (Ctrl+C to quit)", url)

	// Block until the user hits Ctrl+C.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("shutting down")
}
