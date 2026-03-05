package websockets

import (
	"log"
	"net/http"

	"pog/internal"
	"pog/middleware"
)

type WebsocketsHandler struct {
	hub *Hub
}

func NewWebsocketsHandler(hub *Hub) *WebsocketsHandler {
	return &WebsocketsHandler{hub: hub}
}

func (h *WebsocketsHandler) RegisterRoutes(mux *http.ServeMux) {
	auth := middleware.Auth
	mux.Handle("GET "+internal.APIPrefix+"/v1/ws/{masterId}", auth(http.HandlerFunc(h.HandleWebSocket)))
}

func (h *WebsocketsHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	masterID := r.PathValue("masterId")
	userID := internal.MustUserID(r.Context())

	log.Printf("[WS] Upgrade attempt: masterID=%s, userID=%s", masterID, userID)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WS] Upgrade failed: %v", err)
		return
	}

	log.Printf("[WS] Upgrade successful: masterID=%s, userID=%s", masterID, userID)


	client := &Client{
		UserID:   userID,
		MasterID: masterID,
		Conn:     conn,
		Send:     make(chan []byte, sendBufSize),
	}

	h.hub.Register <- client
	go client.WritePump()

	defer func() { h.hub.Unregister <- client }()

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}
