# GMCore Skeleton

Template application for GMCore framework.

## Requirements

- Go 1.21+
- PostgreSQL (or other supported database)

## Installation

```bash
# Clone the skeleton
git clone https://github.com/gmcorenet/skeleton myapp
cd myapp

# Install dependencies
go mod tidy

# Copy and configure environment
cp .env .env.local

# Run migrations
./app migrate

# Start development server
./app serve
```

## Directory Structure

```
├── cmd/              # Entry points
├── config/           # Configuration
├── internal/         # Application code
│   ├── controller/   # HTTP handlers
│   ├── service/      # Business logic
│   ├── repository/    # Data access
│   ├── model/         # Domain models
│   └── ...
├── public/           # Static files
├── resources/        # Templates, migrations, translations
├── tests/             # Test files
└── var/              # Runtime files (cache, logs)
```

## Available Commands

```bash
./app serve           # Start web server
./app migrate         # Run database migrations
```

## Configuration

Edit `.env` for environment-specific settings.

Edit `app.yaml` for application metadata.

## License

MIT
