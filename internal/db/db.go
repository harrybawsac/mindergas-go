package db

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Conn is a thin wrapper around pgxpool.Pool to simplify usage in the CLI.
type Conn struct{ pool *pgxpool.Pool }

// Connect creates a pgxpool connection from the DSN.
func Connect(ctx context.Context, url string) (*Conn, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		// Try passing the raw url to New instead of ParseConfig in case the
		// DSN is space-separated libpq style. pgxpool accepts a connection
		// string directly via NewWithConfig, but ParseConfig only supports
		// a connection string too; keep simple: use New.
		pool, err := pgxpool.New(ctx, url)
		if err != nil {
			return nil, err
		}
		return &Conn{pool: pool}, nil
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &Conn{pool: pool}, nil
}

func (c *Conn) Close(ctx context.Context) {
	if c == nil || c.pool == nil {
		return
	}
	c.pool.Close()
}

type Reading struct {
	ID        string
	Timestamp time.Time
	Value     float64
}

// SelectEarliestToday queries p1.external_readings for the earliest created_at
// for the current date in the Europe/Amsterdam timezone. It returns the
// earliest row (id, created_at, value). The table is assumed to be in schema
// p1 and table external_readings with columns id (uuid or text), created_at
// (timestamptz), and value (numeric/float).
func SelectEarliestToday(ctx context.Context, c *Conn) (*Reading, error) {
	if c == nil || c.pool == nil {
		return nil, errors.New("no db connection")
	}

	// Current time in Europe/Amsterdam
	loc, err := time.LoadLocation("Europe/Amsterdam")
	if err != nil {
		loc = time.UTC
	}
	now := time.Now().In(loc)
	// start of day and end of day in that timezone
	y, m, d := now.Date()
	start := time.Date(y, m, d, 0, 0, 0, 0, loc)
	end := start.Add(24 * time.Hour)

	// Query for earliest created_at between start (inclusive) and end (exclusive)
	// Note: use parameterized query to avoid SQL injection.
	const q = `SELECT id, created_at, value FROM p1.external_readings
WHERE created_at >= $1 AND created_at < $2
ORDER BY created_at ASC LIMIT 1`

	row := c.pool.QueryRow(ctx, q, start, end)
	var r Reading
	if err := row.Scan(&r.ID, &r.Timestamp, &r.Value); err != nil {
		return nil, err
	}
	return &r, nil
}
