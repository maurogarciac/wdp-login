package ws

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type ClientList map[*Client]bool

type Client struct {
	user   string
	conn   *websocket.Conn
	server *ChatServer
	egress chan Event
}

var (
	pongWait     = 10 * time.Second
	pingInterval = (pongWait * 9) / 10
)

func NewClient(logger *zap.SugaredLogger, username string, connection *websocket.Conn, chatServer *ChatServer) *Client {
	return &Client{
		user:   username,
		conn:   connection,
		server: chatServer,
		egress: make(chan Event),
	}
}

// goroutine that reads input messages from client
func (c *Client) readMessages() {
	defer func() {
		c.server.removeClient(c)
	}()

	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		c.server.lg.Error(err)
		return
	}
	c.conn.SetPongHandler(c.pongHandler)

	for {
		_, message, err := c.conn.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.server.lg.Errorf("error reading message: %v", err)
			}
			break
		}
		c.server.lg.Infof("Payload: %s", string(message))

		if err := c.server.broadcastMessage(message, c.user); err != nil {
			c.server.lg.Error(err)
			return
		}
	}
}

func (c *Client) pongHandler(pongMsg string) error {
	c.server.lg.Debug("pong")
	return c.conn.SetReadDeadline(time.Now().Add(pongWait))
}

// goroutine that listens for new messages to output to all connected Clients
func (c *Client) writeMessages() {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		c.server.removeClient(c)
	}()

	for {
		select {
		case message, ok := <-c.egress:
			if !ok {
				if err := c.conn.WriteMessage(websocket.CloseMessage, nil); err != nil {
					c.server.lg.Errorf("connection closed: ", err)
				}
				return
			}

			data, err := json.Marshal(message)
			if err != nil {
				c.server.lg.Error(err)
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				c.server.lg.Error(err)
			}
			log.Println("Sent message" + string(data))
			// todo add to db here
		case <-ticker.C:
			c.server.lg.Debug("ping")
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				c.server.lg.Errorf("writemsg: ", err)
				return
			}
		}

	}
}
