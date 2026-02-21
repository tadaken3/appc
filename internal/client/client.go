package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultBaseURL = "https://api.appstoreconnect.apple.com"

type TokenProvider interface {
	Token() (string, error)
}

type Client struct {
	BaseURL    string
	MaxRetries int
	tokens     TokenProvider
	http       *http.Client
}

func New(tokens TokenProvider) *Client {
	return &Client{
		BaseURL:    defaultBaseURL,
		MaxRetries: 3,
		tokens:     tokens,
		http:       &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) Get(ctx context.Context, path string) (*http.Response, error) {
	resp, err := c.do(ctx, path)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		resp.Body.Close()
		return nil, fmt.Errorf("API error: status %d", resp.StatusCode)
	}
	return resp, nil
}

func (c *Client) GetRaw(ctx context.Context, path string) (*http.Response, error) {
	return c.do(ctx, path)
}

func (c *Client) Post(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	return c.doMethod(ctx, http.MethodPost, path, body)
}

func (c *Client) do(ctx context.Context, path string) (*http.Response, error) {
	return c.doMethod(ctx, http.MethodGet, path, nil)
}

func (c *Client) doMethod(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	token, err := c.tokens.Token()
	if err != nil {
		return nil, fmt.Errorf("getting token: %w", err)
	}

	url := c.BaseURL + path

	for attempt := 0; attempt <= c.MaxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, url, body)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
		if method == http.MethodPost {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.http.Do(req)
		if err != nil {
			return nil, fmt.Errorf("executing request: %w", err)
		}

		if resp.StatusCode == http.StatusTooManyRequests && attempt < c.MaxRetries {
			resp.Body.Close()
			wait := 1 * time.Second
			if ra := resp.Header.Get("Retry-After"); ra == "0" {
				wait = 0
			}
			if wait > 0 {
				time.Sleep(wait)
			}
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("API error: status 429 after %d retries", c.MaxRetries)
}
