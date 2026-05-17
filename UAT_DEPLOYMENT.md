# UAT Deployment Guide — 92-Account-Center

## Prerequisites
- Docker & Docker Compose v2+
- Go 1.21+ (for local builds)
- Node.js 20+ (for frontend builds)
- JDK 17+ (for Android build)
- macOS + Xcode 15+ (for iOS build)

## Quick Start

### 1. Environment Setup
```bash
cp .env.example .env
# Edit .env with your values (JWT secrets, DB password, etc.)
```

### 2. Full Stack (Docker)
```bash
# Build all Docker images + start
docker compose build --no-cache
docker compose up -d

# Verify all services healthy
docker compose ps
# All 14 containers should show "Up (healthy)" or "Up"
```

### 3. Local Development

#### Backend Services
```bash
# Start config-service (must be first)
cd config-service
go run ./cmd

# Start auth-service (in separate terminal)
cd auth-service
go run ./cmd
```

#### Frontend (Web UI)
```bash
cd web-ui
npm install
npm run dev  # serves on :30317, proxies /api → api-gateway :30300
```

#### Config Management UI
```bash
cd config-management-ui
npm install
npm run dev  # serves on :30316, proxies /api → config-service :30315
```

#### WeChat Mini Program
- Open `weapp/` in WeChat DevTools
- Set appid in `project.config.json`
- Development mode: use `https://uat-api.accountcenter.com`

#### Android
```bash
cd android
export JAVA_HOME="/path/to/jdk-17"
export ANDROID_HOME="$HOME/Android/Sdk"
./gradlew assembleDebug
# APK at: android/app/build/outputs/apk/debug/
```

#### iOS
```bash
cd ios
xcodegen generate
open AccountCenter.xcodeproj
# Select simulator target, Build and Run
```

## Service Endpoints

| Service | Port | Health Check |
|---------|------|-------------|
| API Gateway | 30300 | `GET /health` |
| Auth Service | 30302 | `GET /health` |
| Account Service | 30301 | `GET /health` |
| Credit Service | 30312 | `GET /health` |
| Compliance Service | 30313 | `GET /health` |
| Notification Service | 30311 | `GET /health` |
| Data Product Service | 30314 | `GET /health` |
| Config Service | 30315 | `GET /health` |
| Config Management UI | 30316 | — |
| Web UI | 30317 | — |
| PostgreSQL | 5432 | — |
| Redis | 6379 | — |
| VictoriaMetrics | 20010 | — |
| Loki | 3100 | — |
| Grafana | 3001 | (admin/${GF_SECURITY_ADMIN_PASSWORD}) |

## Platform Access

| Platform | URL / Path | Notes |
|----------|-----------|-------|
| **Web UI** | `http://localhost:30317` | User-facing portal |
| **Config Mgmt** | `http://localhost:30316` | Admin config panel |
| **Grafana Logs** | `http://localhost:3001` | Loki data source pre-configured |
| **VictoriaMetrics** | `http://localhost:20010` | Prometheus-compatible metrics |

## Verify Deployment

```bash
# Quick health check across all services
for port in 30300 30302 30301 30312 30313 30311 30314 30315; do
  curl -s "http://localhost:$port/health" && echo " :$port OK" || echo " :$port FAIL"
done

# Test login flow
curl -X POST http://localhost:30300/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"credential":"admin@test.com","password":"test123"}'

# Check config via internal API
curl http://localhost:30315/internal/v1/config/items/JWT_ACCESS_TOKEN_EXPIRE

# Verify Prometheus metrics
curl http://localhost:20010/api/v1/query?query=up
```

## UAT Test Accounts

| Role | User ID | Permission |
|------|---------|-----------|
| System Owner | admin | Full access |
| Config Viewer | (create via permission UI) | Read-only config |

*(Create additional users via the Register page)*
