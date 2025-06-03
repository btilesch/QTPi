package ws

import (
	"bytes"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

// For use in struct fields, method receivers, etc.
type Client interface {
	Id() string
	Conn() *websocket.Conn
	Send() chan []byte
	Receive() chan []byte
}

// For generic use only
type ComparableClient interface {
	Client
	comparable
}

type BaseClient struct {
	id           string
	conn         *websocket.Conn
	send         chan []byte
	receive      chan []byte
	OnUnregister func()
}

func NewBaseClient(id string, conn *websocket.Conn) *BaseClient {
	return &BaseClient{
		id:      id,
		conn:    conn,
		send:    make(chan []byte, 256),
		receive: make(chan []byte, 256),
	}
}

func (c *BaseClient) triggerUnregister() {
	if c.OnUnregister != nil {
		c.OnUnregister()
	}
}

// Interface methods
func (c *BaseClient) Id() string {
	return c.id
}

func (c *BaseClient) Conn() *websocket.Conn {
	return c.conn
}

func (c *BaseClient) Send() chan []byte {
	return c.send
}

func (c *BaseClient) Receive() chan []byte {
	return c.receive
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

func (c *BaseClient) ReadPump() {
	defer func() {
		c.triggerUnregister()
		c.conn.Close()
		close(c.Receive())
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))

		// Send to domain handler
		select {
		case c.Receive() <- message:
		default:
			log.Printf("Receive channel full for client %s", c.id)
		}
	}
}

func (c *BaseClient) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for range n {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
