# Config

Skeleton config follows a layered, programmatic approach — not Symfony-style YAML
service definitions. Here's how each config concern works.

## app.yaml — Application configuration

The main config file loaded at `cmd/main.go:53` via `loadConfig("app.yaml")`.

```yaml
server:
  host: "0.0.0.0"
  port: "8080"

app:
  name: gmcore-app
  version: 1.0.0
  env: dev
  debug: false
  # ... metadata fields
```

Values are read into an `AppConfig` struct and merged into the `kernel.Config`.
Environment variables (`SERVER_HOST`, `SERVER_PORT`, `APP_ENV`, `APP_DEBUG`) override
the file values — see `cmd/main.go:68-73`.

## Routes — defined in Go code

Routes are **not** declared in a YAML file. Define them programmatically in
`cmd/main.go:setupRoutes()`:

```go
r.GET("/", func(w http.ResponseWriter, req *http.Request, params map[string]string) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Hello"))
})
```

The `gmcore-router` SDK also supports loading routes from YAML via
`router.LoadConfig()`, but the skeleton keeps its own routes as the single
source of truth in `setupRoutes()`. Use whichever approach fits your app.

## Services — registered in Go code

Service registration happens in `cmd/main.go:80` via the kernel's built-in
defaults and the container:

```go
k.RegisterDefaultServices()       // registers logger, env, etc.

c := k.Container()
c.Set("my_service", NewMyService())
```

There is no YAML service compiler. The `container.Set()`, `container.Factory()`,
and `RegisterDefaultServices()` methods are the programmatic equivalent of
Symfony's `services.yaml`.

## config/packages/ — Component-level config

| File               | Purpose                                          |
|--------------------|--------------------------------------------------|
| `framework.yaml`   | Error handling, timezone, secrets (env vars)     |
| `cache.yaml`       | Cache adapter, TTL, prefix                       |
| `database.yaml`    | Database driver, DSN (env var), connection pool  |

These are reference config files. Load them in your app if you use the
corresponding SDK or service. The skeleton does not auto-parse them.

## Environment variables

Sensitive values use environment variables, referenced with `%env(VAR)%` or read
directly via `os.Getenv()`. The `app.yaml` loading in `cmd/main.go` reads them at
startup (lines 68-73). Never commit secrets to config files.
