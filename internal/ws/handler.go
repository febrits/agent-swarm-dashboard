package ws

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/febrits/agent-swarm-dashboard/internal/agents"
	"github.com/febrits/agent-swarm-dashboard/internal/hub"
	"github.com/febrits/agent-swarm-dashboard/internal/models"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for development. In production, restrict this.
		return true
	},
}

// Handler manages WebSocket connections and commands.
type Handler struct {
	hub     *hub.Hub
	manager *agents.Manager
}

// NewHandler creates a new WebSocket handler.
func NewHandler(h *hub.Hub, m *agents.Manager) *Handler {
	return &Handler{
		hub:     h,
		manager: m,
	}
}

// ServeHTTP upgrades the HTTP connection to WebSocket and starts pumps.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	client := &hub.Client{
		Hub:  h.hub,
		Conn: conn,
		Send: make(chan []byte, 256),
	}

	h.hub.Register <- client

	// Send initial system status.
	stats := h.manager.GetStats()
	h.hub.BroadcastSystemStatus(stats)

	go client.WritePump()
	go h.handleClient(client)
}

// handleClient processes incoming messages from a WebSocket client.
func (h *Handler) handleClient(client *hub.Client) {
	defer func() {
		h.hub.Unregister <- client
		client.Conn.Close()
	}()

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseNormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		var cmd models.WSCommand
		if err := json.Unmarshal(message, &cmd); err != nil {
			h.hub.BroadcastError("Invalid command format: " + err.Error())
			continue
		}

		h.processCommand(client, cmd)
	}
}

// processCommand routes a WebSocket command to the appropriate handler.
func (h *Handler) processCommand(client *hub.Client, cmd models.WSCommand) {
	switch cmd.Action {
	case "spawn":
		var req models.WSSpawnRequest
		if err := json.Unmarshal(cmd.Payload, &req); err != nil {
			h.hub.BroadcastError("Invalid spawn payload: " + err.Error())
			return
		}
		if req.Name == "" || req.Role == "" || req.Prompt == "" {
			h.hub.BroadcastError("Spawn requires name, role, and prompt")
			return
		}
		h.manager.SpawnAgent(req.Name, req.Role, req.Prompt)

	case "stop":
		var req struct {
			AgentID string `json:"agent_id"`
		}
		if err := json.Unmarshal(cmd.Payload, &req); err != nil {
			h.hub.BroadcastError("Invalid stop payload: " + err.Error())
			return
		}
		if req.AgentID == "" {
			h.hub.BroadcastError("Stop requires agent_id")
			return
		}
		if !h.manager.StopAgent(req.AgentID) {
			h.hub.BroadcastError("Agent not found: " + req.AgentID)
		}

	case "steer":
		var req models.WSSteerRequest
		if err := json.Unmarshal(cmd.Payload, &req); err != nil {
			h.hub.BroadcastError("Invalid steer payload: " + err.Error())
			return
		}
		if req.AgentID == "" || req.Instruction == "" {
			h.hub.BroadcastError("Steer requires agent_id and instruction")
			return
		}
		if !h.manager.SteerAgent(req.AgentID, req.Instruction) {
			h.hub.BroadcastError("Agent not found: " + req.AgentID)
		}

	default:
		h.hub.BroadcastError("Unknown command: " + cmd.Action)
	}
}
