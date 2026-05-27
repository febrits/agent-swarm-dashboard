package handler

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// ─── Models ───────────────────────────────────────────────────────────────────

type AgentStatus string

const (
	StatusIdle      AgentStatus = "idle"
	StatusRunning   AgentStatus = "running"
	StatusCompleted AgentStatus = "completed"
	StatusError     AgentStatus = "error"
)

type Agent struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Role      string     `json:"role"`
	Status    AgentStatus `json:"status"`
	Prompt    string     `json:"prompt"`
	Logs      []LogEntry `json:"logs"`
	Tokens    int        `json:"tokens"`
	Cost      float64    `json:"cost"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type LogEntry struct {
	Time    time.Time `json:"time"`
	Level   string    `json:"level"`
	Message string    `json:"message"`
}

type SpawnRequest struct {
	Name   string `json:"name"`
	Role   string `json:"role"`
	Prompt string `json:"prompt"`
}

type SteerRequest struct {
	Instruction string `json:"instruction"`
}

type SysStats struct {
	TotalAgents   int     `json:"total_agents"`
	RunningAgents int     `json:"running_agents"`
	TotalCost     float64 `json:"total_cost"`
	TotalTokens   int     `json:"total_tokens"`
}

// ─── WebSocket Hub ───────────────────────────────────────────────────────────

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type Client struct {
	Send chan []byte
}

type Hub struct {
	mu      sync.RWMutex
	clients map[*Client]bool
	bcast   chan []byte
	reg     chan *Client
	unreg   chan *Client
}

func newHub() *Hub {
	return &Hub{
		clients: make(map[*Client]bool),
		bcast:   make(chan []byte, 256),
		reg:     make(chan *Client),
		unreg:   make(chan *Client),
	}
}

func (h *Hub) run() {
	for {
		select {
		case c := <-h.reg:
			h.mu.Lock()
			h.clients[c] = true
			h.mu.Unlock()
		case c := <-h.unreg:
			h.mu.Lock()
			if h.clients[c] {
				delete(h.clients, c)
				close(c.Send)
			}
			h.mu.Unlock()
		case msg := <-h.bcast:
			h.mu.RLock()
			for c := range h.clients {
				select {
				case c.Send <- msg:
				default:
					delete(h.clients, c)
					close(c.Send)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) broadcast(v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		return
	}
	h.bcast <- data
}

// ─── Agent Manager ───────────────────────────────────────────────────────────

var phrases = []string{
	"Initializing context and loading instructions",
	"Analyzing task requirements and planning approach",
	"Retrieving relevant information from knowledge base",
	"Processing input data through reasoning chain",
	"Evaluating multiple solution strategies",
	"Cross-referencing with available tool outputs",
	"Refining output for clarity and accuracy",
	"Validating results against task objectives",
	"Calling external API for supplementary data",
	"Parsing structured data from tool response",
	"Generating intermediate reasoning steps",
	"Checking context window limits",
	"Performing self-verification of intermediate results",
	"Compacting conversation history for continuity",
	"Applying transformation logic to raw output",
}

type Manager struct {
	mu      sync.RWMutex
	agents  map[string]*Agent
	hub     *Hub
	cancels map[string]context.CancelFunc
}

func newManager(h *Hub) *Manager {
	return &Manager{
		agents:  make(map[string]*Agent),
		hub:     h,
		cancels: make(map[string]context.CancelFunc),
	}
}

func (m *Manager) spawn(name, role, prompt string) *Agent {
	m.mu.Lock()
	defer m.mu.Unlock()

	a := &Agent{
		ID:        uuid.New().String(),
		Name:      name,
		Role:      role,
		Status:    StatusRunning,
		Prompt:    prompt,
		Logs:      []LogEntry{},
		Tokens:    rand.Intn(200) + 100,
		Cost:      0,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	a.Cost = float64(a.Tokens) / 1000.0 * 0.002
	m.agents[a.ID] = a

	ctx, cancel := context.WithCancel(context.Background())
	m.cancels[a.ID] = cancel

	m.hub.broadcast(map[string]interface{}{"type": "agent_spawned", "agent": a})

	go m.simulate(ctx, a.ID, prompt)
	log.Printf("Spawned agent %s (%s)", a.ID[:8], name)
	return a
}

func (m *Manager) stop(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	cancel, ok := m.cancels[id]
	if !ok {
		return false
	}
	cancel()
	delete(m.cancels, id)

	if a, exists := m.agents[id]; exists {
		a.Status = StatusError
		a.UpdatedAt = time.Now().UTC()
		a.Logs = append(a.Logs, LogEntry{
			Time: time.Now().UTC(), Level: "warn", Message: "Agent stopped by user",
		})
		m.hub.broadcast(map[string]interface{}{"type": "agent_stopped", "agent_id": id})
		return true
	}
	return false
}

func (m *Manager) steer(id, instruction string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	a, ok := m.agents[id]
	if !ok {
		return false
	}
	a.Logs = append(a.Logs, LogEntry{
		Time: time.Now().UTC(), Level: "info", Message: "Steering: " + instruction,
	})
	a.UpdatedAt = time.Now().UTC()
	m.hub.broadcast(map[string]interface{}{"type": "agent_update", "agent": a})
	return true
}

func (m *Manager) list() []*Agent {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*Agent, 0, len(m.agents))
	for _, a := range m.agents {
		result = append(result, a)
	}
	return result
}

func (m *Manager) getStats() SysStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s := SysStats{TotalAgents: len(m.agents)}
	for _, a := range m.agents {
		s.TotalTokens += a.Tokens
		s.TotalCost += a.Cost
		if a.Status == StatusRunning {
			s.RunningAgents++
		}
	}
	return s
}

func (m *Manager) addLog(id, level, msg string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	a, ok := m.agents[id]
	if !ok {
		return
	}

	tokens := rand.Intn(150) + 50
	a.Logs = append(a.Logs, LogEntry{Time: time.Now().UTC(), Level: level, Message: msg})
	a.Tokens += tokens
	a.Cost = float64(a.Tokens) / 1000.0 * 0.002
	a.UpdatedAt = time.Now().UTC()

	m.hub.broadcast(map[string]interface{}{
		"type": "agent_log", "agent_id": id,
		"log": a.Logs[len(a.Logs)-1],
	})
	m.hub.broadcast(map[string]interface{}{"type": "agent_update", "agent": a})
}

func (m *Manager) setStatus(id string, s AgentStatus) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if a, ok := m.agents[id]; ok {
		a.Status = s
		a.UpdatedAt = time.Now().UTC()
		m.hub.broadcast(map[string]interface{}{"type": "agent_update", "agent": a})
	}
}

func (m *Manager) simulate(ctx context.Context, id, prompt string) {
	dur := time.Duration(rand.Intn(50)+15) * time.Second
	interval := time.Duration(rand.Intn(3)+2) * time.Second

	defer func() {
		m.mu.Lock()
		delete(m.cancels, id)
		m.mu.Unlock()
	}()

	m.addLog(id, "info", "Agent started. Duration: "+dur.String())

	deadline := time.After(dur)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	step := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-deadline:
			m.addLog(id, "success", "Task completed successfully")
			m.setStatus(id, StatusCompleted)
			m.hub.broadcast(map[string]interface{}{
				"type": "system_status", "stats": m.getStats(),
			})
			return
		case <-ticker.C:
			step++
			msg := phrases[rand.Intn(len(phrases))]
			if step%4 == 0 {
				words := strings.Fields(prompt)
				if len(words) > 0 {
					msg += ". Focus: " + words[rand.Intn(len(words))]
				}
			}
			if rand.Float32() < 0.1 {
				m.addLog(id, "warn", "High detected in current step. Compacting context")
			}
			m.addLog(id, "info", msg)
		}
	}
}

// ─── Globals ─────────────────────────────────────────────────────────────────

var (
	hubMgr *Hub
	agentMgr *Manager
)

func init() {
	hubMgr = newHub()
	agentMgr = newManager(hubMgr)
	go hubMgr.run()

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			hubMgr.broadcast(map[string]interface{}{
				"type": "system_status", "stats": agentMgr.getStats(),
			})
		}
	}()
}

// ─── HTTP Handler ────────────────────────────────────────────────────────────

func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	router := mux.NewRouter()

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]string{"status": "ok"})
	}).Methods("GET")

	router.HandleFunc("/api/agents", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, agentMgr.list())
	}).Methods("GET")

	router.HandleFunc("/api/agents", func(w http.ResponseWriter, r *http.Request) {
		var req SpawnRequest
		if json.NewDecoder(r.Body).Decode(&req) != nil || req.Name == "" || req.Role == "" || req.Prompt == "" {
			writeJSON(w, 400, map[string]string{"error": "name, role, prompt required"})
			return
		}
		writeJSON(w, 201, agentMgr.spawn(req.Name, req.Role, req.Prompt))
	}).Methods("POST")

	router.HandleFunc("/api/agents/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		if agentMgr.stop(id) {
			writeJSON(w, 200, map[string]string{"message": "stopped"})
		} else {
			writeJSON(w, 404, map[string]string{"error": "not found"})
		}
	}).Methods("DELETE")

	router.HandleFunc("/api/agents/{id}/steer", func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		var req SteerRequest
		if json.NewDecoder(r.Body).Decode(&req) != nil || req.Instruction == "" {
			writeJSON(w, 400, map[string]string{"error": "instruction required"})
			return
		}
		if agentMgr.steer(id, req.Instruction) {
			writeJSON(w, 200, map[string]string{"message": "steered"})
		} else {
			writeJSON(w, 404, map[string]string{"error": "not found"})
		}
	}).Methods("POST")

	router.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, agentMgr.getStats())
	}).Methods("GET")

	router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		client := &Client{Send: make(chan []byte, 256)}
		hubMgr.reg <- client
		defer func() { hubMgr.unreg <- client }()

		go func() {
			for msg := range client.Send {
				if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
					return
				}
			}
		}()

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				return
			}
		}
	})

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(indexPage))
	}).Methods("GET")

	router.ServeHTTP(w, r)
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}
