package ddpserver

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Protocol version
	protocolVersion = "1"

	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Session is a middleman between the websocket connection and the server.
type Session struct {
	server *Server

	// The websocket connection.
	socket *websocket.Conn

	// Connected when handshake is successful
	connected bool

	// Session ID
	sessionID string

	// Buffered channel of outbound messages.
	send chan []byte

	// Client subscriptions
	subscriptions map[string]*SubscriptionContext

	// UserID of logged user
	userID string
}

func newSession(server *Server, socket *websocket.Conn) *Session {
	s := &Session{
		server:        server,
		socket:        socket,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]*SubscriptionContext),
	}

	s.run()

	return s
}

func (c *Session) run() {
	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go c.writePump()
	go c.readPump()

	// Send server ID message
	c.sendServerID()
}

// readPump pumps messages from the websocket connection to the server.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Session) readPump() {
	defer func() {
		c.server.unregister <- c
		c.socket.Close()
	}()
	c.socket.SetReadLimit(maxMessageSize)
	c.socket.SetReadDeadline(time.Now().Add(pongWait))
	for {
		_, p, err := c.socket.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("error: %v", err)
			}
			break
		}
		fmt.Printf("recv: %s", p)

		// Decode message
		var m Message
		if err := json.Unmarshal(p, &m); err != nil {
			fmt.Printf("error: %v", err)
			break
		}
		switch m.Msg {
		case "connect":
			c.handleConnect(m)
		case "ping":
			c.handlePing(m)
		case "pong":
			c.handlePong(m)
		case "method":
			c.handleMethod(m)
		case "sub":
			c.handleSubscribe(m)
		case "unsub":
			c.handleUnsubscribe(m)
		default:
			fmt.Println("error: Unknown message received from client")
		}
	}
	fmt.Println("Client disconnected")
}

// writePump pumps messages from the server to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Session) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.socket.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.socket.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The server closed the channel.
				c.socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.socket.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			fmt.Printf("sent: %s\n", message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.sendPing()
		}
	}
}

func (c *Session) writeMessage(msg interface{}) error {
	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	c.send <- b
	return nil
}

func (c *Session) handleConnect(msg Message) {
	if msg.Version != protocolVersion {
		c.sendFailed()
		c.socket.Close()
		return
	}
	c.sessionID = generateID(17)
	c.sendConnected()
}

func (c *Session) handlePing(msg Message) {
	c.sendPong(msg.ID)
}

func (c *Session) handlePong(msg Message) {
	c.socket.SetReadDeadline(time.Now().Add(pongWait))
}

func (c *Session) handleMethod(msg Message) {
	ctx := NewMethodContext(msg, *c)

	fn, ok := c.server.methods[msg.Method]
	if !ok {
		c.sendMethodError(&ctx, &Error{
			Type:    "Method.Error",
			Error:   "unknown-method",
			Reason:  fmt.Sprintf("Method '%s' not found", msg.Method),
			Message: fmt.Sprintf("Method '%s' not found [unknown]", msg.Method),
		})
		return
	}

	r, err := fn(ctx)
	if err != nil {
		c.sendMethodError(&ctx, err)
		return
	}
	c.sendMethodResult(&ctx, r)

	// TODO: Update publications
	c.sendMethodUpdated(&ctx)
}

func (c *Session) handleSubscribe(msg Message) {
	ctx := NewSubscriptionContext(msg, *c)

	id := generateID(17)

	c.subscriptions[id] = ctx

	handler, ok := c.server.publications[msg.Name]
	if !ok {
		c.sendSubscriptionError(ctx, &Error{
			Type:    "Server.Error",
			Error:   "unknown-subscription",
			Reason:  fmt.Sprintf("Subscription '%s' not found", msg.Name),
			Message: fmt.Sprintf("Subscription '%s' not found [unknown-subscription]", msg.Name),
		})
		return
	}

	handler(*ctx)
}

func (c *Session) handleUnsubscribe(msg Message) {
	_, ok := c.subscriptions[msg.ID]
	if !ok {
		msg := map[string]interface{}{
			"msg": "nosub",
			"id":  msg.ID,
			"error": Error{
				Type:    "Server.Error",
				Error:   "unknown-subscription",
				Reason:  fmt.Sprintf("Subscription ID '%s' not found", msg.ID),
				Message: fmt.Sprintf("Subscription ID '%s' not found [unknown-subscription]", msg.ID),
			},
		}
		c.writeMessage(msg)
		return
	}

	delete(c.subscriptions, msg.ID)

	// Update subscriptions
}

func (c *Session) sendServerID() {
	msg := map[string]string{
		"server_id": c.server.id,
	}
	c.writeMessage(msg)
}

func (c *Session) sendConnected() {
	msg := map[string]string{
		"msg":     "connected",
		"session": c.sessionID,
	}
	c.writeMessage(msg)
}

func (c *Session) sendFailed() {
	msg := map[string]string{
		"msg":     "failed",
		"version": protocolVersion,
	}
	c.writeMessage(msg)
}

func (c *Session) sendPing() {
	msg := map[string]string{
		"msg": "ping",
	}
	c.writeMessage(msg)
}

func (c *Session) sendPong(id string) {
	msg := map[string]string{
		"msg": "pong",
	}
	if id != "" {
		msg["id"] = id
	}
	c.writeMessage(msg)
}

func (c *Session) sendMethodResult(ctx *MethodContext, result interface{}) error {
	ctx.done = true
	msg := map[string]interface{}{
		"msg":    "result",
		"id":     ctx.ID,
		"result": result,
	}
	return c.writeMessage(msg)
}

func (c *Session) sendMethodError(ctx *MethodContext, e *Error) error {
	ctx.done = true
	msg := map[string]interface{}{
		"msg":   "result",
		"id":    ctx.ID,
		"error": *e,
	}
	return c.writeMessage(msg)
}

func (c *Session) sendMethodUpdated(ctx *MethodContext) error {
	ctx.updated = true
	msg := map[string]interface{}{
		"msg":     "updated",
		"methods": []string{ctx.ID},
	}
	return c.writeMessage(msg)
}

func (c *Session) sendSubscriptionError(ctx *SubscriptionContext, e *Error) error {
	msg := map[string]interface{}{
		"msg":   "nosub",
		"id":    ctx.ID,
		"error": *e,
	}
	return c.writeMessage(msg)
}

func (c *Session) sendSubscriptionReady(ctx *SubscriptionContext) error {
	ctx.ready = true

	msg := map[string]interface{}{
		"msg":  "ready",
		"subs": []string{ctx.ID},
	}
	return c.writeMessage(msg)
}

func (c *Session) sendAdded(ctx *SubscriptionContext, collection string, id string, fields map[string]interface{}) error {
	msg := map[string]interface{}{
		"msg":        "added",
		"collection": collection,
		"id":         id,
		"fields":     fields,
	}
	return c.writeMessage(msg)
}

func (c *Session) sendChanged(ctx *SubscriptionContext, collection string, id string, fields map[string]interface{}, cleared map[string]interface{}) error {
	msg := map[string]interface{}{
		"msg":        "changed",
		"collection": collection,
		"id":         id,
		"fields":     fields,
		"cleared":    cleared,
	}
	return c.writeMessage(msg)
}

func (c *Session) sendRemoved(ctx *SubscriptionContext, collection string, id string) error {
	msg := map[string]interface{}{
		"msg":        "removed",
		"collection": collection,
		"id":         id,
	}
	return c.writeMessage(msg)
}
