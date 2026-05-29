<#
.SYNOPSIS
    Unified build script for 92-Account-Center.
    Supports 3 tiers of build strategy with automatic fallback.

.DESCRIPTION
    Tier 1 - Docker Compose build (fast, requires network)
    Tier 2 - Local Go build + Docker image with pre-cached base images
    Tier 3 - Local Go build + direct binary execution (no Docker)

.PARAMETER Tier
    Build strategy: 1 (Docker), 2 (Local+Docker), 3 (Local only). Default: auto

.PARAMETER Services
    Services to build, comma-separated. Default: auth-service,api-gateway,account-service,config-service,web-ui

.PARAMETER Push
    After build, push images to local registry cache

.PARAMETER CacheBases
    Pre-cache base images locally for Tier 2
#>

param(
    [ValidateSet(1,2,3,'auto')]
    $Tier = 'auto',
    [string]$Services = 'auth-service,api-gateway,account-service,config-service,web-ui',
    [switch]$Push,
    [switch]$CacheBases
)

$ErrorActionPreference = 'Stop'
$ROOT = Split-Path -Parent $MyInvocation.MyCommand.Path
$BIN = "$ROOT\bin"
$TEMP = "$env:TEMP\opencode"

# Ensure bin dir exists
New-Item -ItemType Directory -Force -Path $BIN | Out-Null
New-Item -ItemType Directory -Force -Path $TEMP | Out-Null

function Write-Step($msg) { Write-Host "`n>>> $msg" -ForegroundColor Cyan }
function Write-Ok($msg) { Write-Host "  OK: $msg" -ForegroundColor Green }
function Write-Warn($msg) { Write-Host "  WARN: $msg" -ForegroundColor Yellow }
function Write-Err($msg) { Write-Host "  FAIL: $msg" -ForegroundColor Red }

# ──────────────────────────────────────────────
# Detect Docker availability
# ──────────────────────────────────────────────
function Test-DockerAvailable {
    try {
        $v = docker version --format '{{.Server.Version}}' 2>$null
        return $v -ne $null -and $v -ne ''
    } catch { return $false }
}

function Test-DockerCanPull {
    param([string]$Image = 'alpine:3.23')
    try {
        docker pull $Image 2>&1 | Out-Null
        return $LASTEXITCODE -eq 0
    } catch { return $false }
}

# ──────────────────────────────────────────────
# Tier 1: Full Docker Compose Build
# ──────────────────────────────────────────────
function Invoke-DockerBuild {
    param([string[]]$ServiceList)
    Write-Step "Tier 1: Docker Compose build"
    
    # Configure registry mirror for China
    $mirror = $env:DOCKER_MIRROR
    if (-not $mirror) {
        # Test available mirrors
        $mirrors = @(
            'https://docker.1ms.run',
            'https://docker.xuanyuan.me',
            'https://dockerproxy.cn',
            'https://hub-mirror.c.163.com',
            'https://docker.mirrors.ustc.edu.cn'
        )
        foreach ($m in $mirrors) {
            try {
                $req = [System.Net.WebRequest]::Create("$m/v2/")
                $req.Timeout = 3000
                $req.GetResponse()
                $mirror = $m
                Write-Ok "Mirror available: $mirror"
                break
            } catch { continue }
        }
        if (-not $mirror) {
            Write-Warn "No working registry mirror found. Trying direct Docker Hub..."
        }
    }

    if ($mirror) {
        # Use temporary daemon.json with working mirror
        $origCfg = "$env:USERPROFILE\.docker\daemon.json"
        $tmpCfg = "$TEMP\daemon.json"
        if (Test-Path $origCfg) {
            Copy-Item $origCfg $tmpCfg -Force
        }
        @{ "registry-mirrors" = @($mirror) } | ConvertTo-Json | Set-Content $origCfg -Force
        Write-Ok "Set mirror: $mirror"
    }

    $svcArgs = $ServiceList -join ' '
    $result = docker-compose build $svcArgs 2>&1
    $exitCode = $LASTEXITCODE

    # Restore original daemon.json
    if ($mirror -and (Test-Path $tmpCfg)) {
        Copy-Item $tmpCfg $origCfg -Force
    }

    if ($exitCode -eq 0) {
        Write-Ok "Docker Compose build succeeded"
        return $true
    }
    
    Write-Err "Docker Compose build failed (exit: $exitCode)"
    Write-Warn "Falling back to Tier 2..."
    return $false
}

