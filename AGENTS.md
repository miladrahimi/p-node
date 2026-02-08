# AGENTS.md

## Project summary
P-Node is a lightweight Xray node managed by P-Manager written in Go.
It runs Xray binary, exposes a small HTTP API for stats/config updates, and persists settings in a JSON file store.

## Project Structure
- `cmd`: Cobra commands
- `configs`: Configuration files
- `internal/app`: App lifecycle orchestration
- `internal/config`: Configuration loader, constants, and paths
- `internal/coordinator`: Periodic sync between current state, local Xray process, and P-Manager
- `internal/data`: database schema, models and default values
- `internal/http`: Echo HTTP server, routes, and API handlers
- `pkg/database`: Simple generic file-backed JSON store
- `pkg/http/client`: HTTP client
- `pkg/http/middleware`: Shared middleware used in this project and P-Manager
- `pkg/http/validator`: Echo validator implementation based on `go-playground/validator`
- `pkg/logger`: A logging wrapper around `zap` logger
- `pkg/util`: Generic utils
- `pkg/worker`: Simple worker implementation for periodic tasks
- `pkg/xray`: Xray wrapper (Xray process lifecycle, config model, gRPC stats)
- `scripts`: Scripts for project and server setup
- `storage`: Application data storage directory
- `third_party`: Third-party binaries and libraries

## Runtime Flow
- `main.go` → `cmd/root.go` (Cobra) → `cmd/serve.go`.
- `internal/app.New()`: Builds config, logger, database, xray, http server, http client, coordinator, and other modules.
- `App.Run()`: Initializes and runs long-running modules like database, xray, coordinator, and http server, etc.
- Signals handled in `internal/app/app.go` to stop gracefully.

## Database

- Database is simple file-backed JSON store
- Database driver is `pkg/database`
- Database schema is defined in `internal/data/data.go`
- Database directory path is defined in `internal/config/config.go` as `DatabaseDirectory`
- Database file path is `$DatabaseDirectory/data.json`

## HTTP APIs
- All handlers are located in `internal/http/handlers`
- Requests and responses are in JSON format
- Authentication is token-based, token is stored in database (`http_token` in `internal/data/settings.go`)
- Header `Authorization: Bearer <token>` is checked for authenticated routes

## Major Dependencies
- `github.com/xtls/xray-core`: Interaction with running Xray process
- `github.com/labstack/echo`: HTTP server and router
- `github.com/spf13/cobra`: CLI framework
- `github.com/go-playground/validator`: Validating config models
- `github.com/cockroachdb/errors`: Errors and stacktrace
- `go.uber.org/zap`: Logging

## Key Behavior Notes / Gotchas
- `pkg/database` locks Load/Save and creates parent dir on Save.
- `pkg/xray` serializes Run/Stop/Restart; `Reconfigure` validates config; `QueryStats` lazily connects.
- `pkg/http/client` skips TLS verification (intentionally).
- `settings.http_token` is generated randomly on first run (see `internal/data/settings.go`).
- `make update` is destructive (`git reset --hard` + `git clean -fd`).

## Xray
Xray is a proxy platform which can be used to run proxy servers. with different protocols like Shadowsocks, VMess, VLess,
Socks, etc. It supports chain of proxies. The high level flow is:

```
[ Client ] -> [ Xray on Server 1 ] -> [ Xray on Server 2 ] -> Internet
```

Each xray node receives traffic on its inbound port and forwards it to the next node in the chain.
Based on the routing rules and balancers it forwards traffic to the appropriate outbound.

Xray provides a reverse proxy feature that allows a connection to be initiated from the next node back to the previous
node. This is particularly useful when the previous node is behind a firewall (such as the GFW) that restricts outbound
connections to the network where the next node resides.

## External Links
- [Xray: Proxy Platform](https://github.com/XTLS/Xray-core)
- [Xray Config Examples](https://github.com/XTLS/Xray-examples)

## AI Development Guidance
- Keep it simple stupid.
- Prefer small, inline improvements; keep architecture intact.
