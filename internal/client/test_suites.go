package client

import (
	"context"
	"fmt"
	"net/http"
)

// TestSuite represents a test suite in TestLLM.
type TestSuite struct {
	ID          string `json:"id"`
	OrgID       string `json:"org_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type testSuiteCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type testSuiteUpdateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (c *Client) CreateTestSuite(ctx context.Context, orgID, name, description string) (*TestSuite, error) {
	request := testSuiteCreateRequest{
		Name:        name,
		Description: description,
	}
	var response TestSuite
	path := fmt.Sprintf("/api/orgs/%s/suites", orgID)
	if err := c.doRequest(ctx, http.MethodPost, path, request, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *Client) GetTestSuite(ctx context.Context, orgID, suiteID string) (*TestSuite, error) {
	var response TestSuite
	path := fmt.Sprintf("/api/orgs/%s/suites/%s", orgID, suiteID)
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *Client) UpdateTestSuite(ctx context.Context, orgID, suiteID, name, description string) (*TestSuite, error) {
	request := testSuiteUpdateRequest{
		Name:        name,
		Description: description,
	}
	var response TestSuite
	path := fmt.Sprintf("/api/orgs/%s/suites/%s", orgID, suiteID)
	if err := c.doRequest(ctx, http.MethodPatch, path, request, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *Client) DeleteTestSuite(ctx context.Context, orgID, suiteID string) error {
	path := fmt.Sprintf("/api/orgs/%s/suites/%s", orgID, suiteID)
	return c.doRequest(ctx, http.MethodDelete, path, nil, nil)
}
