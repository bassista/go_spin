# Migliorie Applicabili alla Codebase

## 🔴 Alta Priorità

| # | Area | Problema | File |
|---|------|----------|------|
| 1 | Docker | Path build errato (cmd/api → server) + Go 1.25.6 inesistente | Dockerfile |
| 2 | Docker | GOARCH=arm64 hardcoded - non portabile | Dockerfile |
| 3 | Sicurezza | Nessuna autenticazione sulle API - chiunque può start/stop container | Tutti i controller |
| 4 | Sicurezza | Container Docker gira come root | Dockerfile |
| 5 | Testing | Zero test (*_test.go) in tutta la codebase | Workspace |
| 6 | Resilienza | Nessun retry/backoff su operazioni Docker | docker_runtime.go |

## 🟠 Media Priorità

| # | Area | Problema | File |
|---|------|----------|------|
| 7 | Performance | Deep-copy via JSON marshal per ogni operazione cache | store.go |
| 8 | Performance | Scheduler chiama IsRunning() per ogni container ad ogni tick (N+1) | polling_scheduler.go |
| 9 | Configuration | Timeout API hardcoded 1s - non configurabile | route.go |
| 10 | Code Quality | Pattern CRUD duplicato in 3 controller | container_controller.go, group_controller.go, schedule_controller.go |
| 11 | Code Quality | Messaggi errore in italiano mischiati con codebase inglese | docker_runtime.go |
| 12 | Docker | Nessun HEALTHCHECK definito | Dockerfile |
| 13 | Observability | Solo log.Printf - nessun log strutturato/livelli | Tutta la codebase |

## 🟢 Bassa Priorità

| # | Area | Problema | File |
|---|------|----------|------|
| 14 | Code Quality | MemoryRuntime.IsRunning non ritorna errore se container non esiste (diverso da Docker) | memory_runtime.go |
| 15 | Resilienza | log.Fatalf su errori watcher - crash invece di graceful degradation | app.go |
| 16 | Config | RuntimeType non validato in config, solo in factory | config.go |
| 17 | UI | Path ui hardcoded - non configurabile | ui_route.go |

## 🐛 Bug Trovati Durante i Test

| # | Area | Problema | File | Fix Suggerita |
|---|------|----------|------|---------------|
| 1 | CORS | Manca header `Vary: Origin` per origini specifiche | cors.go | Aggiungere `c.Header("Vary", "Origin")` quando `allowOrigin != "*"` |

---

## Migliorie che posso implementare subito

Quali vuoi che applichi?

1. **Fix Dockerfile completo** (path, Go version, GOARCH auto, USER non-root, HEALTHCHECK)
2. **Timeout API configurabile** (aggiungere `misc.api_timeout_secs` in config)
3. **Messaggi errore in inglese** (docker_runtime.go)
4. **MemoryRuntime consistente con DockerRuntime** (errore se container non esiste)
5. **Basic Auth opzionale** per proteggere le API
6. **Logging strutturato** con livelli (info/warn/error)
7. **Retry con backoff** su operazioni Docker
