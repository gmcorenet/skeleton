# GMCore Skeleton

App template for GMCore applications.

## Structure

```
.
├── cmd/
│   └── server/
│       └── main.go          # Application entry point
├── internal/
│   ├── config/              # Configuration loading
│   ├── handler/             # HTTP handlers
│   ├── middleware/          # HTTP middleware
│   ├── model/               # Domain models
│   ├── repository/          # Data access layer
│   └── service/             # Business logic
├── resources/
│   └── config/
│       └── app.yaml         # App configuration
├── go.mod
```

## Usage

This template is used by `gmcore create` to scaffold new applications.

## Requirements

- Go 1.21+
- GMCore Framework v1.0.0