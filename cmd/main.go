package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/example/mindergas/internal/csvreader"
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
		DbDSN   string `json:"db_dsn"`
		CSVPath string `json:"csv_path"`
		Token   string `json:"token"`
	}
	if err := json.Unmarshal(cfgBytes, &cfg); err != nil {
		logger.Fatalf("parse config: %v", err)
	}
	if cfg.Token == "" {
		logger.Fatal("token missing in config")
	}
	if cfg.DbDSN == "" && cfg.CSVPath == "" {
		logger.Fatal("either db_dsn or csv_path must be specified in config")
	}
	if cfg.DbDSN != "" && cfg.CSVPath != "" {
		logger.Fatal("only one of db_dsn or csv_path can be specified in config")
	}
	token := cfg.Token

	ctx := context.Background()

	postURL := "https://www.mindergas.nl/api/meter_readings"
	client := httpclient.New(postURL)

	loc, err := time.LoadLocation("Europe/Amsterdam")
	if err != nil {
		loc = time.UTC
	}

	if cfg.DbDSN != "" {
		// DB mode: get earliest reading for today and POST once
		conn, err := db.Connect(ctx, cfg.DbDSN)
		if err != nil {
			logger.Fatalf("db connect: %v", err)
		}
		defer conn.Close(ctx)

		r, err := db.SelectEarliestToday(ctx, conn)
		if err != nil {
			logger.Fatalf("select earliest from db: %v", err)
		}

		// Normalize timestamp to start of day (midnight)
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

		if err := client.PostJSON(ctx, b, token); err != nil {
			logger.Fatalf("post failed: %v", err)
		}
		logger.Printf("delivered payload to %s", postURL)
	} else {
		// CSV mode: read all rows and POST each with 3s delay
		readings, err := csvreader.ReadAll(cfg.CSVPath)
		if err != nil {
			logger.Fatalf("read csv: %v", err)
		}

		logger.Printf("found %d readings in CSV", len(readings))

		for i, r := range readings {
			// Normalize timestamp to start of day (midnight)
			rt := r.Timestamp.In(loc)
			y, m, d := rt.Date()
			midnight := time.Date(y, m, d, 0, 0, 0, 0, loc)

			payload := models.MeterReading{
				Date:    midnight.Format("2006-01-02T15:04:05"),
				Reading: r.Value,
			}

			b, _ := json.MarshalIndent(payload, "", "  ")
			logger.Printf("[%d/%d] date=%s reading=%v", i+1, len(readings), payload.Date, payload.Reading)

			if dryRun {
				fmt.Println(string(b))
			} else {
				if err := client.PostJSON(ctx, b, token); err != nil {
					logger.Printf("post failed for %s: %v", payload.Date, err)
					continue
				}
				logger.Printf("delivered payload to %s", postURL)
			}

			// Wait 3 seconds between POSTs (except after the last one)
			if i < len(readings)-1 {
				time.Sleep(3 * time.Second)
			}
		}
	}
}
