package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/febrits/agent-swarm-dashboard/internal/agents"
	"github.com/febrits/agent-swarm-dashboard/internal/hub"
	"github.com/febrits/agent-swarm-dashboard/internal/models"
	"github.com/febrits/agent-swarm-dashboard/internal/ws"
	"github.com/gorilla/mux"
)

func main() {
	// Initialize hub and agent manager.
	h := hub.New()
	m := agents.NewManager(h)

	// Start the hub in a goroutine.
	go h.Run()

	// Start periodic stats broadcaster.
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			stats := m.GetStats()
			h.BroadcastSystemStatus(stats)
		}
	}()

	// Set up HTTP router.
	r := mux.NewRouter()

	// CORS middleware.
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	// Health check.
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}).Methods("GET")

	// API routes.
	api := r.PathPrefix("/api").Subrouter()

	// List all agents.
	api.HandleFunc("/agents", func(w http.ResponseWriter, r *http.Request) {
		agents := m.ListAgents()
		writeJSON(w, http.StatusOK, agents)
	}).Methods("GET")

	// Spawn a new agent.
	api.HandleFunc("/agents", func(w http.ResponseWriter, r *http.Request) {
		var req models.SpawnRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
			return
		}
		if req.Name == "" || req.Role == "" || req.Prompt == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name, role, and prompt are required"})
			return
		}
		agent := m.SpawnAgent(req.Name, req.Role, req.Prompt)
		writeJSON(w, http.StatusCreated, agent)
	}).Methods("POST")

	// Stop an agent.
	api.HandleFunc("/agents/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		if m.StopAgent(id) {
			writeJSON(w, http.StatusOK, map[string]string{"message": "Agent stopped"})
		} else {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Agent not found"})
		}
	}).Methods("DELETE")

	// Steer an agent.
	api.HandleFunc("/agents/{id}/steer", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		var req models.SteerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
			return
		}
		if m.SteerAgent(id, req.Instruction) {
			writeJSON(w, http.StatusOK, map[string]string{"message": "Steering instruction sent"})
		} else {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Agent not found"})
		}
	}).Methods("POST")

	// System stats.
	api.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		stats := m.GetStats()
		writeJSON(w, http.StatusOK, stats)
	}).Methods("GET")

	// WebSocket endpoint.
	wsHandler := ws.NewHandler(h, m)
	r.HandleFunc("/ws", wsHandler.ServeHTTP)

	// Serve frontend.
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/index.html")
	}).Methods("GET")

	// Create server.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine.
	go func() {
		log.Printf("Agent Swarm Dashboard starting on port %s", port)
		log.Printf("WebSocket endpoint: ws://localhost:%s/ws", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server stopped gracefully")
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
	}
}
