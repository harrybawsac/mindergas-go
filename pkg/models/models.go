package models

type MeterReading struct {
	ID        string         `json:"id"`
	Timestamp string         `json:"timestamp"`
	Value     float64        `json:"value"`
	Metadata  map[string]any `json:"metadata"`
}
