# API Documentation

## REST Endpoints

### Health Check
```
GET /health
```
Returns server status.

### List Agents
```
GET /api/agents
```
Returns all agents with their current status and metrics.

### Spawn Agent
```
POST /api/agents
Content-Type: application/json

{
  "name": "Code Reviewer",
  "role": "reviewer",
  "prompt": "Review the authentication module"
}
```

### Stop Agent
```
DELETE /api/agents/{id}
```

### Steer Agent
```
POST /api/agents/{id}/steer
Content-Type: application/json

{
  "instruction": "Focus on security aspects"
}
```

### System Stats
```
GET /api/stats
```
Returns aggregate statistics (total agents, running, cost, tokens).

## WebSocket Protocol

Connect to `/ws` for real-time updates.

### Server → Client Messages

| Type | Description |
|------|-------------|
| `agent_spawned` | New agent created |
| `agent_update` | Agent status/tokens changed |
| `agent_log` | New log entry from agent |
| `agent_stopped` | Agent was stopped |
| `system_status` | Aggregate stats update |
| `error` | Error notification |

### Client → Server Commands

| Action | Payload |
|--------|---------|
| `spawn` | `{name, role, prompt}` |
| `stop` | `{agent_id}` |
| `steer` | `{agent_id, instruction}` |
