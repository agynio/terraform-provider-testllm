package client

import (
	"bytes"
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
	Text       string `json:"text"`
	AnyContent *bool  `json:"any_content,omitempty"`
}

type anthropicSystemBlocksContent struct {
	Blocks     json.RawMessage `json:"blocks"`
	AnyContent *bool           `json:"any_content,omitempty"`
}

type AnthropicSystemContent struct {
	Text       string
	Blocks     json.RawMessage
	AnyContent bool
}

type anthropicMessageContent struct {
	Role       string          `json:"role"`
	Content    json.RawMessage `json:"content"`
	AnyContent *bool           `json:"any_content,omitempty"`
}

type AnthropicMessageContent struct {
	Role          string
	Content       string
	ContentBlocks json.RawMessage
	AnyContent    bool
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

func NewAnthropicSystemTextItem(text string, anyContent *bool) (TestItem, error) {
	systemPayload := anthropicSystemTextContent{Text: text}
	if anyContent != nil && *anyContent {
		systemPayload.AnyContent = anyContent
	}
	payload, err := json.Marshal(systemPayload)
	if err != nil {
		return TestItem{}, err
	}
	return TestItem{Type: "anthropic_system", Content: payload}, nil
}

func NewAnthropicSystemBlocksItem(blocksJSON json.RawMessage, anyContent *bool) (TestItem, error) {
	systemPayload := anthropicSystemBlocksContent{Blocks: blocksJSON}
	if anyContent != nil && *anyContent {
		systemPayload.AnyContent = anyContent
	}
	payload, err := json.Marshal(systemPayload)
	if err != nil {
		return TestItem{}, err
	}
	return TestItem{Type: "anthropic_system", Content: payload}, nil
}

func NewAnthropicMessageStringItem(role, content string, anyContent *bool) (TestItem, error) {
	contentPayload, err := json.Marshal(content)
	if err != nil {
		return TestItem{}, err
	}
	messagePayload := anthropicMessageContent{Role: role, Content: json.RawMessage(contentPayload)}
	if anyContent != nil && *anyContent {
		messagePayload.AnyContent = anyContent
	}
	payload, err := json.Marshal(messagePayload)
	if err != nil {
		return TestItem{}, err
	}
	return TestItem{Type: "anthropic_message", Content: payload}, nil
}

func NewAnthropicMessageBlocksItem(role string, blocksJSON json.RawMessage, anyContent *bool) (TestItem, error) {
	messagePayload := anthropicMessageContent{Role: role, Content: blocksJSON}
	if anyContent != nil && *anyContent {
		messagePayload.AnyContent = anyContent
	}
	payload, err := json.Marshal(messagePayload)
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
	anyContent := false
	if anyContentPayload, hasAnyContent := payload["any_content"]; hasAnyContent {
		if err := json.Unmarshal(anyContentPayload, &anyContent); err != nil {
			return AnthropicSystemContent{}, err
		}
	}
	if hasText && hasBlocks {
		return AnthropicSystemContent{}, fmt.Errorf("anthropic_system content must include either text or blocks")
	}
	if hasText {
		var text string
		if err := json.Unmarshal(textPayload, &text); err != nil {
			return AnthropicSystemContent{}, err
		}
		return AnthropicSystemContent{Text: text, AnyContent: anyContent}, nil
	}
	if hasBlocks {
		return AnthropicSystemContent{Blocks: blocksPayload, AnyContent: anyContent}, nil
	}
	return AnthropicSystemContent{}, fmt.Errorf("anthropic_system content missing text or blocks")
}

func ParseAnthropicMessageContent(item TestItem) (AnthropicMessageContent, error) {
	var payload anthropicMessageContent
	if err := json.Unmarshal(item.Content, &payload); err != nil {
		return AnthropicMessageContent{}, err
	}

	trimmed := bytes.TrimLeft(payload.Content, " \t\r\n")
	if len(trimmed) == 0 {
		return AnthropicMessageContent{}, fmt.Errorf("anthropic_message content must be a JSON string or array")
	}

	switch trimmed[0] {
	case '"':
		var textContent string
		if err := json.Unmarshal(payload.Content, &textContent); err != nil {
			return AnthropicMessageContent{}, fmt.Errorf("anthropic_message string content: %w", err)
		}
		return AnthropicMessageContent{Role: payload.Role, Content: textContent, AnyContent: boolFromPointer(payload.AnyContent)}, nil
	case '[':
		var blocks []json.RawMessage
		if err := json.Unmarshal(payload.Content, &blocks); err != nil {
			return AnthropicMessageContent{}, fmt.Errorf("anthropic_message content must be a JSON string or array: %w", err)
		}
		return AnthropicMessageContent{Role: payload.Role, ContentBlocks: payload.Content, AnyContent: boolFromPointer(payload.AnyContent)}, nil
	default:
		return AnthropicMessageContent{}, fmt.Errorf("anthropic_message content must be a JSON string or array")
	}
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
