package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client wraps the TestLLM Management API.
type Client struct {
	BaseURL    *url.URL
	Token      string
	HTTPClient *http.Client
}

// New creates a new client with the provided base URL and token.
func New(baseURL *url.URL, token string) *Client {
	return &Client{
		BaseURL: baseURL,
		Token:   token,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) doRequest(ctx context.Context, method, path string, requestBody, responseBody any) error {
	fullURL, err := url.JoinPath(c.BaseURL.String(), path)
	if err != nil {
		return fmt.Errorf("build url: %w", err)
	}

	var body io.Reader
	if requestBody != nil {
		payload, err := json.Marshal(requestBody)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		body = bytes.NewReader(payload)
	}

	request, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	if requestBody != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	response, err := c.HTTPClient.Do(request)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer response.Body.Close()

	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return &APIError{
			StatusCode: response.StatusCode,
			Body:       string(responseBytes),
			Method:     method,
			Path:       path,
		}
	}

	if responseBody == nil {
		return nil
	}
	if len(responseBytes) == 0 {
		return fmt.Errorf("empty response body")
	}
	if err := json.Unmarshal(responseBytes, responseBody); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}
