package models

type MeterReading struct {
	Date    string  `json:"date"`
	Reading float64 `json:"reading"`
}
