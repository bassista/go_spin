# Copilot Instructions for go_spin

## Overview
**go_spin** è un'applicazione Go per la gestione schedulata di container Docker. Permette di definire container, gruppi e schedule con timer per avviare/fermare automaticamente i container in base a orari e giorni configurati. Include una UI web SPA per la gestione visuale.

## Architettura

### Struttura dei moduli (`internal/`)
```
internal/
├── api/              # Layer HTTP (Gin)
│   ├── controller/   # Handler per containers, groups, schedules, runtime
│   ├── middleware/   # CORS middleware
│   └── route/        # Setup delle route per ogni risorsa + UI
├── app/              # Application container (dependency injection)
├── cache/            # In-memory store con dirty-tracking
├── config/           # Configurazione via Viper + dotenv
├── repository/       # Persistenza JSON con file-watching (fsnotify)
├── runtime/          # Interfaccia ContainerRuntime (Docker/Memory) + factory
└── scheduler/        # PollingScheduler per automazione start/stop
```

### Struttura UI (`ui/`)
```
ui/
├── index.html        # SPA Alpine.js + TailwindCSS
└── assets/
    └── app.js        # Logica Alpine.js per CRUD containers/groups/schedules
```

### Flusso dati principale
1. **Load**: `JSONRepository` carica `config/data/config.json` → validazione → `DataDocument`
2. **Cache**: `Store` mantiene copia in-memory con `dirty` flag per modifiche pending
3. **API**: Controllers modificano cache, **non persistono direttamente**
4. **Persistence**: Goroutine schedulata (`cache.StartPersistenceScheduler`) salva periodicamente se dirty
5. **File-watching**: `fsnotify` rileva modifiche esterne al file JSON e ricarica automaticamente

### Pattern chiave
- **Interface-driven**: Ogni modulo espone interfacce minimali (es. `ContainerRuntime`, `Repository`, `AppStore`)
- **Dirty-tracking**: Modifiche via API marcano cache come dirty; salvataggio asincrono evita I/O bloccante
- **Optimistic locking**: `metadata.lastUpdate` (Unix ms) previene sovrascritture di modifiche esterne

## Modelli dati (`internal/repository/model.go`)
```go
DataDocument {
    Metadata   { LastUpdate int64 }       // Optimistic locking
    Containers []Container                // name, friendly_name, url, running, active
    Order      []string                   // Ordinamento containers
    Groups     []Group                    // Raggruppamenti di container
    Schedules  []Schedule                 // Timer per start/stop automatici
}
```

## Configurazione

### File: `config/config.yaml`
```yaml
server:
  port: 8080
  shutdown_timeout_secs: 5
data:
  file_path: ./config/data/config.json
  persist_interval_secs: 5
misc:
  scheduling_enabled: true
  scheduling_poll_interval_secs: 30
  scheduling_timezone: "Local"
  runtime_type: "docker"           # "docker" o "memory"
  cors_allowed_origins: "*"        # CORS origins (default "*")
```

### Environment variables (prefix: `GO_SPIN_`)
- `GO_SPIN_CONFIG_PATH`: percorso cartella config (default: `./config`)
- `GO_SPIN_MISC.GIN_MODE`: `debug` o `release`
- `GO_SPIN_MISC.RUNTIME_TYPE`: `docker` o `memory`
- `GO_SPIN_MISC.CORS_ALLOWED_ORIGINS`: origini CORS permesse
- Supporto `.env` via `godotenv`

### Auto-creazione directory
Se la directory del file dati (`data.file_path`) non esiste, viene creata automaticamente all'avvio.

## Developer Workflows

### Build & Run
```bash
go build -o .build/main ./cmd/server/main.go
./.build/main
```

