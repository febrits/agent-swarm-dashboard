package agents

import (
	"context"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/febrits/agent-swarm-dashboard/internal/hub"
	"github.com/febrits/agent-swarm-dashboard/internal/models"
	"github.com/google/uuid"
)

// simulatedPhrases cycles through realistic agent log output.
var simulatedPhrases = []string{
	"Initializing context and loading instructions...",
	"Analyzing task requirements and planning approach.",
	"Retrieving relevant information from knowledge base.",
	"Processing input data through reasoning chain.",
	"Evaluating multiple solution strategies.",
	"Formulating response based on analysis.",
	"Cross-referencing with available tool outputs.",
	"Refining output for clarity and accuracy.",
	"Validating results against task objectives.",
	"Preparing final response payload.",
	"Calling external API for supplementary data.",
	"Parsing structured data from tool response.",
	"Applying transformation logic to raw output.",
	"Generating intermediate reasoning steps.",
	"Optimizing token usage for efficiency.",
	"Checking context window limits.",
	"Compacting conversation history for continuity.",
	"Performing self-verification of intermediate results.",
	"Detecting potential errors in reasoning chain.",
	"Correcting course and retrying with adjusted parameters.",
}

// Manager handles the lifecycle of all agents.
type Manager struct {
	mu     sync.RWMutex
	agents map[string]*models.Agent
	hub    *hub.Hub

	// agentCancelFuncs stores cancel functions for running agent goroutines.
	agentCancelFuncs map[string]context.CancelFunc
}

// NewManager creates a new agent manager.
func NewManager(h *hub.Hub) *Manager {
	return &Manager{
		agents:           make(map[string]*models.Agent),
		hub:              h,
		agentCancelFuncs: make(map[string]context.CancelFunc),
	}
}

// SpawnAgent creates a new agent and begins its simulated execution.
func (m *Manager) SpawnAgent(name, role, prompt string) *models.Agent {
	m.mu.Lock()
	defer m.mu.Unlock()

	agent := &models.Agent{
		ID:        uuid.New().String(),
		Name:      name,
		Role:      role,
		Status:    models.AgentStatusRunning,
		Prompt:    prompt,
		Logs:      make([]models.Log, 0),
		Tokens:    0,
		Cost:      0,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	m.agents[agent.ID] = agent

	ctx, cancel := context.WithCancel(context.Background())
	m.agentCancelFuncs[agent.ID] = cancel

	m.hub.BroadcastAgentSpawned(agent)

	log.Printf("Spawned agent %s (%s) with role %s", agent.ID, name, role)

	go m.simulateExecution(ctx, agent.ID, prompt)

	return agent
}

// GetAgent returns an agent by ID.
func (m *Manager) GetAgent(id string) (*models.Agent, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	agent, ok := m.agents[id]
	return agent, ok
}

// ListAgents returns all agents.
func (m *Manager) ListAgents() []*models.Agent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*models.Agent, 0, len(m.agents))
	for _, agent := range m.agents {
		result = append(result, agent)
	}
	return result
}

// StopAgent stops a running agent by ID.
func (m *Manager) StopAgent(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	cancel, ok := m.agentCancelFuncs[id]
	if !ok {
		return false
	}

	cancel()
	delete(m.agentCancelFuncs, id)

	if agent, exists := m.agents[id]; exists {
		agent.Status = models.AgentStatusError
		agent.UpdatedAt = time.Now().UTC()
		agent.Logs = append(agent.Logs, models.Log{
			Time:    time.Now().UTC(),
			Level:   "warn",
			Message: "Agent stopped by user request.",
		})
		m.hub.BroadcastAgentUpdate(agent)
		m.hub.BroadcastAgentStopped(id)
		log.Printf("Stopped agent %s", id)
		return true
	}

	return false
}

// SteerAgent sends a mid-task instruction to a running agent.
func (m *Manager) SteerAgent(id string, instruction string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	agent, ok := m.agents[id]
	if !ok {
		return false
	}

	logEntry := models.Log{
		Time:    time.Now().UTC(),
		Level:   "info",
		Message: "Received steering instruction: " + instruction,
	}

	agent.Logs = append(agent.Logs, logEntry)
	agent.UpdatedAt = time.Now().UTC()

	m.hub.BroadcastAgentUpdate(agent)
	log.Printf("Steered agent %s with instruction: %s", id, instruction)

	return true
}

// GetStats returns aggregated statistics across all agents.
func (m *Manager) GetStats() models.Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := models.Stats{
		TotalAgents: len(m.agents),
	}

	for _, agent := range m.agents {
		stats.TotalTokens += agent.Tokens
		stats.TotalCost += agent.Cost
		if agent.Status == models.AgentStatusRunning {
			stats.RunningAgents++
		}
	}

	return stats
}

