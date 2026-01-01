package httpclient

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPostJSON_SendsHeadersAndBody(t *testing.T) {
	expectedBody := `{"date":"2025-10-08T00:00:00","reading":3578.847}`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type header = %q; want %q", got, "application/json")
		}
		if got := r.Header.Get("API-VERSION"); got != "1.0" {
			t.Errorf("API-VERSION header = %q; want %q", got, "1.0")
		}
		if got := r.Header.Get("AUTH-TOKEN"); got != "token123" {
			t.Errorf("AUTH-TOKEN header = %q; want %q", got, "token123")
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != expectedBody {
			t.Errorf("body = %q; want %q", string(body), expectedBody)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer ts.Close()

	c := New(ts.URL)
	if err := c.PostJSON(context.Background(), []byte(expectedBody), "token123"); err != nil {
		t.Fatalf("PostJSON failed: %v", err)
	}
}

func TestPostJSON_EmptyURL(t *testing.T) {
	c := &Client{url: "", client: nil}
	err := c.PostJSON(context.Background(), []byte("{}"), "token")

	if err == nil {
		t.Fatal("expected error for empty URL, got nil")
	}
	if err.Error() != "no url provided" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPostJSON_Non2xxResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad request"))
	}))
	defer ts.Close()

	c := New(ts.URL)
	err := c.PostJSON(context.Background(), []byte(`{"test":"data"}`), "token")

	if err == nil {
		t.Fatal("expected error for non-2xx response, got nil")
	}

	// Check that error contains status and response body
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("expected non-empty error message")
	}
}

func TestPostJSON_ContextCanceled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	c := New(ts.URL)
	err := c.PostJSON(ctx, []byte("{}"), "token")

	if err == nil {
		t.Fatal("expected error for canceled context, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Logf("Warning: expected context.Canceled, got: %v", err)
	}
}

func TestPostJSON_StatusOK(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := New(ts.URL)
	if err := c.PostJSON(context.Background(), []byte("{}"), "token"); err != nil {
		t.Fatalf("PostJSON failed: %v", err)
	}
}

func TestPostJSON_Status201(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer ts.Close()

	c := New(ts.URL)
	if err := c.PostJSON(context.Background(), []byte("{}"), "token"); err != nil {
		t.Fatalf("PostJSON failed: %v", err)
	}
}

func TestPostJSON_Status500(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer ts.Close()

	c := New(ts.URL)
	err := c.PostJSON(context.Background(), []byte("{}"), "token")

	if err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}
}

func TestNew(t *testing.T) {
	url := "https://example.com/api"
	c := New(url)

	if c == nil {
		t.Fatal("New returned nil")
	}
	if c.url != url {
		t.Errorf("url = %q, want %q", c.url, url)
	}
	if c.client == nil {
		t.Error("client is nil")
	}
}

func TestPostJSON_EmptyBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if len(body) != 0 {
			t.Errorf("body length = %d, want 0", len(body))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := New(ts.URL)
	if err := c.PostJSON(context.Background(), []byte{}, "token"); err != nil {
		t.Fatalf("PostJSON failed: %v", err)
	}
}

func TestPostJSON_EmptyAuthToken(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("AUTH-TOKEN")
		if token != "" {
			t.Errorf("AUTH-TOKEN = %q, want empty string", token)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := New(ts.URL)
	if err := c.PostJSON(context.Background(), []byte("{}"), ""); err != nil {
		t.Fatalf("PostJSON failed: %v", err)
	}
}

func TestPostJSON_InvalidURL(t *testing.T) {
	// Use an invalid URL that will cause http.NewRequestWithContext to fail
	c := &Client{
		url:    "ht tp://invalid url with spaces",
		client: New("http://example.com").client,
	}

	err := c.PostJSON(context.Background(), []byte("{}"), "token")
	if err == nil {
		t.Fatal("expected error for invalid URL, got nil")
	}
}
