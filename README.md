# Agent Swarm Dashboard

A real-time dashboard for monitoring and controlling AI agents. Spawn, steer, and track multiple agents from a single interface.

**Live Demo:** https://agent-swarm-dashboard-eight.vercel.app

![Dashboard Preview](https://img.shields.io/badge/Theme-Dark-6c5ce7) ![Real-time](https://img.shields.io/badge/Real--time-WebSocket-00cec9) ![Agents](https://img.shields.io/badge/Agents-Multi--role-ff6b6b)

## UI Features

- **Dark theme** dashboard with purple/teal accent colors
- **Sidebar** with live stats (active agents, cost, tokens)
- **Agent cards** with status indicators (pulsing green for running)
- **Real-time log viewer** with color-coded levels (info/success/warn/error)
- **Role selection** grid with emoji icons (Researcher, Coder, Reviewer, Writer, Analyst, Tester)
- **Stop & Steer** controls per agent
- **Toast notifications** for agent events
- **WebSocket auto-reconnect** with connection status badge

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
