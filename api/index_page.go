package handler

const indexPage = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Agent Swarm Dashboard</title>
<style>
  *{margin:0;padding:0;box-sizing:border-box}
  :root{
    --bg:#0d0f17;--bg2:#131620;--bg3:#1a1e2e;--bg4:#22273a;
    --border:#2a2f42;--border2:#353b52;
    --text:#e2e6f0;--text2:#8b92a8;--text3:#5a6078;
    --accent:#6c5ce7;--accent2:#a29bfe;--accent-glow:rgba(108,92,231,.3);
    --green:#00cec9;--green2:#55efc4;
    --red:#ff6b6b;--orange:#fdcb6e;--blue:#74b9ff;--pink:#fd79a8;
    --radius:12px;--radius-sm:8px;--radius-xs:4px;
  }
  html,body{height:100%;overflow:hidden}
  body{font-family:'SF Mono','Fira Code','Cascadia Code',monospace;background:var(--bg);color:var(--text);font-size:13px;line-height:1.5}

  /* ── Layout ── */
  .app{display:flex;height:100vh}
  .sidebar{width:260px;min-width:260px;background:var(--bg2);border-right:1px solid var(--border);display:flex;flex-direction:column}
  .main{flex:1;display:flex;flex-direction:column;overflow:hidden}

  /* ── Sidebar ── */
  .sidebar-header{padding:20px;border-bottom:1px solid var(--border)}
  .sidebar-header h1{font-size:16px;font-weight:700;background:linear-gradient(135deg,var(--accent),var(--green));-webkit-background-clip:text;-webkit-text-fill-color:transparent;letter-spacing:-.5px}
  .sidebar-header .subtitle{font-size:11px;color:var(--text3);margin-top:2px}
  .sidebar-stats{padding:16px;border-bottom:1px solid var(--border)}
  .stat-row{display:flex;justify-content:space-between;align-items:center;margin-bottom:8px;font-size:12px}
  .stat-row:last-child{margin-bottom:0}
  .stat-label{color:var(--text2)}
  .stat-value{font-weight:600}
  .stat-value.running{color:var(--green)}
  .stat-value.cost{color:var(--orange)}
  .stat-value.tokens{color:var(--blue)}
  .sidebar-section{padding:16px;flex:1;overflow-y:auto}
  .sidebar-section h3{font-size:10px;text-transform:uppercase;letter-spacing:1.5px;color:var(--text3);margin-bottom:12px}
  .agent-list{display:flex;flex-direction:column;gap:6px}
  .agent-item{background:var(--bg3);border:1px solid var(--border);border-radius:var(--radius-sm);padding:10px 12px;cursor:pointer;transition:all .2s;position:relative;overflow:hidden}
  .agent-item:hover{border-color:var(--accent);background:var(--bg4)}
  .agent-item.active{border-color:var(--accent);box-shadow:0 0 0 1px var(--accent),0 0 20px var(--accent-glow)}
  .agent-item .agent-name{font-weight:600;font-size:12px;margin-bottom:2px;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}
  .agent-item .agent-role{font-size:10px;color:var(--text3)}
  .agent-item .agent-status-dot{position:absolute;top:10px;right:10px;width:8px;height:8px;border-radius:50%}
  .status-running{background:var(--green);box-shadow:0 0 8px var(--green);animation:pulse 2s infinite}
  .status-completed{background:var(--blue)}
  .status-error{background:var(--red)}
  .status-idle{background:var(--text3)}
  @keyframes pulse{0%,100%{opacity:1}50%{opacity:.4}}
  .sidebar-footer{padding:16px;border-top:1px solid var(--border)}
  .btn{width:100%;padding:10px;border:none;border-radius:var(--radius-sm);font-family:inherit;font-size:12px;font-weight:600;cursor:pointer;transition:all .2s}
  .btn-primary{background:linear-gradient(135deg,var(--accent),#8b5cf6);color:#fff}
  .btn-primary:hover{transform:translateY(-1px);box-shadow:0 4px 20px var(--accent-glow)}
  .btn-primary:active{transform:translateY(0)}
  .btn-danger{background:transparent;border:1px solid var(--red);color:var(--red);margin-top:8px}
  .btn-danger:hover{background:rgba(255,107,107,.1)}
  .btn-ghost{background:var(--bg3);border:1px solid var(--border);color:var(--text2)}
  .btn-ghost:hover{border-color:var(--text2);color:var(--text)}

  /* ── Main Content ── */
  .topbar{padding:12px 20px;border-bottom:1px solid var(--border);display:flex;align-items:center;justify-content:space-between;background:var(--bg2)}
  .topbar-left{display:flex;align-items:center;gap:12px}
  .topbar-title{font-size:14px;font-weight:600}
  .topbar-badge{font-size:10px;padding:2px 8px;border-radius:20px;background:var(--bg4);color:var(--text2);border:1px solid var(--border)}
  .topbar-badge.live{background:rgba(0,206,201,.1);color:var(--green);border-color:rgba(0,206,201,.3)}
  .topbar-actions{display:flex;gap:8px}
  .topbar-actions .btn{width:auto;padding:6px 14px;font-size:11px}

  .content{flex:1;display:flex;overflow:hidden}
  .panel-left{width:380px;min-width:380px;border-right:1px solid var(--border);display:flex;flex-direction:column;overflow:hidden}
  .panel-right{flex:1;display:flex;flex-direction:column;overflow:hidden}

  /* ── Agent Detail ── */
  .agent-detail-header{padding:16px 20px;border-bottom:1px solid var(--border)}
  .agent-detail-header h2{font-size:15px;font-weight:700;margin-bottom:4px}
  .agent-detail-header .meta{font-size:11px;color:var(--text3);display:flex;gap:16px}
  .agent-detail-header .meta span{display:flex;align-items:center;gap:4px}
  .agent-metrics{display:grid;grid-template-columns:repeat(3,1fr);gap:12px;padding:16px 20px;border-bottom:1px solid var(--border)}
  .metric-card{background:var(--bg3);border:1px solid var(--border);border-radius:var(--radius-sm);padding:12px;text-align:center}
  .metric-card .metric-value{font-size:18px;font-weight:700}
  .metric-card .metric-label{font-size:10px;color:var(--text3);text-transform:uppercase;letter-spacing:.5px;margin-top:2px}
  .metric-card:nth-child(1) .metric-value{color:var(--green)}
  .metric-card:nth-child(2) .metric-value{color:var(--orange)}
  .metric-card:nth-child(3) .metric-value{color:var(--blue)}

  /* ── Log Viewer ── */
  .log-viewer{flex:1;overflow-y:auto;padding:12px 16px;background:#0a0c14;font-size:12px}
  .log-entry{display:flex;gap:8px;padding:3px 0;border-bottom:1px solid rgba(255,255,255,.02)}
  .log-time{color:var(--text3);font-size:10px;white-space:nowrap;padding-top:1px}
  .log-level{font-size:9px;font-weight:700;padding:1px 5px;border-radius:3px;text-transform:uppercase;white-space:nowrap;height:fit-content}
  .log-level.info{background:rgba(116,185,255,.15);color:var(--blue)}
  .log-level.success{background:rgba(0,206,201,.15);color:var(--green)}
  .log-level.warn{background:rgba(253,203,110,.15);color:var(--orange)}
  .log-level.error{background:rgba(255,107,107,.15);color:var(--red)}
  .log-msg{flex:1;word-break:break-word}
  .log-empty{display:flex;align-items:center;justify-content:center;height:100%;color:var(--text3);font-size:13px}

  /* ── Right Panel: Spawn Form ── */
  .spawn-panel{padding:20px;overflow-y:auto}
  .spawn-panel h3{font-size:13px;font-weight:600;margin-bottom:16px;display:flex;align-items:center;gap:8px}
  .form-group{margin-bottom:14px}
  .form-group label{display:block;font-size:11px;color:var(--text2);margin-bottom:6px;text-transform:uppercase;letter-spacing:.5px}
  .form-group input,.form-group select,.form-group textarea{width:100%;padding:10px 12px;background:var(--bg3);border:1px solid var(--border);border-radius:var(--radius-sm);color:var(--text);font-family:inherit;font-size:13px;transition:border-color .2s;outline:none}
  .form-group input:focus,.form-group select:focus,.form-group textarea:focus{border-color:var(--accent)}
  .form-group textarea{resize:vertical;min-height:80px}
  .role-grid{display:grid;grid-template-columns:repeat(2,1fr);gap:6px}
  .role-option{padding:8px 10px;background:var(--bg3);border:1px solid var(--border);border-radius:var(--radius-xs);cursor:pointer;text-align:center;font-size:11px;transition:all .2s}
  .role-option:hover{border-color:var(--text2)}
  .role-option.selected{border-color:var(--accent);background:rgba(108,92,231,.1);color:var(--accent2)}
  .role-option .role-icon{font-size:16px;display:block;margin-bottom:2px}

  /* ── Empty State ── */
  .empty-state{display:flex;flex-direction:column;align-items:center;justify-content:center;height:100%;color:var(--text3);text-align:center;padding:40px}
  .empty-state .icon{font-size:48px;margin-bottom:16px;opacity:.5}
  .empty-state h3{font-size:16px;color:var(--text2);margin-bottom:8px}
  .empty-state p{font-size:12px;max-width:300px}

  /* ── Scrollbar ── */
  ::-webkit-scrollbar{width:6px;height:6px}
  ::-webkit-scrollbar-track{background:transparent}
  ::-webkit-scrollbar-thumb{background:var(--border2);border-radius:3px}
  ::-webkit-scrollbar-thumb:hover{background:var(--text3)}

  /* ── Animations ── */
  @keyframes fadeIn{from{opacity:0;transform:translateY(4px)}to{opacity:1;transform:translateY(0)}}
  .agent-item{animation:fadeIn .3s ease}
  .log-entry{animation:fadeIn .15s ease}

  /* ── Toast ── */
  .toast{position:fixed;bottom:20px;right:20px;padding:12px 20px;background:var(--bg3);border:1px solid var(--border);border-radius:var(--radius);font-size:12px;z-index:1000;animation:fadeIn .3s ease;box-shadow:0 8px 32px rgba(0,0,0,.4)}
  .toast.success{border-color:var(--green);color:var(--green)}
  .toast.error{border-color:var(--red);color:var(--red)}

  /* ── Responsive ── */
  @media(max-width:900px){
    .sidebar{width:200px;min-width:200px}
    .panel-left{width:280px;min-width:280px}
  }
</style>
</head>
<body>
<div class="app">
  <!-- Sidebar -->
  <div class="sidebar">
    <div class="sidebar-header">
      <h1>🐝 Agent Swarm</h1>
      <div class="subtitle">Real-time AI Agent Dashboard</div>
    </div>
    <div class="sidebar-stats">
      <div class="stat-row">
        <span class="stat-label">Active Agents</span>
        <span class="stat-value running" id="stat-running">0</span>
      </div>
      <div class="stat-row">
        <span class="stat-label">Total Agents</span>
        <span class="stat-value" id="stat-total">0</span>
      </div>
      <div class="stat-row">
        <span class="stat-label">Total Cost</span>
        <span class="stat-value cost" id="stat-cost">$0.0000</span>
      </div>
      <div class="stat-row">
        <span class="stat-label">Total Tokens</span>
        <span class="stat-value tokens" id="stat-tokens">0</span>
      </div>
    </div>
    <div class="sidebar-section">
      <h3>Agents</h3>
      <div class="agent-list" id="agent-list">
        <div class="empty-state" style="padding:20px">
          <div class="icon" style="font-size:24px">🤖</div>
          <p style="font-size:11px">No agents yet. Spawn one!</p>
        </div>
      </div>
    </div>
    <div class="sidebar-footer">
      <button class="btn btn-primary" onclick="showSpawnForm()">+ Spawn Agent</button>
      <button class="btn btn-ghost" onclick="refreshAll()" style="margin-top:8px">↻ Refresh</button>
    </div>
  </div>

  <!-- Main -->
  <div class="main">
    <div class="topbar">
      <div class="topbar-left">
        <span class="topbar-title" id="topbar-title">Dashboard</span>
        <span class="topbar-badge live" id="ws-status">● Connected</span>
      </div>
      <div class="topbar-actions">
        <button class="btn btn-ghost" onclick="showSpawnForm()">+ New Agent</button>
      </div>
    </div>
    <div class="content">
      <!-- Left: Agent Detail / Empty -->
      <div class="panel-left" id="panel-left">
        <div class="empty-state" id="empty-state">
          <div class="icon">🎯</div>
          <h3>Select an Agent</h3>
          <p>Choose an agent from the sidebar to view details, logs, and manage its lifecycle.</p>
        </div>
        <div id="agent-detail" style="display:none;flex-direction:column;height:100%">
          <div class="agent-detail-header">
            <h2 id="detail-name">Agent Name</h2>
            <div class="meta">
              <span id="detail-role">🔬 Researcher</span>
              <span id="detail-id">ID: xxx</span>
              <span id="detail-time">Started: xx:xx</span>
            </div>
          </div>
          <div class="agent-metrics">
            <div class="metric-card">
              <div class="metric-value" id="metric-status">Running</div>
              <div class="metric-label">Status</div>
            </div>
            <div class="metric-card">
              <div class="metric-value" id="metric-cost">$0.00</div>
              <div class="metric-label">Cost</div>
            </div>
            <div class="metric-card">
              <div class="metric-value" id="metric-tokens">0</div>
              <div class="metric-label">Tokens</div>
            </div>
          </div>
          <div style="padding:8px 16px;border-bottom:1px solid var(--border);display:flex;gap:8px;align-items:center">
            <span style="font-size:10px;color:var(--text3);text-transform:uppercase;letter-spacing:1px">Logs</span>
            <div style="flex:1"></div>
            <button class="btn btn-ghost" style="padding:4px 10px;font-size:10px" onclick="stopSelectedAgent()">■ Stop</button>
            <button class="btn btn-ghost" style="padding:4px 10px;font-size:10px" onclick="steerSelectedAgent()">↗ Steer</button>
          </div>
          <div class="log-viewer" id="log-viewer">
            <div class="log-empty">Waiting for agent logs...</div>
          </div>
        </div>
      </div>

      <!-- Right: Spawn Form -->
      <div class="panel-right">
        <div class="spawn-panel" id="spawn-panel">
          <h3>🚀 Spawn New Agent</h3>
          <div class="form-group">
            <label>Agent Name</label>
            <input type="text" id="spawn-name" placeholder="e.g. Code Reviewer, Data Analyst..." />
          </div>
          <div class="form-group">
            <label>Role</label>
            <div class="role-grid" id="role-grid">
              <div class="role-option selected" data-role="researcher" onclick="selectRole(this)">
                <span class="role-icon">🔬</span>Researcher
              </div>
              <div class="role-option" data-role="coder" onclick="selectRole(this)">
                <span class="role-icon">💻</span>Coder
              </div>
              <div class="role-option" data-role="reviewer" onclick="selectRole(this)">
                <span class="role-icon">🔍</span>Reviewer
              </div>
              <div class="role-option" data-role="writer" onclick="selectRole(this)">
                <span class="role-icon">✍️</span>Writer
              </div>
              <div class="role-option" data-role="analyst" onclick="selectRole(this)">
                <span class="role-icon">📊</span>Analyst
              </div>
              <div class="role-option" data-role="tester" onclick="selectRole(this)">
                <span class="role-icon">🧪</span>Tester
              </div>
            </div>
          </div>
          <div class="form-group">
            <label>Task Prompt</label>
            <textarea id="spawn-prompt" placeholder="Describe what this agent should do..."></textarea>
          </div>
          <button class="btn btn-primary" onclick="spawnAgent()" style="margin-top:4px">Launch Agent</button>
        </div>
      </div>
    </div>
  </div>
</div>

<script>
const WS_URL = location.origin.replace('http', 'ws') + '/ws';
const API_URL = '/api';

let agents = {};
let selectedAgentId = null;
let selectedRole = 'researcher';
let ws = null;
let reconnectTimer = null;

// ── WebSocket ──
function connectWS() {
  ws = new WebSocket(WS_URL);
  ws.onopen = () => {
    document.getElementById('ws-status').textContent = '● Connected';
    document.getElementById('ws-status').className = 'topbar-badge live';
    clearTimeout(reconnectTimer);
  };
  ws.onclose = () => {
    document.getElementById('ws-status').textContent = '● Disconnected';
    document.getElementById('ws-status').className = 'topbar-badge';
    reconnectTimer = setTimeout(connectWS, 3000);
  };
  ws.onerror = () => { ws.close() };
  ws.onmessage = (e) => {
    try {
      const msg = JSON.parse(e.data);
      handleWSMessage(msg);
    } catch(err) {}
  };
}

function handleWSMessage(msg) {
  switch(msg.type) {
    case 'agent_spawned':
      agents[msg.agent.id] = msg.agent;
      renderAgentList();
      toast('Agent spawned: ' + msg.agent.name, 'success');
      break;
    case 'agent_update':
      if(agents[msg.agent.id]) {
        agents[msg.agent.id] = msg.agent;
        renderAgentList();
        if(selectedAgentId === msg.agent.id) renderAgentDetail(msg.agent);
      }
      break;
    case 'agent_log':
      if(agents[msg.agent_id]) {
        if(!agents[msg.agent_id].logs) agents[msg.agent_id].logs = [];
        agents[msg.agent_id].logs.push(msg.log);
        if(selectedAgentId === msg.agent_id) appendLog(msg.log);
      }
      break;
    case 'agent_stopped':
      if(agents[msg.agent_id]) {
        agents[msg.agent_id].status = 'error';
        renderAgentList();
        if(selectedAgentId === msg.agent_id) renderAgentDetail(agents[msg.agent_id]);
      }
      break;
    case 'system_status':
      updateStats(msg.stats);
      break;
  }
}

// ── API Calls ──
async function api(path, opts = {}) {
  const res = await fetch(API_URL + path, {
    headers: {'Content-Type': 'application/json'},
    ...opts
  });
  return res.json();
}

async function spawnAgent() {
  const name = document.getElementById('spawn-name').value.trim();
  const prompt = document.getElementById('spawn-prompt').value.trim();
  if(!name || !prompt) { toast('Name and prompt required', 'error'); return; }
  const agent = await api('/agents', {
    method: 'POST',
    body: JSON.stringify({name, role: selectedRole, prompt})
  });
  document.getElementById('spawn-name').value = '';
  document.getElementById('spawn-prompt').value = '';
  selectAgent(agent.id);
}

async function stopSelectedAgent() {
  if(!selectedAgentId) return;
  await api('/agents/' + selectedAgentId, {method: 'DELETE'});
  toast('Agent stopped', 'success');
}

async function steerSelectedAgent() {
  if(!selectedAgentId) return;
  const instruction = prompt('Enter steering instruction:');
  if(!instruction) return;
  await api('/agents/' + selectedAgentId + '/steer', {
    method: 'POST',
    body: JSON.stringify({instruction})
  });
  toast('Steering sent', 'success');
}

async function refreshAll() {
  const [agentsRes, statsRes] = await Promise.all([
    api('/agents'),
    api('/stats')
  ]);
  agents = {};
  agentsRes.forEach(a => agents[a.id] = a);
  renderAgentList();
  updateStats(statsRes);
}

// ── UI ──
function selectRole(el) {
  document.querySelectorAll('.role-option').forEach(r => r.classList.remove('selected'));
  el.classList.add('selected');
  selectedRole = el.dataset.role;
}

function selectAgent(id) {
  selectedAgentId = id;
  const agent = agents[id];
  if(!agent) return;
  document.getElementById('empty-state').style.display = 'none';
  document.getElementById('agent-detail').style.display = 'flex';
  document.getElementById('topbar-title').textContent = agent.name;
  renderAgentDetail(agent);
  renderAgentList();
}

function renderAgentDetail(a) {
  document.getElementById('detail-name').textContent = a.name;
  document.getElementById('detail-role').textContent = roleIcon(a.role) + ' ' + a.role;
  document.getElementById('detail-id').textContent = 'ID: ' + a.id.slice(0,8);
  document.getElementById('detail-time').textContent = 'Started: ' + fmtTime(a.created_at);
  document.getElementById('metric-status').textContent = a.status;
  document.getElementById('metric-status').style.color = statusColor(a.status);
  document.getElementById('metric-cost').textContent = '$' + (a.cost||0).toFixed(4);
  document.getElementById('metric-tokens').textContent = (a.tokens||0).toLocaleString();
  // Render logs
  const viewer = document.getElementById('log-viewer');
  if(a.logs && a.logs.length > 0) {
    viewer.innerHTML = a.logs.map(log => logEntryHTML(log)).join('');
    viewer.scrollTop = viewer.scrollHeight;
  } else {
    viewer.innerHTML = '<div class="log-empty">Waiting for agent logs...</div>';
  }
}

function appendLog(log) {
  const viewer = document.getElementById('log-viewer');
  const empty = viewer.querySelector('.log-empty');
  if(empty) empty.remove();
  viewer.insertAdjacentHTML('beforeend', logEntryHTML(log));
  viewer.scrollTop = viewer.scrollHeight;
}

function logEntryHTML(log) {
  return ` + "`" + `<div class="log-entry">
    <span class="log-time">${fmtTime(log.time)}</span>
    <span class="log-level ${log.level}">${log.level}</span>
    <span class="log-msg">${escHtml(log.message)}</span>
  </div>` + "`" + `;
}

function renderAgentList() {
  const list = document.getElementById('agent-list');
  const sorted = Object.values(agents).sort((a,b) => new Date(b.created_at) - new Date(a.created_at));
  if(sorted.length === 0) {
    list.innerHTML = '<div class="empty-state" style="padding:20px"><div class="icon" style="font-size:24px">🤖</div><p style="font-size:11px">No agents yet. Spawn one!</p></div>';
    return;
  }
  list.innerHTML = sorted.map(a => ` + "`" + `
    <div class="agent-item ${selectedAgentId === a.id ? 'active' : ''}" onclick="selectAgent('${a.id}')">
      <div class="agent-status-dot status-${a.status}"></div>
      <div class="agent-name">${escHtml(a.name)}</div>
      <div class="agent-role">${roleIcon(a.role)} ${a.role} · ${a.status}</div>
    </div>
  ` + "`" + `).join('');
}

function updateStats(s) {
  document.getElementById('stat-running').textContent = s.running_agents || 0;
  document.getElementById('stat-total').textContent = s.total_agents || 0;
  document.getElementById('stat-cost').textContent = '$' + (s.total_cost||0).toFixed(4);
  document.getElementById('stat-tokens').textContent = (s.total_tokens||0).toLocaleString();
}

function showSpawnForm() {
  document.getElementById('spawn-name').focus();
}

// ── Helpers ──
function roleIcon(role) {
  const icons = {researcher:'🔬',coder:'💻',reviewer:'🔍',writer:'✍️',analyst:'📊',tester:'🧪'};
  return icons[role] || '🤖';
}

function statusColor(s) {
  return {running: 'var(--green)', completed: 'var(--blue)', error: 'var(--red)', idle: 'var(--text3)'}[s] || 'var(--text2)';
}

function fmtTime(t) {
  if(!t) return '--:--';
  const d = new Date(t);
  return d.toLocaleTimeString('en-US', {hour12:false, hour:'2-digit', minute:'2-digit', second:'2-digit'});
}

function escHtml(s) {
  const d = document.createElement('div');
  d.textContent = s;
  return d.innerHTML;
}

function toast(msg, type='') {
  const t = document.createElement('div');
  t.className = 'toast ' + type;
  t.textContent = msg;
  document.body.appendChild(t);
  setTimeout(() => t.remove(), 3000);
}

// ── Init ──
connectWS();
refreshAll();
</script>
</body>
</html>
`
