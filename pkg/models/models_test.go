package models

import (
	"encoding/json"
	"testing"
)

func TestMeterReading_JSONMarshaling(t *testing.T) {
	mr := MeterReading{
		Date:    "2025-12-31T00:00:00",
		Reading: 12345.678,
	}

	b, err := json.Marshal(mr)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	expected := `{"date":"2025-12-31T00:00:00","reading":12345.678}`
	if string(b) != expected {
		t.Errorf("got %s, want %s", string(b), expected)
	}
}

func TestMeterReading_JSONUnmarshaling(t *testing.T) {
	data := `{"date":"2025-12-31T00:00:00","reading":12345.678}`

	var mr MeterReading
	if err := json.Unmarshal([]byte(data), &mr); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if mr.Date != "2025-12-31T00:00:00" {
		t.Errorf("Date = %s, want %s", mr.Date, "2025-12-31T00:00:00")
	}
	if mr.Reading != 12345.678 {
		t.Errorf("Reading = %f, want %f", mr.Reading, 12345.678)
	}
}

func TestMeterReading_ZeroValues(t *testing.T) {
	var mr MeterReading

	if mr.Date != "" {
		t.Errorf("zero Date = %s, want empty string", mr.Date)
	}
	if mr.Reading != 0 {
		t.Errorf("zero Reading = %f, want 0", mr.Reading)
	}

	b, err := json.Marshal(mr)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	expected := `{"date":"","reading":0}`
	if string(b) != expected {
		t.Errorf("got %s, want %s", string(b), expected)
	}
}
