package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/example/mindergas/internal/db"
	"github.com/example/mindergas/internal/httpclient"
	"github.com/example/mindergas/pkg/models"
)

func main() {
	// Flags
	var configPath string
	var dryRun bool

	flag.StringVar(&configPath, "config", "config/example.json", "Path to JSON config file")
	flag.BoolVar(&dryRun, "dry-run", false, "Build payload but do not POST")
	flag.Parse()

	logger := log.New(os.Stderr, "mindergas-go: ", log.LstdFlags)

	// Load config
	cfgBytes, err := os.ReadFile(configPath)
	if err != nil {
		logger.Fatalf("read config: %v", err)
	}
	var cfg struct {
		DbDSN string `json:"db_dsn"`
	}
	if err := json.Unmarshal(cfgBytes, &cfg); err != nil {
		logger.Fatalf("parse config: %v", err)
	}
	dbURL := cfg.DbDSN
	if dbURL == "" {
		logger.Fatal("db_dsn missing in config")
	}

	ctx := context.Background()

	// Connect to DB (stub or real depending on internal/db implementation)
	conn, err := db.Connect(ctx, dbURL)
	if err != nil {
		logger.Fatalf("db connect: %v", err)
	}
	defer conn.Close(ctx)

	// Select earliest reading for today (Europe/Amsterdam)
	r, err := db.SelectEarliestToday(ctx, conn)
	if err != nil {
		logger.Fatalf("select earliest: %v", err)
	}

	// Normalize timestamp to start of day (midnight) in Europe/Amsterdam
	loc, err := time.LoadLocation("Europe/Amsterdam")
	if err != nil {
		loc = time.UTC
	}
	rt := r.Timestamp.In(loc)
	y, m, d := rt.Date()
	midnight := time.Date(y, m, d, 0, 0, 0, 0, loc)

	payload := models.MeterReading{
		Date:    midnight.Format("2006-01-02T15:04:05"),
		Reading: r.Value,
	}

	b, _ := json.MarshalIndent(payload, "", "  ")

	logger.Printf("selected: date=%s reading=%v", payload.Date, payload.Reading)

	if dryRun {
		fmt.Println(string(b))
		return
	}

	// Hardcoded post URL per request
	postURL := "https://mindergas.nl/api/meter_readings"

	// Use a small default retry count. Retries flag removed per spec.
	client := httpclient.New(postURL, 3)
	if err := client.PostJSON(ctx, b); err != nil {
		logger.Fatalf("post failed: %v", err)
	}

	logger.Printf("delivered payload to %s", postURL)
}
