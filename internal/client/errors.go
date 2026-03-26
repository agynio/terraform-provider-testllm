package client

import (
	"errors"
	"fmt"
	"net/http"
)

// APIError represents a non-2xx response from the TestLLM API.
type APIError struct {
	StatusCode int
	Body       string
	Method     string
	Path       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("testllm api error: %s %s returned %d: %s", e.Method, e.Path, e.StatusCode, e.Body)
}

func (e *APIError) IsNotFound() bool {
	return e.StatusCode == http.StatusNotFound
}

// IsNotFoundError reports whether err is an APIError with status 404.
func IsNotFoundError(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.IsNotFound()
	}
	return false
}
