// Package ws implements WebSocket handling for real-time feed updates.
//
// Authentication: the first message from the client MUST be a JSON object
// containing a JWT token:
//
//	{ "type": "auth", "token": "<jwt-or-api-key>" }
//
// The server responds with:
//
//	{ "type": "auth_ok" }
//	{ "type": "auth_error", "error": "..." }
package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
)

// MessageType describes the kind of WebSocket event.
type MessageType string

const (
	// Client-to-server types.
	MsgAuth   MessageType = "auth"
	MsgPing   MessageType = "ping"

	// Server-to-client types.
	MsgAuthOK      MessageType = "auth_ok"
	MsgAuthError   MessageType = "auth_error"
	MsgNewArticle  MessageType = "new_article"
	MsgFeedUpdate  MessageType = "feed_update"
	MsgPong        MessageType = "pong"
	MsgError       MessageType = "error"
)

// Message is the envelope for all WebSocket communication.
type Message struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
	Error   string          `json:"error,omitempty"`
}

// Authenticator defines the interface for verifying tokens during WebSocket upgrade.
type Authenticator interface {
	VerifyToken(tokenString string) (*AuthUser, error)
}

// AuthUser represents an authenticated WebSocket client.
type AuthUser struct {
	UserID  string
	IsAdmin bool
}

// Hub manages connected WebSocket clients and broadcasts events.
type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[*Client]struct{} // userID -> set of connections
}

// Client represents a single WebSocket connection.
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	userID string
	done   chan struct{}
}

// NewHub creates a new WebSocket Hub.
func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]map[*Client]struct{}),
	}
}

// HandleConnection upgrades an HTTP connection to WebSocket and runs the client loop.
func (h *Hub) HandleConnection(w http.ResponseWriter, r *http.Request, auth Authenticator) error {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // allow all origins; CORS handled by the middleware layer
	})
	if err != nil {
		return err
	}

	// Wait for auth message.
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	_, msgBytes, err := conn.Read(ctx)
	if err != nil {
		conn.Close(websocket.StatusPolicyViolation, "read timeout or error")
		return err
	}

	var msg Message
	if err := json.Unmarshal(msgBytes, &msg); err != nil || msg.Type != MsgAuth {
		writeMessage(conn, Message{Type: MsgAuthError, Error: "first message must be auth"})
		conn.Close(websocket.StatusPolicyViolation, "auth required")
		return nil
	}

	var payload struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(msg.Payload, &payload); err != nil || payload.Token == "" {
		writeMessage(conn, Message{Type: MsgAuthError, Error: "auth payload must include token"})
		conn.Close(websocket.StatusPolicyViolation, "invalid auth")
		return nil
	}

	user, err := auth.VerifyToken(payload.Token)
	if err != nil {
		writeMessage(conn, Message{Type: MsgAuthError, Error: "invalid token"})
		conn.Close(websocket.StatusPolicyViolation, "invalid token")
		return nil
	}

	client := &Client{
		hub:    h,
		conn:   conn,
		userID: user.UserID,
		done:   make(chan struct{}),
	}

	h.mu.Lock()
	if h.clients[user.UserID] == nil {
		h.clients[user.UserID] = make(map[*Client]struct{})
	}
	h.clients[user.UserID][client] = struct{}{}
	h.mu.Unlock()

	writeMessage(conn, Message{Type: MsgAuthOK})

	slog.Info("websocket client connected", "user_id", user.UserID)

	// Read loop — keep connection alive, handle pings.
	go func() {
		defer func() {
			h.mu.Lock()
			delete(h.clients[user.UserID], client)
			if len(h.clients[user.UserID]) == 0 {
				delete(h.clients, user.UserID)
			}
			h.mu.Unlock()
			conn.Close(websocket.StatusNormalClosure, "connection closed")
			close(client.done)
			slog.Info("websocket client disconnected", "user_id", user.UserID)
		}()

		for {
			_, _, err := conn.Read(context.Background())
			if err != nil {
				return
			}
			// For now, we just drain incoming messages. Pings are handled
			// automatically by the websocket library.
		}
	}()

	return nil
}

// BroadcastToUser sends a message to all connections for a given user.
func (h *Hub) BroadcastToUser(userID string, msg Message) {
	h.mu.RLock()
	clients := h.clients[userID]
	h.mu.RUnlock()

	for client := range clients {
		writeMessage(client.conn, msg)
	}
}

// BroadcastAll sends a message to every connected client.
func (h *Hub) BroadcastAll(msg Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, clients := range h.clients {
		for client := range clients {
			writeMessage(client.conn, msg)
		}
	}
}

func writeMessage(conn *websocket.Conn, msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		slog.Error("websocket marshal error", "error", err)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = conn.Write(ctx, websocket.MessageText, data)
}