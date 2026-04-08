# TicTacGo Service

The backend authority for TicTacGo, built on **Nakama Server** with a custom **Go** game logic plugin.

## 🚀 Features
- **Real-time Matchmaking**: Automatically pairs players in Classic or Timed modes.
- **Authoritative Logic**: Game state, turn validation, and winner detection handled server-side in Go.
- **Match State Persistence**: Managed by Nakama's high-performance engine.
- **Docker Ready**: Fully containerized with PostgreSQL support.

## 🛠️ Tech Stack
- **Nakama 3.22.0**: Game server and social framework.
- **Go**: Backend plugin logic (`/cmd/plugin`).
- **PostgreSQL**: Relational storage for player data and sessions.
- **Docker**: Container orchestration.

## 🏁 Getting Started

### Prerequisites
- Docker & Docker Compose
- Go 1.22+ (for local plugin development)

### Local Setup
1. **Clone and Run**:
   ```bash
   docker-compose up --build
   ```
2. **Nakama Console**: Access the dashboard at [http://localhost:7351](http://localhost:7351) (Username: `admin`, Password: `password`).
3. **API Endpoint**: `http://localhost:7350`