// addLog adds a log entry to an agent and broadcasts it.
func (m *Manager) addLog(agentID string, level, message string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	agent, ok := m.agents[agentID]
	if !ok {
		return
	}

	logEntry := models.Log{
		Time:    time.Now().UTC(),
		Level:   level,
		Message: message,
	}

	agent.Logs = append(agent.Logs, logEntry)
	agent.UpdatedAt = time.Now().UTC()

	// Simulate token usage per log entry.
	tokens := rand.Intn(150) + 50
	agent.Tokens += tokens

	// Approximate cost: $0.002 per 1K tokens (rough average).
	agent.Cost = float64(agent.Tokens) / 1000.0 * 0.002

	m.hub.BroadcastAgentLog(agentID, logEntry)
	m.hub.BroadcastAgentUpdate(agent)
}

// updateAgentStatus updates an agent's status and broadcasts.
func (m *Manager) updateAgentStatus(agentID string, status models.AgentStatus) {
	m.mu.Lock()
	defer m.mu.Unlock()

	agent, ok := m.agents[agentID]
	if !ok {
		return
	}

	agent.Status = status
	agent.UpdatedAt = time.Now().UTC()
	m.hub.BroadcastAgentUpdate(agent)
}

// simulateExecution runs a simulated agent lifecycle in a goroutine.
func (m *Manager) simulateExecution(ctx context.Context, agentID string, prompt string) {
	// Determine execution duration: 10-60 seconds.
	duration := time.Duration(rand.Intn(50)+10) * time.Second
	interval := time.Duration(rand.Intn(3)+2) * time.Second

	defer func() {
		m.mu.Lock()
		delete(m.agentCancelFuncs, agentID)
		m.mu.Unlock()
	}()

	// Initial log.
	m.addLog(agentID, "info", "Agent started. Prompt: "+truncate(prompt, 80))
	m.addLog(agentID, "info", "Estimated duration: "+duration.String())

	// Process the prompt words to create themed output.
	promptWords := extractKeywords(prompt)
	if len(promptWords) == 0 {
		promptWords = []string{"task", "analysis", "processing"}
	}

	deadline := time.After(duration)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	stepCount := 0

	for {
		select {
		case <-ctx.Done():
			return

		case <-deadline:
			// Agent completes successfully.
			m.addLog(agentID, "success", "Task completed successfully.")
			m.addLog(agentID, "info", "Finalizing and cleaning up resources.")
			m.updateAgentStatus(agentID, models.AgentStatusCompleted)

			// Broadcast final stats.
			if agent, ok := m.GetAgent(agentID); ok {
				m.hub.BroadcastSystemStatus(m.GetStats())
				log.Printf("Agent %s completed. Tokens: %d, Cost: $%.6f",
					agentID, agent.Tokens, agent.Cost)
			}
			return

		case <-ticker.C:
			stepCount++

			// Pick a phrase and optionally inject prompt keywords.
			phrase := simulatedPhrases[rand.Intn(len(simulatedPhrases))]

			// Every few steps, reference the prompt topic.
			if stepCount%4 == 0 && len(promptWords) > 0 {
				keyword := promptWords[rand.Intn(len(promptWords))]
				phrase = phrase + " Focus area: " + keyword + "."
			}

			// Occasionally add a warning.
			if rand.Float32() < 0.1 {
				m.addLog(agentID, "warn", "High token usage detected in current step. Compacting context.")
			}

			level := "info"
			if stepCount > 3 && rand.Float32() < 0.2 {
				level = "success"
				phrase = "Milestone " + strings.Repeat(".", stepCount%3+1) + " " + phrase
			}

			m.addLog(agentID, level, phrase)
		}
	}
}

// truncate cuts a string to maxLen and appends "..." if needed.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// extractKeywords pulls meaningful words from a prompt string.
func extractKeywords(prompt string) []string {
	ignored := map[string]bool{
		"the": true, "a": true, "an": true, "is": true, "are": true,
		"was": true, "were": true, "be": true, "been": true, "being": true,
		"have": true, "has": true, "had": true, "do": true, "does": true,
		"did": true, "will": true, "would": true, "could": true, "should": true,
		"may": true, "might": true, "shall": true, "can": true, "to": true,
		"of": true, "in": true, "for": true, "on": true, "with": true,
		"at": true, "by": true, "from": true, "as": true, "into": true,
		"through": true, "during": true, "before": true, "after": true,
		"and": true, "or": true, "but": true, "not": true, "no": true,
		"this": true, "that": true, "it": true, "its": true, "i": true,
		"you": true, "he": true, "she": true, "we": true, "they": true,
		"my": true, "your": true, "his": true, "her": true, "our": true,
		"their": true, "me": true, "him": true, "them": true, "us": true,
	}

	words := strings.Fields(strings.ToLower(prompt))
	var keywords []string
	for _, w := range words {
		w = strings.Trim(w, ".,!?;:\"'()[]{}")
		if len(w) > 3 && !ignored[w] {
			keywords = append(keywords, w)
		}
	}
	return keywords
}
