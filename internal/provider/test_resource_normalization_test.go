package provider

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/agynio/terraform-provider-testllm/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestNormalizeJSON(t *testing.T) {
	t.Run("sorts keys", func(t *testing.T) {
		normalized, err := normalizeJSON(`{"z":1,"a":2}`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if normalized != `{"a":2,"z":1}` {
			t.Fatalf("expected normalized JSON, got %q", normalized)
		}
	})

	t.Run("idempotent", func(t *testing.T) {
		input := `{"a":2,"z":1}`
		normalized, err := normalizeJSON(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if normalized != input {
			t.Fatalf("expected %q, got %q", input, normalized)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		if _, err := normalizeJSON(`{"a":}`); err == nil {
			t.Fatalf("expected error for invalid JSON")
		}
	})

	t.Run("trailing data", func(t *testing.T) {
		if _, err := normalizeJSON(`{"a":1} {"b":2}`); err == nil {
			t.Fatalf("expected trailing data error")
		} else if !strings.Contains(err.Error(), "unexpected trailing data") {
			t.Fatalf("expected trailing data error, got %v", err)
		}
	})

	t.Run("numeric precision", func(t *testing.T) {
		normalized, err := normalizeJSON(`{"big":9007199254740993}`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if normalized != `{"big":9007199254740993}` {
			t.Fatalf("expected precision preserved, got %q", normalized)
		}
	})

	t.Run("empty string", func(t *testing.T) {
		if _, err := normalizeJSON(""); err == nil {
			t.Fatalf("expected error for empty string")
		}
	})

	t.Run("whitespace string", func(t *testing.T) {
		if _, err := normalizeJSON(" \n\t "); err == nil {
			t.Fatalf("expected error for whitespace string")
		}
	})
}

func TestExpandTestItems_normalizesJSON(t *testing.T) {
	arguments := `{"z":1,"a":2}`
	blocks := `[
  {"type":"tool_use","id":"tool-1","name":"get_weather","input":{"z":1,"a":2}}
]`
	systemBlocks := `[
  {"text":"Welcome","type":"text","meta":{"z":1,"a":2}}
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
		{
			Type:          types.StringValue("anthropic_system"),
			Role:          types.StringNull(),
			Content:       types.StringNull(),
			Text:          types.StringNull(),
			ContentBlocks: types.StringValue(systemBlocks),
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
	if len(expanded) != 3 {
		t.Fatalf("expected 3 items, got %d", len(expanded))
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

	systemContent, err := client.ParseAnthropicSystemContent(expanded[2])
	if err != nil {
		t.Fatalf("parse anthropic_system item: %v", err)
	}
	if systemContent.Blocks == nil {
		t.Fatalf("expected system content blocks to be set")
	}
	if string(systemContent.Blocks) != `[{"meta":{"a":2,"z":1},"text":"Welcome","type":"text"}]` {
		t.Fatalf("expected normalized system blocks, got %s", systemContent.Blocks)
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

	systemBlocks := json.RawMessage(`[{"text":"Welcome","type":"text","meta":{"z":1,"a":2}}]`)
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

	if flattened[0].Arguments.ValueString() != `{"a":2,"z":1}` {
		t.Fatalf("expected normalized arguments, got %q", flattened[0].Arguments.ValueString())
	}
	if flattened[1].ContentBlocks.ValueString() != `[{"id":"tool-1","input":{"a":2,"z":1},"name":"get_weather","type":"tool_use"}]` {
		t.Fatalf("expected normalized content blocks, got %q", flattened[1].ContentBlocks.ValueString())
	}
	if flattened[2].ContentBlocks.ValueString() != `[{"meta":{"a":2,"z":1},"text":"Welcome","type":"text"}]` {
		t.Fatalf("expected normalized system blocks, got %q", flattened[2].ContentBlocks.ValueString())
	}
}
