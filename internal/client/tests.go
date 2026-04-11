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
	Role       string `json:"role"`
	Content    string `json:"content"`
	AnyRole    *bool  `json:"any_role,omitempty"`
	AnyContent *bool  `json:"any_content,omitempty"`
	Repeat     *bool  `json:"repeat,omitempty"`
}

type MessageContent struct {
	Role       string
	Content    string
	AnyRole    bool
	AnyContent bool
	Repeat     bool
}

type anthropicSystemTextContent struct {
	Text string `json:"text"`
}

type anthropicSystemBlocksContent struct {
	Blocks json.RawMessage `json:"blocks"`
}

type AnthropicSystemContent struct {
	Text   string
	Blocks json.RawMessage
}

type anthropicMessageContent struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

type AnthropicMessageContent struct {
	Role          string
	Content       string
	ContentBlocks json.RawMessage
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

func NewMessageItem(role, content string, anyRole, anyContent, repeat *bool) (TestItem, error) {
	messagePayload := messageContent{Role: role, Content: content}
	if anyRole != nil && *anyRole {
		messagePayload.AnyRole = anyRole
	}
	if anyContent != nil && *anyContent {
		messagePayload.AnyContent = anyContent
	}
	if repeat != nil && *repeat {
		messagePayload.Repeat = repeat
	}

	payload, err := json.Marshal(messagePayload)
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

func NewAnthropicSystemTextItem(text string) (TestItem, error) {
	payload, err := json.Marshal(anthropicSystemTextContent{Text: text})
	if err != nil {
		return TestItem{}, err
	}
	return TestItem{Type: "anthropic_system", Content: payload}, nil
}

func NewAnthropicSystemBlocksItem(blocksJSON json.RawMessage) (TestItem, error) {
	payload, err := json.Marshal(anthropicSystemBlocksContent{Blocks: blocksJSON})
	if err != nil {
		return TestItem{}, err
	}
	return TestItem{Type: "anthropic_system", Content: payload}, nil
}

func NewAnthropicMessageStringItem(role, content string) (TestItem, error) {
	contentPayload, err := json.Marshal(content)
	if err != nil {
		return TestItem{}, err
	}
	payload, err := json.Marshal(anthropicMessageContent{Role: role, Content: json.RawMessage(contentPayload)})
	if err != nil {
		return TestItem{}, err
	}
	return TestItem{Type: "anthropic_message", Content: payload}, nil
}

func NewAnthropicMessageBlocksItem(role string, blocksJSON json.RawMessage) (TestItem, error) {
	payload, err := json.Marshal(anthropicMessageContent{Role: role, Content: blocksJSON})
	if err != nil {
		return TestItem{}, err
	}
	return TestItem{Type: "anthropic_message", Content: payload}, nil
}

func ParseMessageContent(item TestItem) (MessageContent, error) {
	var payload messageContent
	if err := json.Unmarshal(item.Content, &payload); err != nil {
		return MessageContent{}, err
	}
	return MessageContent{
		Role:       payload.Role,
		Content:    payload.Content,
		AnyRole:    boolFromPointer(payload.AnyRole),
		AnyContent: boolFromPointer(payload.AnyContent),
		Repeat:     boolFromPointer(payload.Repeat),
	}, nil
}

func boolFromPointer(value *bool) bool {
	if value == nil {
		return false
	}
	return *value
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

func ParseAnthropicSystemContent(item TestItem) (AnthropicSystemContent, error) {
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(item.Content, &payload); err != nil {
		return AnthropicSystemContent{}, err
	}
	textPayload, hasText := payload["text"]
	blocksPayload, hasBlocks := payload["blocks"]
	if hasText && hasBlocks {
		return AnthropicSystemContent{}, fmt.Errorf("anthropic_system content must include either text or blocks")
	}
	if hasText {
		var text string
		if err := json.Unmarshal(textPayload, &text); err != nil {
			return AnthropicSystemContent{}, err
		}
		return AnthropicSystemContent{Text: text}, nil
	}
	if hasBlocks {
		return AnthropicSystemContent{Blocks: blocksPayload}, nil
	}
	return AnthropicSystemContent{}, fmt.Errorf("anthropic_system content missing text or blocks")
}

func ParseAnthropicMessageContent(item TestItem) (AnthropicMessageContent, error) {
	var payload anthropicMessageContent
	if err := json.Unmarshal(item.Content, &payload); err != nil {
		return AnthropicMessageContent{}, err
	}
	var textContent string
	if err := json.Unmarshal(payload.Content, &textContent); err == nil {
		return AnthropicMessageContent{Role: payload.Role, Content: textContent}, nil
	}
	var blocks []json.RawMessage
	if err := json.Unmarshal(payload.Content, &blocks); err != nil {
		return AnthropicMessageContent{}, err
	}
	return AnthropicMessageContent{Role: payload.Role, ContentBlocks: payload.Content}, nil
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
