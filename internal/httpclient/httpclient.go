package httpclient

import (
	"context"
	"errors"
	"net/http"
	"time"

	retryable "github.com/hashicorp/go-retryablehttp"
)

type Client struct {
	url     string
	retries int
	client  *retryable.Client
}

func New(url string, retries int) *Client {
	rc := retryable.NewClient()
	rc.HTTPClient = &http.Client{Timeout: 10 * time.Second}
	rc.RetryMax = retries
	rc.Logger = nil
	return &Client{url: url, retries: retries, client: rc}
}

func (c *Client) PostJSON(ctx context.Context, body []byte) error {
	if c.url == "" {
		return errors.New("no url provided")
	}
	req, err := retryable.NewRequest("POST", c.url, body)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.StandardClient().Do(req.Request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.New("non-2xx response")
	}
	return nil
}
