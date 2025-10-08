package httpclient

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
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