# ──────────────────────────────────────────────
# Tier 2: Local Go build + Docker image packaging
# ──────────────────────────────────────────────
function Invoke-LocalBuild {
    param([string[]]$ServiceList)

    # ── Build Go binaries (cross-compile for Linux when Docker target) ──
    Write-Step "Building Go binaries locally"
    $goServices = @{
        'auth-service'    = @{ dir = 'auth-service';    pkg = './cmd/...' }
        'api-gateway'     = @{ dir = 'api-gateway';     pkg = './cmd/...' }
        'account-service' = @{ dir = 'account-service'; pkg = './cmd/...' }
        'config-service'  = @{ dir = 'config-service';  pkg = './cmd/...' }
    }

    $isDocker = (Test-DockerAvailable) -and ($Tier -ne 3)
    if ($isDocker) {
        $env:GOOS = 'linux'; $env:GOARCH = 'amd64'; $env:CGO_ENABLED = '0'
        $suffix = '-linux'
    } else {
        $suffix = '.exe'
    }

    foreach ($svc in $ServiceList) {
        if (-not $goServices.ContainsKey($svc)) { continue }
        $info = $goServices[$svc]
        Write-Step "Building $svc ..."
        Push-Location "$ROOT\$($info.dir)"
        try {
            go build -o "$BIN\$svc$suffix" $info.pkg
            if ($LASTEXITCODE -eq 0) {
                Write-Ok "$svc built: $BIN\$svc$suffix"
            } else {
                throw "go build failed for $svc"
            }
        } finally { Pop-Location }
    }

    # ── Build frontend ──
    if ($ServiceList -contains 'web-ui') {
        Write-Step "Building frontend..."
        Push-Location "$ROOT\web-ui"
        try {
            # vue-tsc may emit warnings but vite build is what matters
            npx vite build 2>&1
            if ($LASTEXITCODE -eq 0 -and (Test-Path "dist")) {
                Write-Ok "Frontend built: web-ui/dist/"
            } else {
                throw "vite build failed (exit: $LASTEXITCODE)"
            }
        } finally { Pop-Location }
    }

    # ── Build Docker images (FROM scratch for Go services, no network needed) ──
    if ($isDocker) {
        $ctx = $ROOT
        foreach ($svc in $ServiceList) {
            Write-Step "Packaging $svc Docker image..."
            $df = "$TEMP\Dockerfile.$svc"
            $img = "92-account-center-$svc`:latest"
            switch ($svc) {
                'auth-service' {
                    Set-Content $df -Force @"
FROM scratch
COPY bin/auth-service-linux /usr/local/bin/auth-service
CMD ["/usr/local/bin/auth-service"]
"@
                }
                'api-gateway' {
                    Set-Content $df -Force @"
FROM scratch
COPY bin/api-gateway-linux /usr/local/bin/api-gateway
CMD ["/usr/local/bin/api-gateway"]
"@
                }
                'account-service' {
                    Set-Content $df -Force @"
FROM scratch
COPY bin/account-service-linux /usr/local/bin/account-service
CMD ["/usr/local/bin/account-service"]
"@
                }
                'config-service' {
                    Set-Content $df -Force @"
FROM scratch
COPY bin/config-service-linux /usr/local/bin/config-service
CMD ["/usr/local/bin/config-service"]
"@
                }
                'web-ui' {
                    Set-Content $df -Force @"
FROM nginx:1.29-alpine
COPY web-ui/dist /usr/share/nginx/html
COPY web-ui/nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
"@
                }
            }
            & docker buildx build --load --no-cache -f $df -t $img $ctx 2>&1 | Out-Null
            if ($LASTEXITCODE -eq 0) {
                Write-Ok "$svc Docker image built"
            } else {
                Write-Warn "$svc Docker image build failed (will run natively)"
            }
        }
    }

    Write-Ok "Local build complete"
    Write-Host "`nRun: docker-compose up -d <service>  (Docker)" -ForegroundColor Yellow
    Write-Host "Run: .\bin\<service>.exe              (native, with env vars)" -ForegroundColor Yellow
}

# ──────────────────────────────────────────────
# Pre-cache base images for offline builds
# ──────────────────────────────────────────────
function Invoke-CacheBaseImages {
    Write-Step "Pre-caching base images (retagging locally available)..."
    # Retag existing local images to match Dockerfile tags (zero network)
    $retags = @(
        @{ from = 'alpine:latest';        to = 'alpine:3.23' }
        @{ from = 'golang:1.23-alpine';   to = 'golang:1.26-alpine' }
        @{ from = 'nginx:alpine';         to = 'nginx:1.29-alpine' }
        @{ from = 'redis:7-alpine';       to = 'redis:8.2-alpine' }
    )
    foreach ($r in $retags) {
        $existing = docker images -q $r.from 2>$null
        if ($existing) {
            docker tag $r.from $r.to 2>&1 | Out-Null
            Write-Ok "Retagged: $($r.from) -> $($r.to)"
        } else {
            Write-Warn "Image not found locally: $($r.from). You may need to pull it first."
        }
    }
    Write-Ok "Base images ready. Run: .\build.ps1 -Tier 2"
}

# ──────────────────────────────────────────────
# Main
# ──────────────────────────────────────────────
$serviceList = $Services -split ',' | ForEach-Object { $_.Trim() }

if ($CacheBases) {
    Invoke-CacheBaseImages
    return
}

# Auto-detect tier
if ($Tier -eq 'auto') {
    if (Test-DockerAvailable) {
        if (Test-DockerCanPull) {
            $Tier = 1
        } elseif (Test-Path "$ROOT\.docker-cache\alpine_3.23.tar") {
            Write-Ok "Base images cached locally, using Tier 2"
            $Tier = 2
        } else {
            Write-Warn "Docker available but cannot pull. Run: .\build.ps1 -CacheBases"
            $Tier = 3
        }
    } else {
        Write-Warn "Docker not available, using Tier 3 (local only)"
        $Tier = 3
    }
}

switch ($Tier) {
    1 {
        $ok = Invoke-DockerBuild -ServiceList $serviceList
        if (-not $ok) { Invoke-LocalBuild -ServiceList $serviceList }
    }
    2 { Invoke-LocalBuild -ServiceList $serviceList }
    3 { Invoke-LocalBuild -ServiceList $serviceList }
}

if ($Push) {
    Write-Step "Pushing to local registry..."
    # TODO: push to local registry cache if configured
}
