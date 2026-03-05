package websockets

import (
	"log"
	"sync"
)

// BroadcastMessage is the internal envelope passed to Hub.Broadcast.
type BroadcastMessage struct {
	MasterID      string
	UserID        string // non-empty only for PersonalScope messages
	ExcludeUserID string // non-empty to skip broadcasting to a specific user (e.g. the sender)
	Scope         string // "group" or "personal"
	Data          []byte
}

// Hub maintains the set of active clients and routes broadcast messages.
type Hub struct {
	groups     map[string]map[*Client]bool
	personal   map[string]*Client
	Broadcast  chan BroadcastMessage
	Register   chan *Client
	Unregister chan *Client
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		Broadcast:  make(chan BroadcastMessage, 64),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		groups:     make(map[string]map[*Client]bool),
		personal:   make(map[string]*Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.addClient(client)
		case client := <-h.Unregister:
			h.removeClient(client)
		case msg := <-h.Broadcast:
			h.deliver(msg)
		}
	}
}

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
	if h.personal[c.UserID] == c {
		delete(h.personal, c.UserID)
	}
	log.Printf("[hub] unregistered user=%s master=%s", c.UserID, c.MasterID)
}

func (h *Hub) deliver(msg BroadcastMessage) {
	h.mu.Lock()
	defer h.mu.Unlock()

	switch msg.Scope {
	case "group":
		group := h.groups[msg.MasterID]
		for c := range group {
			// Skip the sender if ExcludeUserID is set
			if msg.ExcludeUserID != "" && c.UserID == msg.ExcludeUserID {
				continue
			}
			h.sendOrEvict(c, msg.Data, group, msg.MasterID)
		}

	case "personal":
		if c, ok := h.personal[msg.UserID]; ok {
			h.sendOrEvict(c, msg.Data, h.groups[c.MasterID], c.MasterID)
		}
	}
}

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
