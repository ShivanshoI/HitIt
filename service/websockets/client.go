package websockets

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait   = 10 * time.Second
	sendBufSize = 256
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Client represents a single WebSocket connection.
type Client struct {
	UserID   string
	MasterID string
	Conn     *websocket.Conn
	Send     chan []byte
}

func (c *Client) WritePump() {
	defer c.Conn.Close()

	for msg := range c.Send {
		c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
		if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Printf("[ws] write error user=%s: %v", c.UserID, err)
			return
		}
	}

	c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
	c.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
}
