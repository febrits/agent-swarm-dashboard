package hub

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/febrits/agent-swarm-dashboard/internal/models"
	"github.com/gorilla/websocket"
)

type Client struct {
	Hub  *Hub
	Conn *websocket.Conn
	Send chan []byte
}

type Hub struct {
	mu         sync.RWMutex
	clients    map[*Client]bool
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
}

func New() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		Broadcast:  make(chan []byte, 256),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("Client connected. Total: %d", len(h.clients))
		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			h.mu.Unlock()
			log.Printf("Client disconnected. Total: %d", len(h.clients))
		case message := <-h.Broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					delete(h.clients, client)
					close(client.Send)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) BroadcastAgentUpdate(agent *models.Agent) {
	msg := models.WSMessage{Type: "agent_update", Agent: agent}
	data, _ := json.Marshal(msg)
	h.Broadcast <- data
}

func (h *Hub) BroadcastAgentSpawned(agent *models.Agent) {
	msg := models.WSMessage{Type: "agent_spawned", Agent: agent}
	data, _ := json.Marshal(msg)
	h.Broadcast <- data
}

func (h *Hub) BroadcastAgentStopped(agentID string) {
	msg := models.WSMessage{Type: "agent_stopped", AgentID: agentID}
	data, _ := json.Marshal(msg)
	h.Broadcast <- data
}

func (h *Hub) BroadcastAgentLog(agentID string, log models.Log) {
	msg := models.WSMessage{Type: "agent_log", AgentID: agentID, Log: &log}
	data, _ := json.Marshal(msg)
	h.Broadcast <- data
}

func (h *Hub) BroadcastSystemStatus(stats models.Stats) {
	msg := models.WSMessage{Type: "system_status", Stats: &stats}
	data, _ := json.Marshal(msg)
	h.Broadcast <- data
}

func (h *Hub) BroadcastError(message string) {
	payload, _ := json.Marshal(map[string]string{"message": message})
	msg := models.WSMessage{Type: "error", Payload: payload}
	data, _ := json.Marshal(msg)
	h.Broadcast <- data
}
