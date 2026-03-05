package activity_feed

import (
	"log"
	"net/http"
	"sync"
	"time"

	pogActivityFeedDB "pog/database/activity_feed"

	"github.com/gorilla/websocket"
)

const (
	// writeWait is the time allowed to write a message to a peer.
	writeWait = 10 * time.Second

	// sendBufSize is the number of messages buffered per client before the
	// client is considered dead and is forcibly evicted.
	sendBufSize = 256
)

var upgrader = websocket.Upgrader{
	// TODO: restrict allowed origins before deploying to production.
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Client represents a single WebSocket connection.
type Client struct {
	UserID   string
	MasterID string
	Conn     *websocket.Conn
	// Send is a buffered channel of outbound JSON payloads.
	Send chan []byte
}

// broadcastMessage is the internal envelope passed to Hub.broadcast.
type broadcastMessage struct {
	masterID string
	userID   string // non-empty only for PersonalScope messages
	scope    pogActivityFeedDB.MessageScope
	data     []byte
}

// Hub maintains the set of active clients and routes broadcast messages.
// All map mutations go through the select loop to avoid data races.
type Hub struct {
	// groups holds all clients keyed by MasterID.
	groups map[string]map[*Client]bool
	// personal holds the most-recent client for each UserID.
	personal map[string]*Client

	broadcast  chan broadcastMessage
	register   chan *Client
	unregister chan *Client

	// mu guards groups and personal when they are read outside the Run loop
	// (currently only for diagnostics; the Run loop itself does not need it).
	mu sync.RWMutex
}

// NewHub allocates and returns a ready-to-use Hub.
func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan broadcastMessage, 64), // buffered to reduce back-pressure
		register:   make(chan *Client),
		unregister: make(chan *Client),
		groups:     make(map[string]map[*Client]bool),
		personal:   make(map[string]*Client),
	}
}

// Run processes registrations, unregistrations and broadcasts.
// It must be called in its own goroutine (see InitModule).
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.addClient(client)

		case client := <-h.unregister:
			h.removeClient(client)

		case msg := <-h.broadcast:
			h.deliver(msg)
		}
	}
}

// ---------- Run helpers (called only from the Run goroutine) ----------

func (h *Hub) addClient(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.groups[c.MasterID] == nil {
		h.groups[c.MasterID] = make(map[*Client]bool)
	}
	h.groups[c.MasterID][c] = true
	h.personal[c.UserID] = c
	log.Printf("[hub] registered  user=%s master=%s", c.UserID, c.MasterID)
}

func (h *Hub) removeClient(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if group, ok := h.groups[c.MasterID]; ok {
		if _, ok := group[c]; ok {
			delete(group, c)
			close(c.Send)
			if len(group) == 0 {
				delete(h.groups, c.MasterID)
			}
		}
	}
	// Only delete from personal if this client is still the one registered
	// (a newer connection for the same user may have replaced it).
	if h.personal[c.UserID] == c {
		delete(h.personal, c.UserID)
	}
	log.Printf("[hub] unregistered user=%s master=%s", c.UserID, c.MasterID)
}

// deliver sends msg to the appropriate client set based on scope.
func (h *Hub) deliver(msg broadcastMessage) {
	h.mu.Lock()
	defer h.mu.Unlock()

	switch msg.scope {
	case pogActivityFeedDB.GroupScope:
		group := h.groups[msg.masterID]
		for c := range group {
			h.sendOrEvict(c, msg.data, group, msg.masterID)
		}

	case pogActivityFeedDB.PersonalScope:
		if c, ok := h.personal[msg.userID]; ok {
			h.sendOrEvict(c, msg.data, h.groups[c.MasterID], c.MasterID)
		}
	}
}

// sendOrEvict attempts a non-blocking send to c.Send.
// If the channel is full the client is considered unresponsive and is evicted.
func (h *Hub) sendOrEvict(c *Client, data []byte, group map[*Client]bool, masterID string) {
	select {
	case c.Send <- data:
	default:
		log.Printf("[hub] evicting unresponsive client user=%s", c.UserID)
		if group != nil {
			delete(group, c)
			if len(group) == 0 {
				delete(h.groups, masterID)
			}
		}
		if h.personal[c.UserID] == c {
			delete(h.personal, c.UserID)
		}
		close(c.Send)
	}
}

// ---------- Client ----------

// writePump pumps outbound messages from c.Send to the WebSocket connection.
// One goroutine per client; exits when c.Send is closed by the hub.
func (c *Client) writePump() {
	defer c.Conn.Close()

	for msg := range c.Send {
		c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
		if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Printf("[ws] write error user=%s: %v", c.UserID, err)
			return
		}
	}

	// c.Send was closed by the hub; send a clean close frame.
	c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
	c.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
}