package main

// Todo is a single task in the list.
//
// The json tags control how this struct is serialized to / from the
// JSON that travels over the REST API.
type Todo struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Done      bool   `json:"done"`
	CreatedAt string `json:"created_at"`
}
