# mindergas-go

mindergas-go is a small command-line utility written in Go that reads gas meter readings from either a Postgres database or a CSV file and delivers them as JSON to the mindergas.nl API. It is intended for automated delivery of meter readings to the mindergas.nl service.

**Data sources:**
- **Database mode**: Connects to a Postgres database and selects the earliest reading for today, then POSTs it once.
- **CSV mode**: Reads all rows from a CSV file and POSTs each reading with a 3-second delay between requests to avoid rate limiting.

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

### Database mode
- Connects to a Postgres database (schema `p1`, table `external_readings`) and queries for meter readings recorded within the current day in the `Europe/Amsterdam` timezone.
- Picks the earliest reading (minimum timestamp) for that day.
- Serializes the reading as JSON and POSTs it once.

### CSV mode
- Reads all rows from a CSV file with format `timestamp,value`.
- Loops over each reading and POSTs them sequentially.
- Waits 3 seconds between each POST to prevent rate limiting.

### Payload format
Both modes send JSON with this shape:

  {
    "date": "2025-10-08T00:00:00",
    "reading": 3578.847
  }

- Sends that JSON in a POST request to `https://www.mindergas.nl/api/meter_readings` with headers `Content-Type: application/json`, `API-VERSION: 1.0`, and `AUTH-TOKEN: <token>`.

## Quick start

### Using a database
1. Copy `config/example.json` to `config/config.json` and edit with your Postgres DSN and API token.
2. Build the binary:

```bash
go build -o bin/mindergas ./cmd/main.go
```

3. Run the CLI (dry-run to see payload without sending):

```bash
./bin/mindergas --config=config/config.json --dry-run
```

Or run normally to POST the payload:

```bash
./bin/mindergas --config=config/config.json
```

### Using a CSV file
1. Create a config file with `csv_path` instead of `db_dsn`:

```json
{
  "csv_path": "path/to/readings.csv",
  "token": "your-mindergas-token"
}
```

2. Prepare a CSV file with header row and format `timestamp,value`:

```csv
timestamp,value
2025-12-29T00:00:00,12345.678
2025-12-30T00:00:00,12346.123
2025-12-31T00:00:00,12347.456
```

3. Run the CLI:

```bash
./bin/mindergas --config=config/config_csv.json --dry-run
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

The CLI reads a JSON config file (path via `--config`, default `config/example.json`). You must specify either `db_dsn` OR `csv_path` (not both). The `token` is always required.

### Database configuration

```json
{
  "db_dsn": "host=127.0.0.1 port=5432 user=p1 password='password' dbname=postgres sslmode=disable options='-c search_path=p1'",
  "token": "token from mindergas.nl"
}
```

- `db_dsn`: Postgres DSN used to open a `pgxpool` connection. The `db` package expects the `p1` schema and a table named `external_readings` with columns `created_at` (timestamptz) and `value`.

### CSV configuration

```json
{
  "csv_path": "path/to/readings.csv",
  "token": "token from mindergas.nl"
}
```

- `csv_path`: Path to a CSV file containing meter readings.

### CSV file format

The CSV file must have a header row and two columns:
- `timestamp`: Date/time in RFC3339 or `YYYY-MM-DDTHH:MM:SS` format
- `value`: Meter reading as a decimal number

Example:
```csv
timestamp,value
2025-12-29T00:00:00,12345.678
2025-12-30T00:00:00,12346.123
2025-12-31T00:00:00,12347.456
```

You can set a custom config path using `--config` flag.

## Usage

CLI flags (implemented in `cmd/main.go`):

- `--config` (default `config/example.json`) — path to JSON config.
- `--dry-run` — when set, build and print the JSON payload(s) to stdout but do not POST.

### Database mode examples

```bash
# dry-run mode prints the payload
go run ./cmd --config=config/config.json --dry-run

# send to mindergas.nl
go run ./cmd --config=config/config.json
```

### CSV mode examples

```bash
# dry-run mode prints all payloads (with 3s delay between each)
go run ./cmd --config=config/config_csv.json --dry-run

# send all readings to mindergas.nl (with 3s delay between each POST)
go run ./cmd --config=config/config_csv.json
```

In CSV mode, the tool will:
1. Read all rows from the CSV file
2. POST each reading sequentially
3. Wait 3 seconds between each POST to avoid rate limiting
4. Continue to the next reading even if one fails

## How it works (high-level)

### Database mode
1. `main` loads config and validates presence of `db_dsn` and `token`.
2. It connects to Postgres using `internal/db.Connect` which returns a `Conn` wrapper over `pgxpool.Pool`.
3. `internal/db.SelectEarliestToday` queries for the earliest `created_at` between the local day's start (midnight) and next midnight in `Europe/Amsterdam` timezone.
4. Construct a `models.MeterReading` payload (`date` formatted as `2006-01-02T15:04:05`, `reading` as float64).
5. If `--dry-run` is set, print the payload and exit.
6. Otherwise, POST the JSON to `https://www.mindergas.nl/api/meter_readings`.

### CSV mode
1. `main` loads config and validates presence of `csv_path` and `token`.
2. `internal/csvreader.ReadAll` reads all rows from the CSV file.
3. For each reading:
   - Construct a `models.MeterReading` payload.
   - If `--dry-run`, print the payload; otherwise POST it.
   - Wait 3 seconds before processing the next reading.

## Project structure

- `cmd/main.go` — CLI entry point.
- `internal/db/` — DB connection and query helpers.
- `internal/csvreader/` — CSV file reader for meter readings.
- `internal/httpclient/` — HTTP client wrapper used to POST JSON with retries.
- `pkg/models/` — data models (MeterReading struct).
- `config/` — example config templates and sample CSV.
- `build-all.sh` — helper to cross-compile binaries for multiple platforms.

## Key implementation details

- Timezone: the code uses `Europe/Amsterdam` when computing the day's boundaries. If `time.LoadLocation` fails it falls back to UTC.
- DB access: uses `github.com/jackc/pgx/v5/pgxpool` for connection pooling. `SelectEarliestToday` expects the schema/table `p1.external_readings`.
- CSV parsing: uses Go's `encoding/csv` package. Supports multiple timestamp formats (RFC3339, `YYYY-MM-DDTHH:MM:SS`, `YYYY-MM-DD HH:MM:SS`).
- Rate limiting: CSV mode waits 3 seconds between POSTs to avoid hitting mindergas.nl rate limits.
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