Always execute go vet ./... before building, testing, or committing code. ensure code passes go fmt ./... for consistent formatting. ensure code passes golangci-lint run for linting. ensure all dependencies are properly managed with go mod tidy. ensure all unit tests pass with go test ./... before pushing changes. ensure commit messages follow conventional commit standards. ensure code is documented with comments where necessary for clarity. ensure pull requests include a clear description of changes made. ensure code reviews are conducted for all pull requests before merging. ensure branch protection rules are in place to enforce checks before merging. ensure continuous integration is set up to run tests and checks on each push. ensure code coverage is monitored and maintained at an acceptable level. ensure security vulnerabilities are regularly scanned and addressed. ensure dependencies are kept up to date with regular reviews. ensure coding standards and best practices are followed throughout the codebase. ensure proper error handling is implemented consistently. ensure logging is used effectively for debugging and monitoring. ensure performance considerations are taken into account during development. ensure scalability is considered in the architecture and design decisions. ensure user input is validated and sanitized to prevent security issues. ensure sensitive information is not hardcoded and is managed securely. ensure configuration management follows best practices for different environments. ensure documentation is kept up to date with code changes. 
```

### Hot-reload con Air
```bash
air -c .air.toml        # Linux/macOS
air -c .air_win.toml    # Windows
```

### Test
```bash
go test ./...
```

### Docker
```bash
docker-compose up       # Produzione
docker-compose -f dev.docker-compose.yml up  # Sviluppo
```

## Runtime implementations

### `DockerRuntime` (`internal/runtime/docker_runtime.go`)
Usa il client Moby per interagire con Docker daemon:
```go
cli, err := client.New(client.FromEnv)  // Auto API version negotiation
```
- `ContainerInspect` → verifica stato running
- `ContainerStart/Stop` → gestione lifecycle
- Errori "not found" via `errdefs.IsNotFound(err)`

### `MemoryRuntime`
Runtime mock per testing senza Docker socket.

### Runtime Factory (`internal/runtime/factory.go`)
```go
rt, err := runtime.NewRuntimeFromConfig(runtimeType, doc)
// runtimeType: "docker" (default) o "memory"
```

## Web UI (Alpine.js SPA)

Accessibile su `/ui` - gestione visuale di containers, groups e schedules.

### Funzionalità
| Tab | Operazioni |
|-----|------------|
| **Containers** | Lista, Aggiungi, Modifica, Elimina, Start/Stop runtime |
| **Groups** | Lista, Aggiungi, Modifica, Elimina, selezione multi-container |
| **Schedules** | Lista, Aggiungi, Modifica, Elimina + editor completo timers |

### Stack frontend
- **Alpine.js**: Reattività e stato
- **TailwindCSS**: Styling (via CDN)
- **htmx**: Incluso ma non usato (API JSON-based)

### File
- `ui/index.html`: Layout HTML con componenti Alpine
- `ui/assets/app.js`: Logica applicativa (fetch API, form handling)

## CORS

Middleware CORS configurabile in `internal/api/middleware/cors.go`:
- Default: `*` (tutte le origini)
- Configurabile via `misc.cors_allowed_origins`
- Supporta preflight OPTIONS

## API REST (Gin)

| Method | Endpoint | Descrizione |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/containers` | Lista containers |
| POST | `/container` | Crea/aggiorna container |
| DELETE | `/container/:name` | Elimina container |
| GET | `/groups` | Lista gruppi |
| POST | `/group` | Crea/aggiorna gruppo |
| DELETE | `/group/:name` | Elimina gruppo |
| GET | `/schedules` | Lista schedules |
| POST | `/schedule` | Crea/aggiorna schedule |
| DELETE | `/schedule/:id` | Elimina schedule |
| GET | `/runtime/:name/status` | Verifica se container è running |
| POST | `/runtime/:name/start` | Avvia container |
| POST | `/runtime/:name/stop` | Ferma container |
| GET | `/ui` | Web UI (SPA Alpine.js) |
| GET | `/ui/assets/*` | Asset statici UI |

## Convenzioni codice

### Validazione
- Usa `go-playground/validator/v10` con tag struct (`validate:"required,url"`)
- Validazione sia in repository (load/save) che in controllers

### Gestione errori
- Errori custom in `cache/` (es. `ErrContainerNotFound`)
- Wrap errors con `fmt.Errorf("context: %w", err)`

### Concorrenza
- `sync.RWMutex` per protezione cache e runtime memory
- Context propagation per graceful shutdown

### Logging
- Logger dedicati per componenti: `[json-repo]`, `[persist]`, `[sched]`

## Dipendenze principali
- **gin-gonic/gin**: HTTP framework
- **moby/moby/client**: Docker client (versione modulare)
- **spf13/viper**: Configurazione
- **fsnotify/fsnotify**: File watching
- **enrichman/httpgrace**: Graceful shutdown HTTP server
- **go-playground/validator**: Validazione struct
