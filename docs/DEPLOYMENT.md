# Deployment & Operations Guide

## 🐳 Docker Deployment

### Production Docker Setup

#### Single Container Deployment

**Dockerfile (Production):**
```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main ./cmd/server/main.go

FROM alpine:latest

# Security: Create non-root user
RUN addgroup -g 1001 -S gospin && \
    adduser -u 1001 -S gospin -G gospin

# Install runtime dependencies
RUN apk --no-cache add ca-certificates curl

WORKDIR /app

# Copy binary and set ownership
COPY --from=builder /app/main .
COPY --from=builder /app/ui ./ui
RUN chown -R gospin:gospin /app

# Create config and data directories
RUN mkdir -p /app/config /app/config/data && \
    chown -R gospin:gospin /app/config

USER gospin

EXPOSE 8084
HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
  CMD curl -f http://localhost:8084/health || exit 1

CMD ["./main"]
```

**docker-compose.yml (Production):**
```yaml
version: '3.8'

services:
  go-spin:
    build: .
    image: go-spin:latest
    container_name: go-spin
    restart: unless-stopped
    
    ports:
      - "8084:8084"
      - "8085:8085"  # Waiting server port
    
    volumes:
      # Docker socket for container management
      - /var/run/docker.sock:/var/run/docker.sock:ro
      
      # Configuration persistence
      - ./config:/app/config:rw
      - go-spin-data:/app/config/data:rw
      
      # Logs (optional)
      - ./logs:/app/logs:rw
    
    environment:
      # Production configuration
      - GO_SPIN_MISC_GIN_MODE=release
      - GO_SPIN_SERVER_PORT=8084
      - WAITING_SERVER_PORT=8085
      
      # Security
      - GO_SPIN_MISC_CORS_ALLOWED_ORIGINS=https://yourdomain.com,https://app.yourdomain.com
      
      # Performance
      - GO_SPIN_DATA_PERSIST_INTERVAL_SECS=30
      - GO_SPIN_MISC_SCHEDULING_POLL_INTERVAL_SECS=60
      
      # Runtime
      - GO_SPIN_MISC_RUNTIME_TYPE=docker
      - GO_SPIN_MISC_SCHEDULING_ENABLED=true
      
    healthcheck:\n      test: [\"CMD\", \"curl\", \"-f\", \"http://localhost:8084/health\"]\n      interval: 30s\n      timeout: 10s\n      retries: 3\n      start_period: 40s\n    \n    # Resource limits\n    deploy:\n      resources:\n        limits:\n          memory: 512M\n          cpus: '1.0'\n        reservations:\n          memory: 256M\n          cpus: '0.5'\n    \n    # Security settings\n    security_opt:\n      - no-new-privileges:true\n    \n    # Logging\n    logging:\n      driver: \"json-file\"\n      options:\n        max-size: \"10m\"\n        max-file: \"3\"\n\nvolumes:\n  go-spin-data:\n    driver: local\n\nnetworks:\n  default:\n    name: go-spin-network\n```\n\n#### Multi-Environment Setup\n\n**Production Environment:**\n```bash\n# Create production configuration\nmkdir -p ./config/prod\ncp config/config.yaml ./config/prod/config.yaml\n\n# Edit production-specific settings\nvim ./config/prod/config.yaml\n\n# Deploy\ndocker-compose -f docker-compose.prod.yml up -d\n```\n\n**Staging Environment:**\n```yaml\n# docker-compose.staging.yml\nversion: '3.8'\n\nservices:\n  go-spin-staging:\n    extends:\n      file: docker-compose.yml\n      service: go-spin\n    container_name: go-spin-staging\n    ports:\n      - \"8184:8084\"  # Different port for staging\n    environment:\n      - GO_SPIN_MISC_GIN_MODE=debug\n      - GO_SPIN_MISC_CORS_ALLOWED_ORIGINS=*\n    volumes:\n      - ./config/staging:/app/config:rw\n```\n\n### Container Registry\n\n**Build and Push:**\n```bash\n# Build multi-arch image\ndocker buildx create --use\ndocker buildx build --platform linux/amd64,linux/arm64 \\\n  -t youregistry.com/go-spin:latest \\\n  -t youregistry.com/go-spin:v1.2.3 \\\n  --push .\n\n# Deploy from registry\ndocker pull youregistry.com/go-spin:latest\ndocker run -d \\\n  --name go-spin \\\n  -p 8084:8084 \\\n  -v /var/run/docker.sock:/var/run/docker.sock:ro \\\n  -v ./config:/app/config \\\n  youregistry.com/go-spin:latest\n```\n\n## 🖥️ Bare Metal Deployment\n\n### System Service (systemd)\n\n**Service File (`/etc/systemd/system/go-spin.service`):**\n```ini\n[Unit]\nDescription=go_spin Container Scheduler\nDocumentation=https://github.com/bassista/go_spin\nAfter=network.target docker.service\nRequires=docker.service\n\n[Service]\n# Service configuration\nType=simple\nUser=gospin\nGroup=docker\nWorkingDirectory=/opt/go-spin\nExecStart=/opt/go-spin/main\nExecReload=/bin/kill -HUP $MAINPID\nRestart=always\nRestartSec=10\nKillMode=mixed\nKillSignal=SIGTERM\nTimeoutStopSec=30\n\n# Environment\nEnvironment=GO_SPIN_CONFIG_PATH=/opt/go-spin/config\nEnvironment=GO_SPIN_MISC_GIN_MODE=release\nEnvironmentFile=-/opt/go-spin/.env\n\n# Security hardening\nNoNewPrivileges=true\nPrivateTmp=true\nPrivateDevices=true\nProtectHome=true\nProtectSystem=strict\nReadWritePaths=/opt/go-spin/config /opt/go-spin/logs\nProtectKernelTunables=true\nProtectKernelModules=true\nProtectControlGroups=true\nRestrictNamespaces=true\nRestrictRealtime=true\nRestrictSUIDSGID=true\nRemoveIPC=true\nLockPersonality=true\n\n# Resource limits\nMemoryMax=512M\nTasksMax=50\n\n# Logging\nStandardOutput=journal\nStandardError=journal\nSyslogIdentifier=go-spin\n\n[Install]\nWantedBy=multi-user.target\n```\n\n**Installation:**\n```bash\n# Create user and directories\nsudo useradd -r -s /bin/false -d /opt/go-spin gospin\nsudo usermod -aG docker gospin\nsudo mkdir -p /opt/go-spin/{config,logs}\n\n# Copy binary and configuration\nsudo cp main /opt/go-spin/\nsudo cp -r config/ /opt/go-spin/\nsudo cp -r ui/ /opt/go-spin/\n\n# Set permissions\nsudo chown -R gospin:gospin /opt/go-spin\nsudo chmod +x /opt/go-spin/main\n\n# Install service\nsudo cp go-spin.service /etc/systemd/system/\nsudo systemctl daemon-reload\nsudo systemctl enable go-spin\nsudo systemctl start go-spin\n\n# Check status\nsudo systemctl status go-spin\nsudo journalctl -u go-spin -f\n```\n\n### Process Manager (PM2)\n\n**PM2 Configuration (`ecosystem.config.js`):**\n```javascript\nmodule.exports = {\n  apps: [{\n    name: 'go-spin',\n    script: './main',\n    cwd: '/opt/go-spin',\n    instances: 1,\n    exec_mode: 'fork',\n    \n    // Auto-restart configuration\n    autorestart: true,\n    watch: false,\n    max_memory_restart: '512M',\n    restart_delay: 5000,\n    \n    // Environment\n    env: {\n      GO_SPIN_CONFIG_PATH: '/opt/go-spin/config',\n      GO_SPIN_MISC_GIN_MODE: 'release'\n    },\n    \n    // Logging\n    log_file: '/opt/go-spin/logs/combined.log',\n    out_file: '/opt/go-spin/logs/out.log',\n    error_file: '/opt/go-spin/logs/error.log',\n    log_type: 'json',\n    merge_logs: true,\n    \n    // Monitoring\n    min_uptime: '30s',\n    max_restarts: 10\n  }]\n};\n```\n\n**Usage:**\n```bash\n# Install PM2\nnpm install -g pm2\n\n# Start application\npm2 start ecosystem.config.js\n\n# Management commands\npm2 status\npm2 logs go-spin\npm2 restart go-spin\npm2 stop go-spin\n\n# Auto-startup\npm2 startup\npm2 save\n```\n\n## 🔒 Reverse Proxy & SSL\n\n### Nginx Configuration\n\n**Main Configuration (`/etc/nginx/sites-available/go-spin`):**\n```nginx\nupstream go_spin_backend {\n    server 127.0.0.1:8084 max_fails=3 fail_timeout=30s;\n    keepalive 32;\n}\n\nupstream go_spin_waiting {\n    server 127.0.0.1:8085 max_fails=3 fail_timeout=30s;\n}\n\n# HTTP to HTTPS redirect\nserver {\n    listen 80;\n    server_name go-spin.yourdomain.com;\n    return 301 https://$server_name$request_uri;\n}\n\n# Main HTTPS server\nserver {\n    listen 443 ssl http2;\n    server_name go-spin.yourdomain.com;\n    \n    # SSL Configuration\n    ssl_certificate /etc/letsencrypt/live/go-spin.yourdomain.com/fullchain.pem;\n    ssl_certificate_key /etc/letsencrypt/live/go-spin.yourdomain.com/privkey.pem;\n    ssl_session_timeout 1d;\n    ssl_session_cache shared:SSL:50m;\n    ssl_session_tickets off;\n    \n    # Modern SSL configuration\n    ssl_protocols TLSv1.2 TLSv1.3;\n    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384;\n    ssl_prefer_server_ciphers off;\n    \n    # Security headers\n    add_header Strict-Transport-Security \"max-age=63072000\" always;\n    add_header X-Frame-Options DENY always;\n    add_header X-Content-Type-Options nosniff always;\n    add_header X-XSS-Protection \"1; mode=block\" always;\n    add_header Referrer-Policy \"strict-origin-when-cross-origin\" always;\n    \n    # Rate limiting\n    limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;\n    \n    # Main application\n    location / {\n        proxy_pass http://go_spin_backend;\n        proxy_http_version 1.1;\n        proxy_set_header Upgrade $http_upgrade;\n        proxy_set_header Connection 'upgrade';\n        proxy_set_header Host $host;\n        proxy_set_header X-Real-IP $remote_addr;\n        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;\n        proxy_set_header X-Forwarded-Proto $scheme;\n        proxy_cache_bypass $http_upgrade;\n        \n        # Timeouts\n        proxy_connect_timeout 30s;\n        proxy_send_timeout 300s;\n        proxy_read_timeout 300s;\n        \n        # Rate limiting for API\n        limit_req zone=api burst=20 nodelay;\n    }\n    \n    # Waiting server (separate upstream)\n    location /waiting {\n        proxy_pass http://go_spin_waiting;\n        proxy_set_header Host $host;\n        proxy_set_header X-Real-IP $remote_addr;\n        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;\n        proxy_set_header X-Forwarded-Proto $scheme;\n    }\n    \n    # Static assets with caching\n    location /ui/assets/ {\n        proxy_pass http://go_spin_backend;\n        expires 1y;\n        add_header Cache-Control \"public, immutable\";\n    }\n    \n    # Health check (no rate limiting)\n    location /health {\n        proxy_pass http://go_spin_backend;\n        access_log off;\n    }\n    \n    # Logging\n    access_log /var/log/nginx/go-spin.access.log;\n    error_log /var/log/nginx/go-spin.error.log;\n}\n```\n\n**Setup Script:**\n```bash\n# Install Certbot for Let's Encrypt\nsudo apt-get install certbot python3-certbot-nginx\n\n# Obtain SSL certificate\nsudo certbot --nginx -d go-spin.yourdomain.com\n\n# Enable site\nsudo ln -s /etc/nginx/sites-available/go-spin /etc/nginx/sites-enabled/\nsudo nginx -t\nsudo systemctl reload nginx\n```\n\n### Traefik Configuration\n\n**docker-compose.yml with Traefik:**\n```yaml\nversion: '3.8'\n\nservices:\n  traefik:\n    image: traefik:v2.9\n    container_name: traefik\n    restart: unless-stopped\n    command:\n      - \"--api.dashboard=true\"\n      - \"--providers.docker=true\"\n      - \"--entrypoints.web.address=:80\"\n      - \"--entrypoints.websecure.address=:443\"\n      - \"--certificatesresolvers.letsencrypt.acme.email=admin@yourdomain.com\"\n      - \"--certificatesresolvers.letsencrypt.acme.storage=/acme.json\"\n      - \"--certificatesresolvers.letsencrypt.acme.httpchallenge.entrypoint=web\"\n    ports:\n      - \"80:80\"\n      - \"443:443\"\n    volumes:\n      - /var/run/docker.sock:/var/run/docker.sock:ro\n      - traefik-acme:/acme.json\n    labels:\n      - \"traefik.http.routers.api.rule=Host(`traefik.yourdomain.com`)\"\n      - \"traefik.http.routers.api.tls.certresolver=letsencrypt\"\n      \n  go-spin:\n    build: .\n    container_name: go-spin\n    restart: unless-stopped\n    volumes:\n      - /var/run/docker.sock:/var/run/docker.sock:ro\n      - ./config:/app/config\n    labels:\n      - \"traefik.enable=true\"\n      - \"traefik.http.routers.go-spin.rule=Host(`go-spin.yourdomain.com`)\"\n      - \"traefik.http.routers.go-spin.tls.certresolver=letsencrypt\"\n      - \"traefik.http.services.go-spin.loadbalancer.server.port=8084\"\n      \nvolumes:\n  traefik-acme:\n```\n\n## 📊 Monitoring & Observability\n\n### Health Checks\n\n**Application Health Check:**\n```bash\n#!/bin/bash\n# health-check.sh\n\nHOST=\"localhost:8084\"\nTIMEOUT=10\n\n# Basic health endpoint\nif ! curl -f -s --max-time $TIMEOUT \"http://$HOST/health\" > /dev/null; then\n    echo \"ERROR: Health check failed\"\n    exit 1\nfi\n\n# Check if containers endpoint responds\nif ! curl -f -s --max-time $TIMEOUT \"http://$HOST/containers\" > /dev/null; then\n    echo \"ERROR: Containers endpoint failed\"\n    exit 1\nfi\n\necho \"OK: Health checks passed\"\nexit 0\n```\n\n**Docker Health Check:**\n```dockerfile\nHEALTHCHECK --interval=30s --timeout=10s --retries=3 --start-period=30s \\\n  CMD [\"./health-check.sh\"]\n```\n\n### Prometheus Metrics\n\n**Add metrics endpoint (future enhancement):**\n```go\n// internal/metrics/prometheus.go\npackage metrics\n\nimport (\n    \"github.com/prometheus/client_golang/prometheus\"\n    \"github.com/prometheus/client_golang/prometheus/promauto\"\n)\n\nvar (\n    containersTotal = promauto.NewGaugeVec(\n        prometheus.GaugeOpts{\n            Name: \"go_spin_containers_total\",\n            Help: \"Total number of managed containers\",\n        },\n        []string{\"status\"},\n    )\n    \n    scheduleExecutions = promauto.NewCounterVec(\n        prometheus.CounterOpts{\n            Name: \"go_spin_schedule_executions_total\",\n            Help: \"Total number of schedule executions\",\n        },\n        []string{\"action\", \"status\"},\n    )\n)\n```\n\n**Prometheus Configuration (`prometheus.yml`):**\n```yaml\nglobal:\n  scrape_interval: 15s\n  \nscrape_configs:\n  - job_name: 'go-spin'\n    static_configs:\n      - targets: ['localhost:8084']\n    metrics_path: '/metrics'\n    scrape_interval: 30s\n```\n\n### Logging\n\n**Centralized Logging with ELK:**\n```yaml\n# docker-compose.logging.yml\nversion: '3.8'\n\nservices:\n  go-spin:\n    # ... existing configuration\n    logging:\n      driver: \"fluentd\"\n      options:\n        fluentd-address: localhost:24224\n        tag: go-spin\n        \n  fluentd:\n    image: fluent/fluentd:v1.14\n    container_name: fluentd\n    ports:\n      - \"24224:24224\"\n    volumes:\n      - ./fluentd/conf:/fluentd/etc\n      - ./fluentd/logs:/var/log/fluentd\n```\n\n**Structured Logging Example:**\n```bash\n# Query logs with jq\ndocker logs go-spin 2>&1 | jq 'select(.level == \"error\")'\ndocker logs go-spin 2>&1 | jq 'select(.component == \"scheduler\")'\n```\n\n## 🔧 Maintenance\n\n### Backup Strategy\n\n**Configuration Backup:**\n```bash\n#!/bin/bash\n# backup-config.sh\n\nBACKUP_DIR=\"/backup/go-spin/$(date +%Y%m%d_%H%M%S)\"\nCONFIG_DIR=\"/opt/go-spin/config\"\n\nmkdir -p \"$BACKUP_DIR\"\n\n# Backup configuration and data\ncp -r \"$CONFIG_DIR\" \"$BACKUP_DIR/\"\n\n# Create archive\ntar -czf \"$BACKUP_DIR.tar.gz\" -C \"$(dirname $BACKUP_DIR)\" \"$(basename $BACKUP_DIR)\"\nrm -rf \"$BACKUP_DIR\"\n\necho \"Backup created: $BACKUP_DIR.tar.gz\"\n\n# Cleanup old backups (keep last 7 days)\nfind /backup/go-spin -name \"*.tar.gz\" -mtime +7 -delete\n```\n\n**Automated Backup (Cron):**\n```bash\n# Add to crontab\n0 2 * * * /opt/go-spin/scripts/backup-config.sh\n```\n\n### Update Procedures\n\n**Rolling Update (Docker):**\n```bash\n#!/bin/bash\n# update.sh\n\nset -e\n\necho \"Starting go-spin update...\"\n\n# Backup current configuration\n./backup-config.sh\n\n# Pull new image\ndocker-compose pull go-spin\n\n# Stop current container\ndocker-compose stop go-spin\n\n# Start with new image\ndocker-compose up -d go-spin\n\n# Wait for health check\necho \"Waiting for application to be healthy...\"\nsleep 30\n\n# Verify health\nif curl -f http://localhost:8084/health; then\n    echo \"Update completed successfully\"\n    # Cleanup old images\n    docker image prune -f\nelse\n    echo \"Health check failed, rolling back...\"\n    docker-compose down\n    # Restore previous version if needed\n    exit 1\nfi\n```\n\n### Performance Tuning\n\n**Go Runtime Tuning:**\n```bash\n# Environment variables for production\nexport GOGC=100                    # GC target percentage\nexport GOMAXPROCS=2               # Limit CPU usage\nexport GODEBUG=gctrace=1          # GC debugging (dev only)\n```\n\n**Docker Resource Limits:**\n```yaml\nservices:\n  go-spin:\n    deploy:\n      resources:\n        limits:\n          memory: 512M\n          cpus: '1.0'\n        reservations:\n          memory: 256M\n          cpus: '0.5'\n    ulimits:\n      nofile:\n        soft: 1024\n        hard: 2048\n```\n\n### Troubleshooting Guide\n\n**Common Issues:**\n\n1. **High Memory Usage:**\n   ```bash\n   # Check memory usage\n   docker stats go-spin\n   \n   # Tune GC if needed\n   export GOGC=50  # More aggressive GC\n   ```\n\n2. **Docker Socket Permission Issues:**\n   ```bash\n   # Check socket permissions\n   ls -la /var/run/docker.sock\n   \n   # Fix permissions\n   sudo chmod 666 /var/run/docker.sock\n   # Or add user to docker group\n   sudo usermod -aG docker $USER\n   ```\n\n3. **Configuration File Corruption:**\n   ```bash\n   # Validate JSON\n   jq . config/data/config.json\n   \n   # Restore from backup if invalid\n   cp backup/latest/config/data/config.json config/data/\n   ```\n\n4. **Schedule Not Executing:**\n   ```bash\n   # Check logs for scheduler errors\n   docker logs go-spin | grep scheduler\n   \n   # Verify timezone\n   docker exec go-spin date\n   \n   # Check schedule configuration\n   curl http://localhost:8084/schedules | jq .\n   ```\n\n**Log Analysis:**\n```bash\n# Performance analysis\ndocker logs go-spin | grep -E \"duration|elapsed\" | tail -20\n\n# Error analysis\ndocker logs go-spin | grep -E \"ERROR|FATAL\" | tail -10\n\n# Schedule analysis\ndocker logs go-spin | grep schedule | tail -20\n```\n\nThis deployment guide provides comprehensive instructions for various deployment scenarios and operational concerns. Choose the deployment method that best fits your infrastructure and security requirements.