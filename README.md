# Agent Swarm Dashboard

A real-time dashboard for monitoring and controlling AI agents. Spawn, steer, and track multiple agents from a single interface.

**Live Demo:** https://agent-swarm.vercel.app

## Features

- Spawn AI agents with custom name, role, and task prompt
- Real-time WebSocket streaming of agent logs and status
- Stop/abort running agents mid-task
- Steer agents with mid-task instructions
- Track token usage and estimated cost per agent
- Aggregate system stats dashboard
- Terminal-style log output per agent

## Architecture

- **Backend:** Go with Gorilla WebSocket + Mux
- **Frontend:** Single-page HTML/CSS/JS (no build step)
- **Real-time:** WebSocket for live agent updates
- **Agents:** Simulated execution with realistic log output

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /health | Health check |
| GET | /api/agents | List all agents |
| POST | /api/agents | Spawn new agent |
| DELETE | /api/agents/{id} | Stop agent |
| POST | /api/agents/{id}/steer | Steer agent |
| GET | /api/stats | System stats |
| GET | /ws | WebSocket endpoint |

## Agent Roles

- Researcher - Gathers and synthesizes information
- Coder - Writes and reviews code
- Reviewer - Evaluates quality and correctness
- Writer - Creates documentation and content
- Analyst - Processes data and generates insights
- Tester - Validates functionality and edge cases

## Local Development

```bash
go mod tidy
go run cmd/server/main.go
# Open http://localhost:8080
```

## Deploy

```bash
vercel --prod
```

## License

MIT
