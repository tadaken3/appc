package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
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
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return nil, fmt.Errorf("API error: status %d: %s", resp.StatusCode, string(body))
	}
	return resp, nil
}

func (c *Client) GetRaw(ctx context.Context, path string) (*http.Response, error) {
	return c.do(ctx, path)
}

func (c *Client) GetRawWithAccept(ctx context.Context, path, accept string) (*http.Response, error) {
	return c.doMethodWithHeaders(ctx, http.MethodGet, path, nil, map[string]string{"Accept": accept})
}

func (c *Client) Post(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	return c.doMethod(ctx, http.MethodPost, path, body)
}

func (c *Client) do(ctx context.Context, path string) (*http.Response, error) {
	return c.doMethod(ctx, http.MethodGet, path, nil)
}

func (c *Client) doMethod(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	return c.doMethodWithHeaders(ctx, method, path, body, nil)
}

func (c *Client) doMethodWithHeaders(ctx context.Context, method, path string, body io.Reader, headers map[string]string) (*http.Response, error) {
	token, err := c.tokens.Token()
	if err != nil {
		return nil, fmt.Errorf("getting token: %w", err)
	}

	// Buffer body for retry support
	var bodyBytes []byte
	if body != nil {
		bodyBytes, err = io.ReadAll(body)
		if err != nil {
			return nil, fmt.Errorf("reading request body: %w", err)
		}
	}

	url := c.BaseURL + path

	for attempt := 0; attempt <= c.MaxRetries; attempt++ {
		var reqBody io.Reader
		if bodyBytes != nil {
			reqBody = bytes.NewReader(bodyBytes)
		}

		req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
		if method == http.MethodPost {
			req.Header.Set("Content-Type", "application/json")
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		resp, err := c.http.Do(req)
		if err != nil {
			return nil, fmt.Errorf("executing request: %w", err)
		}

		if resp.StatusCode == http.StatusTooManyRequests && attempt < c.MaxRetries {
			_ = resp.Body.Close()
			wait := retryAfterDuration(resp.Header.Get("Retry-After"))
			if wait > 0 {
				time.Sleep(wait)
			}
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("API error: status 429 after %d retries", c.MaxRetries)
}

func retryAfterDuration(value string) time.Duration {
	if value == "" {
		return 1 * time.Second
	}
	seconds, err := strconv.Atoi(value)
	if err != nil {
		return 1 * time.Second
	}
	return time.Duration(seconds) * time.Second
}
