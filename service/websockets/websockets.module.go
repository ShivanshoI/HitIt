package websockets

import (
	"net/http"
)

// InitModule wires all websockets dependencies and registers routes on mux.
// It returns the Hub instance so other modules can broadcast to it.
func InitModule(mux *http.ServeMux) *Hub {
	hub := NewHub()
	go hub.Run()

	handler := NewWebsocketsHandler(hub)
	handler.RegisterRoutes(mux)

	return hub
}
