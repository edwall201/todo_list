package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
)

func main() {
	// 1. The "DB": an in-memory store (unchanged).
	store := NewStore()
	store.Create("Welcome! This list now talks over NATS, not REST")
	store.Create("Open a second browser tab — changes sync live")
	store.Create("Press the + to add your own")

	// 2. Connect to NATS. Now that REST is gone, NATS is the ONLY way in
	//    or out. We retry for a while so the service can start before the
	//    broker is ready (e.g. under docker compose).
	url := os.Getenv("NATS_URL")
	if url == "" {
		url = nats.DefaultURL // nats://127.0.0.1:4222
	}
	nc, err := connectNATS(url, "todo-service")
	if err != nil {
		log.Fatalf("cannot connect to NATS at %s: %v", url, err)
	}
	defer nc.Drain()
	log.Printf("todo-service connected to NATS at %s", url)

	svc := &Service{store: store, nc: nc}

	// 3. Commands (request-reply). The browser sends a request on these
	//    subjects and waits for svc to reply — this is what replaces the
	//    old REST endpoints.
	mustSub(nc, "todo.cmd.list", svc.handleList)
	mustSub(nc, "todo.cmd.create", svc.handleCreate)
	mustSub(nc, "todo.cmd.update", svc.handleUpdate)
	mustSub(nc, "todo.cmd.delete", svc.handleDelete)
	log.Printf("listening for commands on todo.cmd.*  /  broadcasting events on todo.event.*")

	// 4. Block until Ctrl+C.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("shutting down")
}

// connectNATS dials NATS, retrying for ~30s so the service can start
// before the broker is ready (handy under docker compose, where all
// containers start at roughly the same time).
func connectNATS(url, name string) (*nats.Conn, error) {
	var nc *nats.Conn
	var err error
	for attempt := 1; attempt <= 30; attempt++ {
		nc, err = nats.Connect(url, nats.Name(name), nats.Timeout(2*time.Second))
		if err == nil {
			return nc, nil
		}
		log.Printf("waiting for NATS at %s (attempt %d/30): %v", url, attempt, err)
		time.Sleep(time.Second)
	}
	return nil, err
}

// mustSub registers a subject handler and aborts if it fails.
func mustSub(nc *nats.Conn, subject string, handler nats.MsgHandler) {
	if _, err := nc.Subscribe(subject, handler); err != nil {
		log.Fatalf("subscribe %s: %v", subject, err)
	}
}
