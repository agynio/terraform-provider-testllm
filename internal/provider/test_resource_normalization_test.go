package provider

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/agynio/terraform-provider-testllm/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func assertJSONSemanticallyEqual(t *testing.T, expected, actual string) {
	t.Helper()
	match, diags := jsontypes.NewNormalizedValue(expected).StringSemanticEquals(
		context.Background(),
		jsontypes.NewNormalizedValue(actual),
	)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if !match {
		t.Fatalf("expected JSON %q to equal %q", expected, actual)
	}
}

func TestExpandTestItems_preservesJSON(t *testing.T) {
	arguments := `{"command": "agyn threads send --message \"Thinking\" > /dev/null && echo ok", "meta": {"z": 1, "a": 2}}`
	blocks := `[ { "type": "text", "text": "Use > /dev/null && echo ok" } ]`
	systemBlocks := `[{"type": "text", "text": "System > /dev/null && echo ok"}]`
	items := []testItemModel{
		{
			Type:          types.StringValue("function_call"),
			Role:          types.StringNull(),
			Content:       types.StringNull(),
			Text:          types.StringNull(),
			ContentBlocks: jsontypes.NewNormalizedNull(),
			AnyRole:       types.BoolValue(false),
			AnyContent:    types.BoolValue(false),
			Repeat:        types.BoolValue(false),
			CallID:        types.StringValue("call-1"),
			FuncName:      types.StringValue("get_data"),
			Arguments:     jsontypes.NewNormalizedValue(arguments),
			Output:        types.StringNull(),
		},
		{
			Type:          types.StringValue("anthropic_message"),
			Role:          types.StringValue("assistant"),
			Content:       types.StringNull(),
			Text:          types.StringNull(),
			ContentBlocks: jsontypes.NewNormalizedValue(blocks),
			AnyRole:       types.BoolValue(false),
			AnyContent:    types.BoolValue(false),
			Repeat:        types.BoolValue(false),
			CallID:        types.StringNull(),
			FuncName:      types.StringNull(),
			Arguments:     jsontypes.NewNormalizedNull(),
			Output:        types.StringNull(),
		},
		{
			Type:          types.StringValue("anthropic_system"),
			Role:          types.StringNull(),
			Content:       types.StringNull(),
			Text:          types.StringNull(),
			ContentBlocks: jsontypes.NewNormalizedValue(systemBlocks),
			AnyRole:       types.BoolValue(false),
			AnyContent:    types.BoolValue(false),
			Repeat:        types.BoolValue(false),
			CallID:        types.StringNull(),
			FuncName:      types.StringNull(),
			Arguments:     jsontypes.NewNormalizedNull(),
			Output:        types.StringNull(),
		},
	}

	expanded, diags := expandTestItems(items)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if len(expanded) != 3 {
		t.Fatalf("expected 3 items, got %d", len(expanded))
	}

	_, _, expandedArgs, err := client.ParseFunctionCallContent(expanded[0])
	if err != nil {
		t.Fatalf("parse function_call item: %v", err)
	}
	if expandedArgs != arguments {
		t.Fatalf("expected arguments %q, got %q", arguments, expandedArgs)
	}

	messageContent, err := client.ParseAnthropicMessageContent(expanded[1])
	if err != nil {
		t.Fatalf("parse anthropic_message item: %v", err)
	}
	if messageContent.ContentBlocks == nil {
		t.Fatalf("expected content blocks to be set")
	}
	assertJSONSemanticallyEqual(t, blocks, string(messageContent.ContentBlocks))

	systemContent, err := client.ParseAnthropicSystemContent(expanded[2])
	if err != nil {
		t.Fatalf("parse anthropic_system item: %v", err)
	}
	if systemContent.Blocks == nil {
		t.Fatalf("expected system content blocks to be set")
	}
	assertJSONSemanticallyEqual(t, systemBlocks, string(systemContent.Blocks))
}

func TestFlattenTestItems_preservesJSON(t *testing.T) {
	arguments := `{"command": "agyn threads send --message \"Thinking\" > /dev/null && echo ok", "meta": {"z": 1, "a": 2}}`
	callItem, err := client.NewFunctionCallItem("call-1", "get_data", arguments)
	if err != nil {
		t.Fatalf("build function_call item: %v", err)
	}

	blocks := json.RawMessage(`[ { "type": "text", "text": "Use > /dev/null && echo ok" } ]`)
	messageItem, err := client.NewAnthropicMessageBlocksItem("assistant", blocks, nil)
	if err != nil {
		t.Fatalf("build anthropic_message item: %v", err)
	}

	systemBlocks := json.RawMessage(`[{"type": "text", "text": "System > /dev/null && echo ok"}]`)
	systemItem, err := client.NewAnthropicSystemBlocksItem(systemBlocks, nil)
	if err != nil {
		t.Fatalf("build anthropic_system item: %v", err)
	}

	flattened, diags := flattenTestItems([]client.TestItem{callItem, messageItem, systemItem})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if len(flattened) != 3 {
		t.Fatalf("expected 3 items, got %d", len(flattened))
	}

	if flattened[0].Arguments.ValueString() != arguments {
		t.Fatalf("expected arguments %q, got %q", arguments, flattened[0].Arguments.ValueString())
	}
	assertJSONSemanticallyEqual(t, string(blocks), flattened[1].ContentBlocks.ValueString())
	assertJSONSemanticallyEqual(t, string(systemBlocks), flattened[2].ContentBlocks.ValueString())
}
