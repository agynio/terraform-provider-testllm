package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Test represents a test in TestLLM.
type Test struct {
	ID          string     `json:"id"`
	TestSuiteID string     `json:"test_suite_id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Items       []TestItem `json:"items"`
	CreatedAt   string     `json:"created_at"`
	UpdatedAt   string     `json:"updated_at"`
}

// TestItem represents a polymorphic test item payload.
type TestItem struct {
	Type    string          `json:"type"`
	Content json.RawMessage `json:"content"`
}

type testCreateRequest struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Items       []TestItem `json:"items"`
}

type testUpdateRequest struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Items       []TestItem `json:"items"`
}

type messageContent struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type functionCallContent struct {
	CallID    string `json:"call_id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type functionCallOutputContent struct {
	CallID string `json:"call_id"`
	Output string `json:"output"`
}

func NewMessageItem(role, content string) (TestItem, error) {
	payload, err := json.Marshal(messageContent{Role: role, Content: content})
	if err != nil {
		return TestItem{}, err
	}
	return TestItem{Type: "message", Content: payload}, nil
}

func NewFunctionCallItem(callID, name, arguments string) (TestItem, error) {
	payload, err := json.Marshal(functionCallContent{CallID: callID, Name: name, Arguments: arguments})
	if err != nil {
		return TestItem{}, err
	}
	return TestItem{Type: "function_call", Content: payload}, nil
}

func NewFunctionCallOutputItem(callID, output string) (TestItem, error) {
	payload, err := json.Marshal(functionCallOutputContent{CallID: callID, Output: output})
	if err != nil {
		return TestItem{}, err
	}
	return TestItem{Type: "function_call_output", Content: payload}, nil
}

func ParseMessageContent(item TestItem) (string, string, error) {
	var payload messageContent
	if err := json.Unmarshal(item.Content, &payload); err != nil {
		return "", "", err
	}
	return payload.Role, payload.Content, nil
}

func ParseFunctionCallContent(item TestItem) (string, string, string, error) {
	var payload functionCallContent
	if err := json.Unmarshal(item.Content, &payload); err != nil {
		return "", "", "", err
	}
	return payload.CallID, payload.Name, payload.Arguments, nil
}

func ParseFunctionCallOutputContent(item TestItem) (string, string, error) {
	var payload functionCallOutputContent
	if err := json.Unmarshal(item.Content, &payload); err != nil {
		return "", "", err
	}
	return payload.CallID, payload.Output, nil
}

func (c *Client) CreateTest(ctx context.Context, orgID, suiteID, name, description string, items []TestItem) (*Test, error) {
	request := testCreateRequest{
		Name:        name,
		Description: description,
		Items:       items,
	}
	var response Test
	path := fmt.Sprintf("/api/orgs/%s/suites/%s/tests", orgID, suiteID)
	if err := c.doRequest(ctx, http.MethodPost, path, request, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *Client) GetTest(ctx context.Context, orgID, suiteID, testID string) (*Test, error) {
	var response Test
	path := fmt.Sprintf("/api/orgs/%s/suites/%s/tests/%s", orgID, suiteID, testID)
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *Client) UpdateTest(ctx context.Context, orgID, suiteID, testID, name, description string, items []TestItem) (*Test, error) {
	request := testUpdateRequest{
		Name:        name,
		Description: description,
		Items:       items,
	}
	var response Test
	path := fmt.Sprintf("/api/orgs/%s/suites/%s/tests/%s", orgID, suiteID, testID)
	if err := c.doRequest(ctx, http.MethodPatch, path, request, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *Client) DeleteTest(ctx context.Context, orgID, suiteID, testID string) error {
	path := fmt.Sprintf("/api/orgs/%s/suites/%s/tests/%s", orgID, suiteID, testID)
	return c.doRequest(ctx, http.MethodDelete, path, nil, nil)
}
