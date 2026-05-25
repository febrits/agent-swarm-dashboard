package models

import (
	"encoding/json"
	"time"
)

// AgentStatus represents the current state of an agent.
type AgentStatus string

const (
	AgentStatusIdle      AgentStatus = "idle"
	AgentStatusRunning   AgentStatus = "running"
	AgentStatusCompleted AgentStatus = "completed"
	AgentStatusError     AgentStatus = "error"
)

// Agent represents a single AI agent instance.
type Agent struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Role      string     `json:"role"`
	Status    AgentStatus `json:"status"`
	Prompt    string     `json:"prompt"`
	Logs      []Log      `json:"logs"`
	Tokens    int        `json:"tokens"`
	Cost      float64    `json:"cost"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// Log represents a single log entry from an agent.
type Log struct {
	Time    time.Time `json:"time"`
	Level   string    `json:"level"`
	Message string    `json:"message"`
}

// Stats holds aggregate statistics across all agents.
type Stats struct {
	TotalAgents   int     `json:"total_agents"`
	RunningAgents int     `json:"running_agents"`
	TotalCost     float64 `json:"total_cost"`
	TotalTokens   int     `json:"total_tokens"`
}

// WSType represents the type of WebSocket message.
type WSType string

const (
	WSTypeAgentUpdate   WSType = "agent_update"
	WSTypeAgentLog      WSType = "agent_log"
	WSTypeAgentSpawned  WSType = "agent_spawned"
	WSTypeAgentStopped  WSType = "agent_stopped"
	WSTypeSystemStatus  WSType = "system_status"
	WSTypeError         WSType = "error"
)

// WSMessage is the envelope for all WebSocket messages.
type WSMessage struct {
	Type    WSType      `json:"type"`
	Payload interface{} `json:"payload"`
}

// WSSpawnRequest is the payload for spawning a new agent via WebSocket.
type WSSpawnRequest struct {
	Name   string `json:"name"`
	Role   string `json:"role"`
	Prompt string `json:"prompt"`
}

// WSSteerRequest is the payload for steering an agent via WebSocket.
type WSSteerRequest struct {
	AgentID    string `json:"agent_id"`
	Instruction string `json:"instruction"`
}

// WSCommand is a generic command from a WebSocket client.
type WSCommand struct {
	Action  string          `json:"action"`
	Payload json.RawMessage `json:"payload"`
}

// SpawnRequest is the HTTP request body for spawning an agent.
type SpawnRequest struct {
	Name   string `json:"name"`
	Role   string `json:"role"`
	Prompt string `json:"prompt"`
}

// SteerRequest is the HTTP request body for steering an agent.
type SteerRequest struct {
	Instruction string `json:"instruction"`
}
