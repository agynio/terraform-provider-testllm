package provider

import (
	"encoding/json"
	"testing"

	"github.com/agynio/terraform-provider-testllm/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestExpandTestItems_normalizesJSON(t *testing.T) {
	arguments := `{"z":1,"a":2}`
	blocks := `[
  {"type":"tool_use","id":"tool-1","name":"get_weather","input":{"z":1,"a":2}}
]`
	items := []testItemModel{
		{
			Type:          types.StringValue("function_call"),
			Role:          types.StringNull(),
			Content:       types.StringNull(),
			Text:          types.StringNull(),
			ContentBlocks: types.StringNull(),
			AnyRole:       types.BoolValue(false),
			AnyContent:    types.BoolValue(false),
			Repeat:        types.BoolValue(false),
			CallID:        types.StringValue("call-1"),
			FuncName:      types.StringValue("get_data"),
			Arguments:     types.StringValue(arguments),
			Output:        types.StringNull(),
		},
		{
			Type:          types.StringValue("anthropic_message"),
			Role:          types.StringValue("assistant"),
			Content:       types.StringNull(),
			Text:          types.StringNull(),
			ContentBlocks: types.StringValue(blocks),
			AnyRole:       types.BoolValue(false),
			AnyContent:    types.BoolValue(false),
			Repeat:        types.BoolValue(false),
			CallID:        types.StringNull(),
			FuncName:      types.StringNull(),
			Arguments:     types.StringNull(),
			Output:        types.StringNull(),
		},
	}

	expanded, diags := expandTestItems(items)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if len(expanded) != 2 {
		t.Fatalf("expected 2 items, got %d", len(expanded))
	}

	_, _, normalizedArgs, err := client.ParseFunctionCallContent(expanded[0])
	if err != nil {
		t.Fatalf("parse function_call item: %v", err)
	}
	if normalizedArgs != `{"a":2,"z":1}` {
		t.Fatalf("expected normalized arguments, got %q", normalizedArgs)
	}

	messageContent, err := client.ParseAnthropicMessageContent(expanded[1])
	if err != nil {
		t.Fatalf("parse anthropic_message item: %v", err)
	}
	if messageContent.ContentBlocks == nil {
		t.Fatalf("expected content blocks to be set")
	}
	if string(messageContent.ContentBlocks) != `[{"id":"tool-1","input":{"a":2,"z":1},"name":"get_weather","type":"tool_use"}]` {
		t.Fatalf("expected normalized content blocks, got %s", messageContent.ContentBlocks)
	}
}

func TestFlattenTestItems_normalizesJSON(t *testing.T) {
	arguments := `{"z":1,"a":2}`
	callItem, err := client.NewFunctionCallItem("call-1", "get_data", arguments)
	if err != nil {
		t.Fatalf("build function_call item: %v", err)
	}

	blocks := json.RawMessage(`[{"type":"tool_use","id":"tool-1","name":"get_weather","input":{"z":1,"a":2}}]`)
	messageItem, err := client.NewAnthropicMessageBlocksItem("assistant", blocks, nil)
	if err != nil {
		t.Fatalf("build anthropic_message item: %v", err)
	}

	flattened, diags := flattenTestItems([]client.TestItem{callItem, messageItem})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if len(flattened) != 2 {
		t.Fatalf("expected 2 items, got %d", len(flattened))
	}

	if flattened[0].Arguments.ValueString() != `{"a":2,"z":1}` {
		t.Fatalf("expected normalized arguments, got %q", flattened[0].Arguments.ValueString())
	}
	if flattened[1].ContentBlocks.ValueString() != `[{"id":"tool-1","input":{"a":2,"z":1},"name":"get_weather","type":"tool_use"}]` {
		t.Fatalf("expected normalized content blocks, got %q", flattened[1].ContentBlocks.ValueString())
	}
}
