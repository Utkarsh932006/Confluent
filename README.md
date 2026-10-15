# Confluent

A lightweight, high-performance internal knowledge base platform inspired by Confluence, designed specifically for small teams and startups.

## Architecture Highlights

- **Backend**: Go with stdlib `net/http` (Go 1.22+ method & path routing), `database/sql`, structured logging (`slog`), JWT authentication, and role-based access control (`admin`, `editor`, `viewer`).
- **Database**: Embedded SQLite with WAL mode (`PRAGMA journal_mode=WAL`), busy timeouts, foreign key cascading, and built-in SQLite **FTS5** (Full Text Search) with automatic triggers.
- **Frontend**: Next.js with React, Tailwind CSS, `@tailwindcss/typography`, recursive page hierarchy trees, markdown editor with live preview, and fast keyboard-driven search (`Cmd+K`).

## Quick Start

### 1. Run Backend Server (Go)

```bash
# Run server directly
go run ./cmd/server

# Or compile and execute
go build -o ./bin/confluent ./cmd/server
./bin/confluent
```

Default credentials seeded on first startup:
- **Email**: `admin@confluent.local`
- **Password**: `admin123`
- **Server Address**: `http://localhost:8080`

### 2. Run Tests

```bash
go test -v ./...
```

### 3. Run with Docker Compose

```bash
docker compose up --build -d
```
Persistent database storage is preserved under Docker volume `confluent-data`.
