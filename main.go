package handler

import (
	"encoding/json"
	"log"
	"net/http"
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
			stats := mgr.GetStats()
			h.BroadcastSystemStatus(stats)
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

	router.HandleFunc("/api/agents", func(w http.ResponseWriter, r *http.Request) {
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
	}).Methods("POST")

	router.HandleFunc("/api/agents/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		if mgr.StopAgent(id) {
			writeJSON(w, http.StatusOK, map[string]string{"message": "Agent stopped"})
		} else {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Agent not found"})
		}
	}).Methods("DELETE")

	router.HandleFunc("/api/agents/{id}/steer", func(w http.ResponseWriter, r *http.Request) {
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
	}).Methods("POST")

	router.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, mgr.GetStats())
	}).Methods("GET")

	router.HandleFunc("/ws", wsHdl.ServeHTTP)

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(indexHTML))
	}).Methods("GET")

	router.ServeHTTP(w, r)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode JSON: %v", err)
	}
}


const indexHTML = "<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n<meta charset=\"utf-8\">\n<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n<title>Agent Swarm Dashboard</title>\n<style>\n:root{--bg:#0a0e17;--surface:#111827;--border:#1e293b;--text:#e2e8f0;--muted:#64748b;--accent:#7c3aed;--accent-light:#a78bfa;--green:#10b981;--red:#ef4444;--yellow:#f59e0b;--blue:#3b82f6;--cyan:#06b6d4}\n*{margin:0;padding:0;box-sizing:border-box}\nbody{background:var(--bg);color:var(--text);font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;min-height:100vh}\n.header{background:linear-gradient(135deg,#0f0720,#1a1040);border-bottom:1px solid var(--border);padding:20px 32px;display:flex;align-items:center;justify-content:space-between}\n.header h1{font-size:22px;font-weight:700}\n.header .status{display:flex;align-items:center;gap:8px;font-size:13px;color:var(--muted)}\n.status-dot{width:8px;height:8px;border-radius:50%;background:var(--red);transition:background .3s}\n.status-dot.connected{background:var(--green)}\n.container{max-width:1400px;margin:0 auto;padding:24px}\n.stats{display:grid;grid-template-columns:repeat(auto-fit,minmax(200px,1fr));gap:16px;margin-bottom:24px}\n.stat-card{background:var(--surface);border:1px solid var(--border);border-radius:12px;padding:20px}\n.stat-card .value{font-size:28px;font-weight:700;color:var(--accent-light)}\n.stat-card .label{font-size:13px;color:var(--muted);margin-top:4px}\n.main-grid{display:grid;grid-template-columns:1fr 400px;gap:24px}\n@media(max-width:1000px){.main-grid{grid-template-columns:1fr}}\n.panel{background:var(--surface);border:1px solid var(--border);border-radius:12px;overflow:hidden}\n.panel-header{padding:16px 20px;border-bottom:1px solid var(--border);display:flex;align-items:center;justify-content:space-between}\n.panel-header h2{font-size:15px;font-weight:600}\n.panel-body{padding:16px;max-height:600px;overflow-y:auto}\n.agent-card{background:#0d1117;border:1px solid var(--border);border-radius:10px;padding:16px;margin-bottom:12px;transition:border-color .2s}\n.agent-card:hover{border-color:var(--accent)}\n.agent-card.running{border-left:3px solid var(--green)}\n.agent-card.completed{border-left:3px solid var(--blue)}\n.agent-card.error{border-left:3px solid var(--red)}\n.agent-card.idle{border-left:3px solid var(--muted)}\n.agent-header{display:flex;align-items:center;justify-content:space-between;margin-bottom:8px}\n.agent-name{font-weight:600;font-size:14px}\n.agent-role{font-size:11px;color:var(--muted);background:var(--surface);padding:2px 8px;border-radius:4px}\n.agent-status{font-size:11px;padding:2px 8px;border-radius:4px;font-weight:600}\n.agent-status.running{background:rgba(16,185,129,.15);color:var(--green)}\n.agent-status.completed{background:rgba(59,130,246,.15);color:var(--blue)}\n.agent-status.error{background:rgba(239,68,68,.15);color:var(--red)}\n.agent-status.idle{background:rgba(100,116,139,.15);color:var(--muted)}\n.agent-meta{display:flex;gap:16px;font-size:12px;color:var(--muted);margin-top:8px}\n.agent-logs{margin-top:12px;background:#080c14;border-radius:8px;padding:12px;max-height:150px;overflow-y:auto;font-family:'JetBrains Mono',monospace;font-size:11px;line-height:1.6}\n.log-entry{display:flex;gap:8px}\n.log-time{color:var(--muted);flex-shrink:0}\n.log-level{font-weight:600;flex-shrink:0;width:45px}\n.log-level.info{color:var(--blue)}\n.log-level.success{color:var(--green)}\n.log-level.warn{color:var(--yellow)}\n.log-level.error{color:var(--red)}\n.spawn-form{display:flex;flex-direction:column;gap:12px}\n.form-group{display:flex;flex-direction:column;gap:6px}\n.form-group label{font-size:13px;color:var(--muted);font-weight:500}\n.form-group input,.form-group select,.form-group textarea{\n  padding:10px 14px;background:#0d1117;border:1px solid var(--border);border-radius:8px;\n  color:var(--text);font-size:14px;outline:none;font-family:inherit\n}\n.form-group input:focus,.form-group select:focus,.form-group textarea:focus{border-color:var(--accent)}\n.form-group textarea{resize:vertical;min-height:80px;font-family:'JetBrains Mono',monospace;font-size:13px}\n.btn{padding:10px 20px;border:none;border-radius:8px;font-size:14px;font-weight:600;cursor:pointer;transition:all .2s}\n.btn-primary{background:linear-gradient(135deg,var(--accent),#9333ea);color:#fff}\n.btn-primary:hover{transform:translateY(-1px);box-shadow:0 4px 20px rgba(124,58,237,.3)}\n.btn-danger{background:rgba(239,68,68,.15);color:var(--red);border:1px solid rgba(239,68,68,.3)}\n.btn-danger:hover{background:rgba(239,68,68,.25)}\n.btn-sm{padding:6px 12px;font-size:12px}\n.agent-actions{display:flex;gap:8px;margin-top:12px}\n.empty-state{text-align:center;padding:40px;color:var(--muted)}\n.empty-state .icon{font-size:40px;margin-bottom:12px}\n.ws-log{background:#080c14;border-radius:8px;padding:12px;max-height:200px;overflow-y:auto;font-family:'JetBrains Mono',monospace;font-size:11px;line-height:1.6;margin-top:12px}\n.ws-entry{display:flex;gap:8px;padding:2px 0}\n.ws-time{color:var(--muted);flex-shrink:0;font-size:10px}\n.ws-type{color:var(--accent-light);flex-shrink:0;font-weight:600}\n</style>\n</head>\n<body>\n<div class=\"header\">\n  <div style=\"display:flex;align-items:center;gap:12px\">\n    <div style=\"width:36px;height:36px;background:linear-gradient(135deg,var(--accent),#9333ea);border-radius:10px;display:flex;align-items:center;justify-content:center;font-size:18px\">🐝</div>\n    <h1>Agent Swarm Dashboard</h1>\n  </div>\n  <div class=\"status\">\n    <div class=\"status-dot\" id=\"statusDot\"></div>\n    <span id=\"statusText\">Disconnected</span>\n  </div>\n</div>\n\n<div class=\"container\">\n  <div class=\"stats\">\n    <div class=\"stat-card\"><div class=\"value\" id=\"totalAgents\">0</div><div class=\"label\">Total Agents</div></div>\n    <div class=\"stat-card\"><div class=\"value\" id=\"runningAgents\">0</div><div class=\"label\">Running</div></div>\n    <div class=\"stat-card\"><div class=\"value\" id=\"completedAgents\">0</div><div class=\"label\">Completed</div></div>\n    <div class=\"stat-card\"><div class=\"value\" id=\"totalCost\">$0.00</div><div class=\"label\">Total Cost</div></div>\n    <div class=\"stat-card\"><div class=\"value\" id=\"totalTokens\">0</div><div class=\"label\">Total Tokens</div></div>\n  </div>\n\n  <div class=\"main-grid\">\n    <div>\n      <div class=\"panel\">\n        <div class=\"panel-header\">\n          <h2>Active Agents</h2>\n          <span style=\"font-size:12px;color:var(--muted)\" id=\"agentCount\">0 agents</span>\n        </div>\n        <div class=\"panel-body\" id=\"agentList\">\n          <div class=\"empty-state\">\n            <div class=\"icon\">🤖</div>\n            <p>No agents running</p>\n            <p style=\"font-size:12px;margin-top:4px\">Spawn an agent to get started</p>\n          </div>\n        </div>\n      </div>\n    </div>\n\n    <div>\n      <div class=\"panel\" style=\"margin-bottom:24px\">\n        <div class=\"panel-header\"><h2>Spawn Agent</h2></div>\n        <div class=\"panel-body\">\n          <div class=\"spawn-form\">\n            <div class=\"form-group\">\n              <label>Agent Name</label>\n              <input type=\"text\" id=\"agentName\" placeholder=\"e.g. Code Reviewer\">\n            </div>\n            <div class=\"form-group\">\n              <label>Role</label>\n              <select id=\"agentRole\">\n                <option value=\"researcher\">Researcher</option>\n                <option value=\"coder\">Coder</option>\n                <option value=\"reviewer\">Reviewer</option>\n                <option value=\"writer\">Writer</option>\n                <option value=\"analyst\">Analyst</option>\n                <option value=\"tester\">Tester</option>\n              </select>\n            </div>\n            <div class=\"form-group\">\n              <label>Task Prompt</label>\n              <textarea id=\"agentPrompt\" placeholder=\"Describe what this agent should do...\"></textarea>\n            </div>\n            <button class=\"btn btn-primary\" onclick=\"spawnAgent()\">Spawn Agent</button>\n          </div>\n        </div>\n      </div>\n\n      <div class=\"panel\">\n        <div class=\"panel-header\"><h2>System Log</h2></div>\n        <div class=\"panel-body\">\n          <div class=\"ws-log\" id=\"wsLog\">\n            <div class=\"ws-entry\"><span class=\"ws-time\">--:--:--</span><span class=\"ws-type\">INFO</span>Waiting for connection...</div>\n          </div>\n        </div>\n      </div>\n    </div>\n  </div>\n</div>\n\n<script>\nconst WS_URL = window.location.origin.replace('http','ws')+'/ws';\nlet ws=null;\nlet agents={};\nlet reconnectTimer=null;\n\nfunction connect(){\n  ws=new WebSocket(WS_URL);\n  ws.onopen=()=>{\n    document.getElementById('statusDot').classList.add('connected');\n    document.getElementById('statusText').textContent='Connected';\n    addLog('INFO','Connected to swarm hub');\n    fetchAgents();\n  };\n  ws.onclose=()=>{\n    document.getElementById('statusDot').classList.remove('connected');\n    document.getElementById('statusText').textContent='Disconnected - Reconnecting...';\n    addLog('WARN','Connection lost, reconnecting...');\n    reconnectTimer=setTimeout(connect,3000);\n  };\n  ws.onerror=(e)=>{\n    addLog('ERROR','WebSocket error');\n  };\n  ws.onmessage=(e)=>{\n    try{\n      const msg=JSON.parse(e.data);\n      handleMessage(msg);\n    }catch(err){\n      addLog('ERROR','Failed to parse message');\n    }\n  };\n}\n\nfunction handleMessage(msg){\n  switch(msg.type){\n    case 'agent_update':\n    case 'agent_spawned':\n      agents[msg.agent.id]=msg.agent;\n      renderAgents();\n      addLog('INFO',`Agent ${msg.agent.name} ${msg.type==='agent_spawned'?'spawned':'updated'} (${msg.agent.status})`);\n      break;\n    case 'agent_stopped':\n      if(agents[msg.agent.id]){\n        agents[msg.agent.id].status='idle';\n        renderAgents();\n        addLog('WARN',`Agent ${msg.agent.name} stopped`);\n      }\n      break;\n    case 'agent_log':\n      if(agents[msg.agent_id]){\n        agents[msg.agent_id].logs=agents[msg.agent_id].logs||[];\n        agents[msg.agent_id].logs.push(msg.log);\n        renderAgentLogs(msg.agent_id);\n      }\n      break;\n    case 'system_status':\n      updateStats(msg.stats);\n      break;\n    case 'error':\n      addLog('ERROR',msg.message);\n      break;\n  }\n}\n\nfunction fetchAgents(){\n  fetch('/api/agents').then(r=>r.json()).then(data=>{\n    data.forEach(a=>agents[a.id]=a);\n    renderAgents();\n  }).catch(()=>{});\n}\n\nfunction spawnAgent(){\n  const name=document.getElementById('agentName').value.trim();\n  const role=document.getElementById('agentRole').value;\n  const prompt=document.getElementById('agentPrompt').value.trim();\n  if(!name){alert('Please enter an agent name');return}\n  if(!prompt){alert('Please enter a task prompt');return}\n  fetch('/api/agents',{\n    method:'POST',\n    headers:{'Content-Type':'application/json'},\n    body:JSON.stringify({name,role,prompt})\n  }).then(r=>r.json()).then(agent=>{\n    document.getElementById('agentName').value='';\n    document.getElementById('agentPrompt').value='';\n  }).catch(()=>{});\n}\n\nfunction stopAgent(id){\n  fetch('/api/agents/'+id,{method:'DELETE'}).catch(()=>{});\n}\n\nfunction steerAgent(id){\n  const instruction=prompt('Enter steering instruction:');\n  if(!instruction)return;\n  fetch('/api/agents/'+id+'/steer',{\n    method:'POST',\n    headers:{'Content-Type':'application/json'},\n    body:JSON.stringify({instruction})\n  }).catch(()=>{});\n}\n\nfunction renderAgents(){\n  const list=document.getElementById('agentList');\n  const vals=Object.values(agents);\n  document.getElementById('agentCount').textContent=vals.length+' agent'+(vals.length!==1?'s':'');\n  if(vals.length===0){\n    list.innerHTML='<div class=\"empty-state\"><div class=\"icon\">🤖</div><p>No agents running</p><p style=\"font-size:12px;margin-top:4px\">Spawn an agent to get started</p></div>';\n    return;\n  }\n  vals.sort((a,b)=>new Date(b.created_at)-new Date(a.created_at));\n  list.innerHTML=vals.map(a=>`\n    <div class=\"agent-card ${a.status}\" id=\"agent-${a.id}\">\n      <div class=\"agent-header\">\n        <div>\n          <div class=\"agent-name\">${escapeHtml(a.name)}</div>\n          <div style=\"display:flex;gap:8px;margin-top:4px\">\n            <span class=\"agent-role\">${escapeHtml(a.role)}</span>\n            <span class=\"agent-status ${a.status}\">${a.status}</span>\n          </div>\n        </div>\n        <div class=\"agent-actions\">\n          ${a.status==='running'?`<button class=\"btn btn-sm btn-primary\" onclick=\"steerAgent('${a.id}')\">Steer</button>\n          <button class=\"btn btn-sm btn-danger\" onclick=\"stopAgent('${a.id}')\">Stop</button>`:''}\n        </div>\n      </div>\n      <div class=\"agent-meta\">\n        <span>Tokens: ${(a.tokens||0).toLocaleString()}</span>\n        <span>Cost: $${(a.cost||0).toFixed(4)}</span>\n        <span>Created: ${new Date(a.created_at).toLocaleTimeString()}</span>\n      </div>\n      <div class=\"agent-logs\" id=\"logs-${a.id}\">\n        ${(a.logs||[]).map(l=>`<div class=\"log-entry\"><span class=\"log-time\">${new Date(l.time).toLocaleTimeString()}</span><span class=\"log-level ${l.level}\">${l.level}</span><span>${escapeHtml(l.message)}</span></div>`).join('')||'<div style=\"color:var(--muted);font-size:11px\">No logs yet</div>'}\n      </div>\n    </div>\n  `).join('');\n  updateStatsFromAgents();\n}\n\nfunction renderAgentLogs(id){\n  const el=document.getElementById('logs-'+id);\n  if(!el||!agents[id])return;\n  const logs=agents[id].logs||[];\n  el.innerHTML=logs.map(l=>`<div class=\"log-entry\"><span class=\"log-time\">${new Date(l.time).toLocaleTimeString()}</span><span class=\"log-level ${l.level}\">${l.level}</span><span>${escapeHtml(l.message)}</span></div>`).join('')||'<div style=\"color:var(--muted);font-size:11px\">No logs yet</div>';\n  el.scrollTop=el.scrollHeight;\n}\n\nfunction updateStatsFromAgents(){\n  const vals=Object.values(agents);\n  document.getElementById('totalAgents').textContent=vals.length;\n  document.getElementById('runningAgents').textContent=vals.filter(a=>a.status==='running').length;\n  document.getElementById('completedAgents').textContent=vals.filter(a=>a.status==='completed').length;\n  document.getElementById('totalCost').textContent='$'+vals.reduce((s,a)=>s+(a.cost||0),0).toFixed(4);\n  document.getElementById('totalTokens').textContent=vals.reduce((s,a)=>s+(a.tokens||0),0).toLocaleString();\n}\n\nfunction updateStats(stats){\n  if(stats.total_agents!==undefined)document.getElementById('totalAgents').textContent=stats.total_agents;\n  if(stats.running!==undefined)document.getElementById('runningAgents').textContent=stats.running;\n  if(stats.completed!==undefined)document.getElementById('completedAgents').textContent=stats.completed;\n  if(stats.total_cost!==undefined)document.getElementById('totalCost').textContent='$'+stats.total_cost.toFixed(4);\n  if(stats.total_tokens!==undefined)document.getElementById('totalTokens').textContent=stats.total_tokens.toLocaleString();\n}\n\nfunction addLog(type,msg){\n  const el=document.getElementById('wsLog');\n  const time=new Date().toLocaleTimeString();\n  const div=document.createElement('div');\n  div.className='ws-entry';\n  div.innerHTML=`<span class=\"ws-time\">${time}</span><span class=\"ws-type\">${type}</span>${escapeHtml(msg)}`;\n  el.appendChild(div);\n  el.scrollTop=el.scrollHeight;\n  if(el.children.length>100)el.removeChild(el.firstChild);\n}\n\nfunction escapeHtml(s){\n  const d=document.createElement('div');\n  d.textContent=s;\n  return d.innerHTML;\n}\n\nconnect();\n</script>\n</body>\n</html>\n"
