# mindergas-go

mindergas-go is a small command-line utility written in Go that selects the earliest gas meter reading for "today" from a Postgres database and delivers that reading as JSON to a configured HTTP endpoint. It is intended for automated daily delivery of a baseline meter reading to an external API (for example: mindergas.nl).

This README is intentionally detailed and covers project purpose, configuration, usage, internals, testing, debugging, and recommended development workflows.

## Table of contents
- What this project does
- Quick start
- Installation / Build
- Configuration
- Usage
- How it works (high-level)
- Project structure
- Key implementation details
- Tests
- Troubleshooting & debugging
- Contributing
- License & acknowledgements

## What this project does

- Connects to a Postgres database (schema `p1`, table `external_readings`) and queries for meter readings recorded within the current day in the `Europe/Amsterdam` timezone.
- Picks the earliest reading (minimum timestamp) for that day.
- Serializes the reading as JSON with the shape:

  {
    "date": "2025-10-08T00:00:00",
    "reading": 3578.847
  }

- Sends that JSON in a POST request to a configured endpoint with headers `Content-Type: application/json`, `API-VERSION: 1.0`, and `AUTH-TOKEN: <token>`.

## Quick start

1. Copy `config/example.json` to `config/config.json` and edit with your Postgres DSN and API token.
2. Build the binary:

```bash
go build -o bin/mindergas ./cmd/main.go
```

3. Run the CLI (dry-run to see payload without sending):

```bash
./bin/mindergas --config=config/example.json --dry-run
```

Or run normally to POST the payload:

```bash
./bin/mindergas --config=config/example.json
```

## Installation / Build

Prerequisites:
- Go 1.24+ (project uses go.mod set to go 1.24)
- Access to the target Postgres instance (or suitable connection string for testing)

Local build (single platform):

```bash
go build -o bin/mindergas ./cmd/main.go
```

Cross-platform builds are automated in `build-all.sh` which produces multiple platform-targeted binaries under `bin/`.

## Configuration

The CLI reads a JSON config file (path via `--config`, default `config/example.json`) with this structure:

```json
{
  "db_dsn": "host=127.0.0.1 port=5432 user=p1 password='password' dbname=postgres sslmode=disable options='-c search_path=p1'",
  "token": "token from mindergas.nl"
}
```

- `db_dsn`: Postgres DSN used to open a `pgxpool` connection. The `db` package expects the `p1` schema and a table named `external_readings` with columns `id`, `created_at` (timestamptz), and `value`.
- `token`: API token used for `AUTH-TOKEN` header when posting to the target endpoint.

You can set a custom config path using `--config` flag.

## Usage

CLI flags (implemented in `cmd/main.go`):

- `--config` (default `config/example.json`) — path to JSON config.
- `--dry-run` — when set, build and print the JSON payload to stdout but do not POST.

Examples:

```bash
# dry-run mode prints the payload
go run ./cmd --config=config/example.json --dry-run

# send to remote endpoint (make sure config.json has correct token & DB DSN)
go run ./cmd --config=config/example.json
```

## How it works (high-level)

1. `main` loads config and validates presence of `db_dsn` and `token`.
2. It connects to Postgres using `internal/db.Connect` which returns a `Conn` wrapper over `pgxpool.Pool`.
3. `internal/db.SelectEarliestToday` queries for the earliest `created_at` between the local day's start (midnight) and next midnight in `Europe/Amsterdam` timezone.
4. Construct a `models.MeterReading` payload (`date` formatted as `2006-01-02T15:04:05`, `reading` as float64).
5. If `--dry-run` is set, print the payload and exit.
6. Otherwise, create an `httpclient.Client` and call `PostJSON` which POSTs the JSON to `https://www.mindergas.nl/api/meter_readings` (URL is hard-coded in `cmd/main.go`) with headers `Content-Type: application/json`, `API-VERSION: 1.0`, and `AUTH-TOKEN`.

## Project structure

- `cmd/main.go` — CLI entry point.
- `internal/db/` — DB connection and query helpers.
- `internal/httpclient/` — HTTP client wrapper used to POST JSON with retries.
- `pkg/models/` — data models (MeterReading struct).
- `config/` — example config templates.
- `build-all.sh` — helper to cross-compile binaries for multiple platforms.

## Key implementation details

- Timezone: the code uses `Europe/Amsterdam` when computing the day's boundaries. If `time.LoadLocation` fails it falls back to UTC.
- DB access: uses `github.com/jackc/pgx/v5/pgxpool` for connection pooling. `SelectEarliestToday` expects the schema/table `p1.external_readings`.
- HTTP client: uses `github.com/hashicorp/go-retryablehttp` for a retrying client. The `PostJSON` method builds a real `*http.Request` and sets headers on it so they are present on the outgoing request.

## Tests

There is a unit test that validates `PostJSON` sends headers and body correctly using a local httptest server. Run all tests with:

```bash
go test ./...
```

Expected output should indicate `internal/httpclient` tests pass (others may have no tests).

## Troubleshooting & debugging

Common issues and how to resolve them:

- 400 Bad Request complaining about missing `meter_reading` or similar
  - Ensure the code sends `Content-Type: application/json` and the body matches the API shape. See `internal/httpclient.PostJSON` — headers must be set on the concrete `*http.Request`.
  - If you see the server rejecting the payload while a curl request succeeds, capture the raw HTTP request from the Go program (add logging in `PostJSON` to print headers and body before send) and compare to curl.

- DB connection errors
  - Validate `db_dsn` in `config/example.json`. Test connectivity with `psql` or similar.
  - Ensure the config's `options` set the search_path to `p1` if your table lives there.

- Timezone / date mismatches
  - The code computes the start of day in `Europe/Amsterdam`. If your DB stores timestamps in UTC (but with timestamptz), ensure trust in `timestamptz` handling and that the `created_at` values have the correct timezone semantics.

Logging recommendations
- The CLI uses the standard library `log` to stderr for high-level messages. Add logging to `internal/httpclient.PostJSON` and `internal/db.SelectEarliestToday` for debugging.

## Development notes

- To change the POST target URL, update the `postURL` variable in `cmd/main.go`. Consider adding it as a flag (`--post-url`) for configurability.
- For secure handling of secrets, avoid committing tokens to repository or example configs. Use environment variables or a secret manager for production runs.
- If you prefer the retryablehttp abstraction, you can construct `retryable.NewRequest` and then call the retryable client to do retries while ensuring headers are assigned to the underlying `Request` (or use `req.SetBasicAuth` / set headers on the returned `Request.Request`). Current implementation uses `http.NewRequestWithContext` then sends using `c.client.StandardClient().Do(req)` which preserves headers and works reliably with the underlying `http.Client`.

## Contributing

Contributions are welcome. Suggested workflow:

1. Fork the repo and create a feature branch.
2. Add tests for new behavior.
3. Open a PR with a description of the change, rationale, and test results.

Style & dependency notes:
- The project uses Go modules. Keep `go.mod` tidy and run `go mod tidy` after adding dependencies.
- Follow standard gofmt / go vet checks.

## License & acknowledgements

This repository contains example code and is provided without an explicit license file. Add a LICENSE file if you plan to publish or distribute this project.

Acknowledgements:
- `github.com/hashicorp/go-retryablehttp` for retrying HTTP client
- `github.com/jackc/pgx` for Postgres connectivity

---

If you'd like, I can:

- Add a small section with explicit curl examples that mirror what the CLI sends.
- Make the POST URL configurable via a flag.
- Add logging of the outgoing request (headers + body) behind a `--verbose` flag.

Tell me which of the above you'd like next.
