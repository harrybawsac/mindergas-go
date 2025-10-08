package httpclient

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	retryable "github.com/hashicorp/go-retryablehttp"
)

type Client struct {
	url    string
	client *retryable.Client
}

func New(url string) *Client {
	rc := retryable.NewClient()
	rc.HTTPClient = &http.Client{Timeout: 10 * time.Second}
	rc.Logger = nil
	return &Client{url: url, client: rc}
}

func (c *Client) PostJSON(ctx context.Context, body []byte, authToken string) error {
	if c.url == "" {
		return errors.New("no url provided")
	}

	// Build a real *http.Request so headers are set on the actual request
	req, err := http.NewRequestWithContext(ctx, "POST", c.url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("API-VERSION", "1.0")
	req.Header.Set("AUTH-TOKEN", authToken)

	resp, err := c.client.StandardClient().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return errors.New("non-2xx response: " + resp.Status + ". Response body: " + string(bodyBytes) + ". My request body: " + string(body))
	}
	return nil
}
