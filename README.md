# 🐳 go_spin

**Scheduled Docker Container Management**

go_spin is a Go application for scheduled management of Docker containers. Define containers, groups, and schedules with timers to automatically start/stop containers based on configured times and days.

## ✨ Features

- **Container Management**: Register and manage Docker containers with friendly names and URLs
- **Groups**: Organize containers into logical groups for batch operations
- **Schedules**: Define time-based schedules with multiple timers per target
- **Automatic Start/Stop**: Containers are automatically started/stopped based on schedules
- **Web UI**: Modern SPA interface built with Alpine.js for visual management
- **REST API**: Full JSON API for programmatic access
- **File Watching**: Auto-reload configuration when the JSON file changes externally
- **Graceful Shutdown**: Proper cleanup on application termination

## 🚀 Quick Start

### Prerequisites

- Go 1.21+ 
- Docker

### Installation

```bash
# Clone the repository
git clone https://github.com/bassista/go_spin.git
cd go_spin

# Build
go build -o .build/main ./cmd/server/main.go

# Run
./.build/main
```

### Access

- **Web UI**: http://localhost:8084/ui
- **API**: http://localhost:8084/
- **Health Check**: http://localhost:8084/health

## ⚙️ Configuration

### Configuration File

Create `config/config.yaml`:

```yaml
server:
  port: 8084
  shutdown_timeout_secs: 5
  read_timeout_secs: 10
  write_timeout_secs: 10
  idle_timeout_secs: 120

data:
  file_path: ./config/data/config.json
  persist_interval_secs: 5

misc:
  gin_mode: release              # "debug" or "release"
  scheduling_enabled: true
  scheduling_poll_interval_secs: 30
  scheduling_timezone: "Local"   # or "Europe/Rome", "UTC", etc.
  runtime_type: docker           # "docker" or "memory"
  cors_allowed_origins: "*"      # CORS origins, default "*"
```

### Environment Variables

All settings can be overridden via environment variables with prefix `GO_SPIN_`:

```bash
# Server port
PORT=8084

# Gin mode
GO_SPIN_MISC_GIN_MODE=debug

# Runtime type (docker or memory for testing)
GO_SPIN_MISC_RUNTIME_TYPE=docker

# CORS allowed origins
GO_SPIN_MISC_CORS_ALLOWED_ORIGINS=*

# Config path
GO_SPIN_CONFIG_PATH=./config
```

### .env File

Environment variables can be provided also via .env file.
Create a `.env` file in the project root:

```env
PORT=8084
GO_SPIN_MISC_GIN_MODE=debug
GO_SPIN_MISC_RUNTIME_TYPE=memory
```

## 📡 API Endpoints

### Health
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |

### Containers
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/containers` | List all containers |
| POST | `/container` | Create/update container |
| DELETE | `/container/:name` | Delete container |

### Groups
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/groups` | List all groups |
| POST | `/group` | Create/update group |
| DELETE | `/group/:name` | Delete group |

### Schedules
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/schedules` | List all schedules |
| POST | `/schedule` | Create/update schedule |
| DELETE | `/schedule/:id` | Delete schedule |

### Runtime Control
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/runtime/:name/status` | Check if container is running |
| POST | `/runtime/:name/start` | Start container |
| POST | `/runtime/:name/stop` | Stop container |

### Web UI
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/ui` | Web UI (Alpine.js SPA) |
| GET | `/ui/assets/*` | Static assets |

## 📦 Data Models

### Container
```json
{
  "name": "nginx",
  "friendly_name": "Web Server",
  "url": "http://localhost:8080",
  "running": false,
  "active": true
}
```

### Group
```json
{
  "name": "WebStack",
  "container": ["nginx", "redis"],
  "active": true
}
```

### Schedule
```json
{
  "id": "schedule-001",
  "target": "nginx",
  "targetType": "container",
  "timers": [
    {
      "startTime": "08:00",
      "stopTime": "18:00",
      "days": [1, 2, 3, 4, 5],
      "active": true
    }
  ]
}
```

> **Days**: 0 = Sunday, 1 = Monday, ..., 6 = Saturday

## 🖥️ Web UI

The web interface provides visual management for:

| Tab | Features |
|-----|----------|
| **Containers** | List, Add, Edit, Delete, Start/Stop |
| **Groups** | List, Add, Edit, Delete, Multi-select containers |
| **Schedules** | List, Add, Edit, Delete, Full timer editor with day selection |

Access the UI at `http://localhost:8084/ui`

## 🛠️ Development

### Hot Reload with Air

```bash
# Linux/macOS
air -c .air.toml

# Windows
air -c .air_win.toml
```

### Docker Development

```bash
# Development with hot-reload
docker-compose -f dev.docker-compose.yml up

# Production
docker-compose up
```

### Testing

```bash
go test ./...
```


## 📊 Coverage Report
👉 [Visualizza il report di coverage](https://bassista.github.io/go_spin/)


## 🏗️ Architecture

```
go_spin/
├── cmd/server/           # Application entrypoint
├── config/               # Configuration files
│   └── data/             # JSON data storage
├── internal/
│   ├── api/
│   │   ├── controller/   # HTTP handlers
│   │   ├── middleware/   # CORS middleware
│   │   └── route/        # Route setup
│   ├── app/              # Application container (DI)
│   ├── cache/            # In-memory store with dirty tracking
│   ├── config/           # Configuration loading
│   ├── repository/       # JSON persistence + file watching
│   ├── runtime/          # Container runtime (Docker/Memory)
│   └── scheduler/        # Polling scheduler for automation
├── ui/                   # Web UI (Alpine.js + TailwindCSS)
│   ├── index.html
│   └── assets/
│       └── app.js
└── docs/                 # Documentation + Postman collection
```

### Key Patterns

- **Interface-driven**: Minimal interfaces for testability
- **Dirty tracking**: Async persistence only when data changes
- **Optimistic locking**: `lastUpdate` timestamp prevents overwrites
- **File watching**: Auto-reload on external config changes

## 📄 License

MIT License

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing`)
3. Commit changes (`git commit -m 'feat: add amazing feature'`)
4. Push to branch (`git push origin feature/amazing`)
5. Open a Pull Request