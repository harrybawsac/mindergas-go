package csvreader

import (
	"encoding/csv"
	"errors"
	"os"
	"strconv"
	"time"
)

// Reading represents a single meter reading from the CSV file.
type Reading struct {
	Timestamp time.Time
	Value     float64
}

// ReadAll reads a CSV file and returns all readings.
//
// Expected CSV format: timestamp,value
// - timestamp: RFC3339 or "2006-01-02 15:04:05" format
// - value: float64 meter reading
//
// The CSV file should have a header row which will be skipped.
func ReadAll(path string) ([]Reading, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return nil, errors.New("csv file has no data rows")
	}

	loc, err := time.LoadLocation("Europe/Amsterdam")
	if err != nil {
		loc = time.UTC
	}

	var readings []Reading

	// Skip header row (index 0)
	for _, record := range records[1:] {
		if len(record) < 2 {
			continue
		}

		ts, err := parseTimestamp(record[0], loc)
		if err != nil {
			continue
		}

		val, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			continue
		}

		readings = append(readings, Reading{
			Timestamp: ts,
			Value:     val,
		})
	}

	if len(readings) == 0 {
		return nil, errors.New("no valid readings found in csv")
	}

	return readings, nil
}

// parseTimestamp tries multiple timestamp formats
func parseTimestamp(s string, loc *time.Location) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z07:00",
	}

	for _, format := range formats {
		if t, err := time.ParseInLocation(format, s, loc); err == nil {
			return t, nil
		}
	}

	return time.Time{}, errors.New("unable to parse timestamp: " + s)
}
