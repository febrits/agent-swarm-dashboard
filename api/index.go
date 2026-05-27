package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/febrits/agent-swarm-dashboard/internal/agents"
	"github.com/febrits/agent-swarm-dashboard/internal/hub"
	"github.com/febrits/agent-swarm-dashboard/internal/models"
	"github.com/febrits/agent-swarm-dashboard/internal/ws"
	"github.com/gorilla/mux"
)

var (
	h     *hub.Hub
	mgr   *agents.Manager
	wsHdl *ws.Handler
)

func init() {
	h = hub.New()
	mgr = agents.NewManager(h)
	wsHdl = ws.NewHandler(h, mgr)
	go h.Run()
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			h.BroadcastSystemStatus(mgr.GetStats())
		}
	}()
}

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
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}).Methods("GET")
	router.HandleFunc("/api/agents", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, mgr.ListAgents())
	}).Methods("GET")
	router.HandleFunc("/api/agents", spawnHandler).Methods("POST")
	router.HandleFunc("/api/agents/{id}", stopHandler).Methods("DELETE")
	router.HandleFunc("/api/agents/{id}/steer", steerHandler).Methods("POST")
	router.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, mgr.GetStats())
	}).Methods("GET")
	router.HandleFunc("/ws", wsHdl.ServeHTTP)
	router.HandleFunc("/", indexHandler).Methods("GET")
	router.ServeHTTP(w, r)
}

func spawnHandler(w http.ResponseWriter, r *http.Request) {
	var req models.SpawnRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}
	if req.Name == "" || req.Role == "" || req.Prompt == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name, role, and prompt are required"})
		return
	}
	agent := mgr.SpawnAgent(req.Name, req.Role, req.Prompt)
	writeJSON(w, http.StatusCreated, agent)
}

func stopHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if mgr.StopAgent(id) {
		writeJSON(w, http.StatusOK, map[string]string{"message": "Agent stopped"})
	} else {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Agent not found"})
	}
}

func steerHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var req models.SteerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}
	if mgr.SteerAgent(id, req.Instruction) {
		writeJSON(w, http.StatusOK, map[string]string{"message": "Steering instruction sent"})
	} else {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Agent not found"})
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(indexPage))
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode JSON: %v", err)
	}
}
