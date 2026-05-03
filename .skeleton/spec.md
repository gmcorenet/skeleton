# GMCore Skeleton Specification

## Overview

Skeleton is the **app template** used by `gmcore create` to scaffold new applications. It should mirror Symfony 8 structure adapted for Go.

## Current Skeleton Issues

1. Missing most directories
2. No proper config structure
3. No public entry point
4. No var/ structure
5. Missing tests directory
6. No bin/cmd structure

## Required Structure

```
skeleton/
├── .skeleton/            # Skeleton metadata (NOT copied)
│   └── structure.md       # This file
├── bin/                   # Entry point scripts
│   └── console            # CLI entry
├── cmd/                   # Go entry points
│   └── server/
│       └── main.go        # HTTP server entry
├── config/                # Configuration (like Symfony config/)
│   ├── app.yaml          # Main app config
│   ├── routes.yaml       # Route definitions
│   ├── services.yaml     # Service definitions
│   └── packages/         # Package configs
│       ├── cache.yaml
│       ├── database.yaml
│       └── framework.yaml
├── internal/              # Private application code
│   ├── controller/
│   │   └── health.go
│   ├── service/
│   │   └── health.go
│   ├── repository/
│   ├── model/
│   │   └── entity.go
│   ├── middleware/
│   │   └── logging.go
│   └── kernel/
│       └── kernel.go
├── public/                 # Web root
│   └── index.go           # Web entry point
├── resources/             # Static resources
│   ├── views/
│   │   └── base.html
│   ├── migrations/
│   └── translations/
├── var/                   # Runtime (gitignored)
│   ├── cache/
│   └── log/
├── tests/                 # Test files
│   ├── controller/
│   ├── service/
│   └── integration/
├── go.mod                 # Module definition
└── README.md
```

## File Specifications

### bin/console

Shell script entry point:

```bash
#!/bin/bash
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
exec go run "$SCRIPT_DIR/../cmd/server/main.go" "$@"
```

### cmd/server/main.go

```go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/gmcorenet/framework"
)

func main() {
    kernel := framework.NewKernel()

    server := &http.Server{
        Addr:    getAddr(),
        Handler: kernel,
    }

    go func() {
        log.Printf("Starting server on %s", server.Addr)
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server failed: %v", err)
        }
    }()

    waitForShutdown(server)
}

func getAddr() string {
    if addr := os.Getenv("ADDR"); addr != "" {
        return addr
    }
    return ":8080"
}

func waitForShutdown(server *http.Server) {
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("Shutting down server...")

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        log.Fatalf("Server forced to shutdown: %v", err)
    }

    log.Println("Server stopped")
}
```

### config/app.yaml

```yaml
app:
  name: gmcore-app
  version: 1.0.0
  environment: dev
  debug: true

server:
  host: localhost
  port: 8080
  read_timeout: 30s
  write_timeout: 30s

paths:
  root: .
  var: var
  cache: var/cache
  log: var/log
  views: resources/views
  migrations: resources/migrations

logging:
  level: info
  format: json
```

### config/routes.yaml

```yaml
routes:
  health:
    path: /health
    handler: HealthController.Check
    methods: [GET]

  api:
    path: /api/v1
    routes:
      users:
        path: /users
        handler: UserController.List
        methods: [GET]
      user:
        path: /users/{id}
        handler: UserController.Get
        methods: [GET]
```

### config/services.yaml

```yaml
services:
  defaults:
    autowire: true
    public: false

  controllers:
    resource: internal/controller/
    tags: [controller]

  services:
    App\Service\UserService:
      arguments:
        $repository: '@App\Repository\UserRepository'

    App\Repository\UserRepository:
      arguments:
        $db: '@database'
```

### internal/kernel/kernel.go

```go
package kernel

import (
    "net/http"

    "github.com/gmcorenet/framework/router"
)

type Kernel struct {
    router *router.Router
}

func NewKernel() *Kernel {
    k := &Kernel{
        router: router.New(),
    }
    k.registerRoutes()
    return k
}

func (k *Kernel) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    k.router.ServeHTTP(w, r)
}

func (k *Kernel) registerRoutes() {
    k.router.GET("/health", healthHandler)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.Write([]byte(`{"status":"ok"}`))
}
```

### internal/controller/health.go

```go
package controller

import (
    "encoding/json"
    "net/http"
)

type HealthController struct{}

func NewHealthController() *HealthController {
    return &HealthController{}
}

type HealthResponse struct {
    Status string `json:"status"`
}

func (h *HealthController) Check(w http.ResponseWriter, r *http.Request) {
    response := HealthResponse{Status: "ok"}
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
```

### internal/model/entity.go

```go
package model

import "time"

type Entity struct {
    ID        string    `json:"id"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

### go.mod

```go
module github.com/gmcorenet/app

go 1.21

require (
    github.com/gmcorenet/framework v1.0.0
    github.com/gmcorenet/sdk/gmcore-config v1.0.0
    github.com/gmcorenet/sdk/gmcore-router v1.0.0
    github.com/gmcorenet/sdk/gmcore-log v1.0.0
)
```

### var/.gitkeep

Create .gitkeep files to ensure directories exist:

```
var/cache/.gitkeep
var/log/.gitkeep
```

## What Gets Copied

Only these files/directories are copied to new app:

```
bin/
cmd/
config/
internal/
public/
resources/
tests/
go.mod
go.sum
README.md
```

## What Gets Transformed

### go.mod

- Module name changes from `github.com/gmcorenet/app` to `github.com/user/app`
- Replace directives get updated paths

### config/app.yaml

- App name gets replaced

### README.md

- App name in title

## Implementation Order

1. Create directory structure
2. Create go.mod template
3. Create cmd/server/main.go
4. Create internal/kernel/kernel.go
5. Create config files
6. Create basic controller/service
7. Create bin/console
8. Create var/.gitkeep files
9. Create README.md template
10. Test with `gmcore create testapp`

## Integration with gmcore create

The `gmcore create` command should:

1. Clone skeleton to temp dir
2. Copy contents (excluding .skeleton/)
3. Transform variables:
   - `{{app_name}}` → user input
   - `{{module_path}}` → derived from app name
4. Run `go mod tidy`
5. Initialize git if requested