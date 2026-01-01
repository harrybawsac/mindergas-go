package csvreader

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestReadAll_Success(t *testing.T) {
	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "readings.csv")

	content := `timestamp,value
2025-12-29T00:00:00,12345.678
2025-12-30T00:00:00,12346.123
2025-12-31T00:00:00,12347.456
`
	if err := os.WriteFile(csvPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test CSV: %v", err)
	}

	readings, err := ReadAll(csvPath)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	if len(readings) != 3 {
		t.Fatalf("got %d readings, want 3", len(readings))
	}

	if readings[0].Value != 12345.678 {
		t.Errorf("readings[0].Value = %f, want 12345.678", readings[0].Value)
	}
	if readings[1].Value != 12346.123 {
		t.Errorf("readings[1].Value = %f, want 12346.123", readings[1].Value)
	}
	if readings[2].Value != 12347.456 {
		t.Errorf("readings[2].Value = %f, want 12347.456", readings[2].Value)
	}
}

func TestReadAll_FileNotFound(t *testing.T) {
	_, err := ReadAll("/nonexistent/path/file.csv")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestReadAll_NoDataRows(t *testing.T) {
	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "empty.csv")

	content := `timestamp,value
`
	if err := os.WriteFile(csvPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test CSV: %v", err)
	}

	_, err := ReadAll(csvPath)
	if err == nil {
		t.Fatal("expected error for CSV with no data rows, got nil")
	}
	if err.Error() != "csv file has no data rows" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestReadAll_InvalidCSV(t *testing.T) {
	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "invalid.csv")

	content := `timestamp,value
"2025-12-29T00:00:00,12345.678
`
	if err := os.WriteFile(csvPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test CSV: %v", err)
	}

	_, err := ReadAll(csvPath)
	if err == nil {
		t.Fatal("expected error for invalid CSV, got nil")
	}
}

func TestReadAll_SkipsInvalidRows(t *testing.T) {
	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "mixed.csv")

	content := `timestamp,value
2025-12-29T00:00:00,12345.678
invalid-timestamp,12346.123
2025-12-30T00:00:00,not-a-number
2025-12-31T00:00:00,12347.456
`
	if err := os.WriteFile(csvPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test CSV: %v", err)
	}

	readings, err := ReadAll(csvPath)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	if len(readings) != 2 {
		t.Fatalf("got %d readings, want 2", len(readings))
	}

	if readings[0].Value != 12345.678 {
		t.Errorf("readings[0].Value = %f, want 12345.678", readings[0].Value)
	}
	if readings[1].Value != 12347.456 {
		t.Errorf("readings[1].Value = %f, want 12347.456", readings[1].Value)
	}
}

func TestReadAll_SkipsShortRows(t *testing.T) {
	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "short.csv")

	content := `timestamp,value
2025-12-29T00:00:00,12345.678
,
2025-12-31T00:00:00,12347.456
`
	if err := os.WriteFile(csvPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test CSV: %v", err)
	}

	readings, err := ReadAll(csvPath)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	if len(readings) != 2 {
		t.Fatalf("got %d readings, want 2", len(readings))
	}
}

func TestReadAll_NoValidReadings(t *testing.T) {
	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "novalid.csv")

	content := `timestamp,value
invalid,invalid
also-invalid,also-invalid
`
	if err := os.WriteFile(csvPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test CSV: %v", err)
	}

	_, err := ReadAll(csvPath)
	if err == nil {
		t.Fatal("expected error for CSV with no valid readings, got nil")
	}
	if err.Error() != "no valid readings found in csv" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestReadAll_DifferentTimestampFormats(t *testing.T) {
	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "formats.csv")

	content := `timestamp,value
2025-12-29T00:00:00,100.1
2025-12-30 12:34:56,200.2
`
	if err := os.WriteFile(csvPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test CSV: %v", err)
	}

	readings, err := ReadAll(csvPath)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	if len(readings) != 2 {
		t.Fatalf("got %d readings, want 2", len(readings))
	}

	if readings[0].Value != 100.1 {
		t.Errorf("readings[0].Value = %f, want 100.1", readings[0].Value)
	}
	if readings[1].Value != 200.2 {
		t.Errorf("readings[1].Value = %f, want 200.2", readings[1].Value)
	}
}

func TestParseTimestamp_RFC3339(t *testing.T) {
	loc, _ := time.LoadLocation("Europe/Amsterdam")
	ts, err := parseTimestamp("2025-12-31T12:34:56Z", loc)
	if err != nil {
		t.Fatalf("parseTimestamp failed: %v", err)
	}

	if ts.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestParseTimestamp_CustomFormat(t *testing.T) {
	loc, _ := time.LoadLocation("Europe/Amsterdam")
	ts, err := parseTimestamp("2025-12-31T12:34:56", loc)
	if err != nil {
		t.Fatalf("parseTimestamp failed: %v", err)
	}

	if ts.Year() != 2025 || ts.Month() != 12 || ts.Day() != 31 {
		t.Errorf("unexpected date: %v", ts)
	}
}

func TestParseTimestamp_SpaceSeparated(t *testing.T) {
	loc, _ := time.LoadLocation("Europe/Amsterdam")
	ts, err := parseTimestamp("2025-12-31 12:34:56", loc)
	if err != nil {
		t.Fatalf("parseTimestamp failed: %v", err)
	}

	if ts.Year() != 2025 || ts.Month() != 12 || ts.Day() != 31 {
		t.Errorf("unexpected date: %v", ts)
	}
}

func TestParseTimestamp_Invalid(t *testing.T) {
	loc, _ := time.LoadLocation("Europe/Amsterdam")
	_, err := parseTimestamp("not-a-timestamp", loc)
	if err == nil {
		t.Fatal("expected error for invalid timestamp, got nil")
	}
	if err.Error() != "unable to parse timestamp: not-a-timestamp" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestReadAll_TimezoneHandling(t *testing.T) {
	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "tz.csv")

	content := `timestamp,value
2025-12-31T00:00:00,12345.678
`
	if err := os.WriteFile(csvPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test CSV: %v", err)
	}

	readings, err := ReadAll(csvPath)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	if len(readings) != 1 {
		t.Fatalf("got %d readings, want 1", len(readings))
	}

	loc := readings[0].Timestamp.Location()
	if loc.String() != "Europe/Amsterdam" && loc.String() != "UTC" {
		t.Logf("Warning: timezone is %s, expected Europe/Amsterdam or UTC", loc.String())
	}
}

func TestReadAll_EmptyValueHandled(t *testing.T) {
	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "emptyval.csv")

	content := `timestamp,value
2025-12-31T00:00:00,
`
	if err := os.WriteFile(csvPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test CSV: %v", err)
	}

	_, err := ReadAll(csvPath)
	if err == nil {
		t.Fatal("expected error for empty value, got nil")
	}
	if err.Error() != "no valid readings found in csv" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestReadAll_NegativeValue(t *testing.T) {
	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "negative.csv")

	content := `timestamp,value
2025-12-31T00:00:00,-100.5
`
	if err := os.WriteFile(csvPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test CSV: %v", err)
	}

	readings, err := ReadAll(csvPath)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	if len(readings) != 1 {
		t.Fatalf("got %d readings, want 1", len(readings))
	}

	if readings[0].Value != -100.5 {
		t.Errorf("readings[0].Value = %f, want -100.5", readings[0].Value)
	}
}
