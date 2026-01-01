package db

import (
	"context"
	"testing"
	"time"
)

func TestConn_CloseNilConn(t *testing.T) {
	var c *Conn
	c.Close(context.Background())
}

func TestConn_CloseNilPool(t *testing.T) {
	c := &Conn{pool: nil}
	c.Close(context.Background())
}

func TestSelectEarliestToday_NilConn(t *testing.T) {
	var c *Conn
	_, err := SelectEarliestToday(context.Background(), c)

	if err == nil {
		t.Fatal("expected error for nil connection, got nil")
	}
	if err.Error() != "no db connection" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSelectEarliestToday_NilPool(t *testing.T) {
	c := &Conn{pool: nil}
	_, err := SelectEarliestToday(context.Background(), c)

	if err == nil {
		t.Fatal("expected error for nil pool, got nil")
	}
	if err.Error() != "no db connection" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestConnect_InvalidDSN(t *testing.T) {
	ctx := context.Background()

	invalidDSNs := []string{
		"",
		"invalid://url",
		"host=nonexistent port=99999 user=nobody dbname=doesnotexist",
	}

	for _, dsn := range invalidDSNs {
		t.Run(dsn, func(t *testing.T) {
			conn, err := Connect(ctx, dsn)

			if dsn == "" {
				if err != nil {
					return
				}
				if conn != nil {
					conn.Close(ctx)
				}
				return
			}

			if conn != nil {
				conn.Close(ctx)
			}
		})
	}
}

func TestReading_Struct(t *testing.T) {
	ts := time.Date(2025, 12, 31, 12, 34, 56, 0, time.UTC)
	r := Reading{
		Timestamp: ts,
		Value:     123.45,
	}

	if r.Value != 123.45 {
		t.Errorf("Value = %f, want 123.45", r.Value)
	}

	if r.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}
}
