---
page_title: "testllm_test Resource"
description: "Manage TestLLM tests."
---

# testllm_test Resource

## Example Usage

```hcl
resource "testllm_organization" "example" {
  name = "Acme Corp"
  slug = "acme-corp"
}

resource "testllm_test_suite" "example" {
  org_id = testllm_organization.example.id
  name   = "Smoke Tests"
}

resource "testllm_test" "example" {
  org_id      = testllm_organization.example.id
  suite_id    = testllm_test_suite.example.id
  name        = "greeting-flow"
  description = "Validate messages and tool calls"

  items = [
    {
      type    = "message"
      role    = "user"
      content = "Say hello"
    },
    {
      type      = "function_call"
      call_id   = "call-1"
      func_name = "lookup_user"
      arguments = jsonencode({ id = "123" })
    },
    {
      type    = "function_call_output"
      call_id = "call-1"
      output  = "ok"
    },
  ]
}
```

## Anthropic Example Usage

```hcl
resource "testllm_organization" "example" {
  name = "Acme Corp"
  slug = "acme-corp"
}

resource "testllm_test_suite" "anthropic" {
  org_id   = testllm_organization.example.id
  name     = "Anthropic Tests"
  protocol = "anthropic"
}

resource "testllm_test" "anthropic" {
  org_id   = testllm_organization.example.id
  suite_id = testllm_test_suite.anthropic.id
  name     = "anthropic-flow"

  items = [
    {
      type = "anthropic_system"
      text = "You are a helpful assistant."
    },
    {
      type    = "anthropic_message"
      role    = "user"
      content = "Hello"
    },
    {
      type           = "anthropic_message"
      role           = "assistant"
      content_blocks = jsonencode([{ type = "text", text = "Hi!" }])
    },
  ]
}
```

## Schema

### Required

- `org_id` (String) Organization ID that owns the test.
- `suite_id` (String) Test suite ID that owns the test.
- `name` (String) Display name for the test.
- `items` (List of Object) Ordered list of test items that define the test flow.

### Optional

- `description` (String) Description of the test.

### Read-Only

- `id` (String) Unique identifier for the test.
- `created_at` (String) Timestamp when the test was created.
- `updated_at` (String) Timestamp when the test was last updated.

### Nested Schema for `items`

Required:

- `type` (String) Item type. One of `message`, `function_call`, `function_call_output`, `anthropic_system`, or `anthropic_message`.

Optional:

- `role` (String) Role for message and anthropic_message items (`user`, `system`, `developer`, `assistant`).
- `content` (String) Content for message items.
- `text` (String) Text content for `anthropic_system` items.
- `content_blocks` (String) JSON-encoded array of Anthropic content blocks.
- `any_role` (Bool) Whether any role is accepted for message items.
- `any_content` (Bool) Whether any content is accepted for message items.
- `repeat` (Bool) Whether the message item can repeat.
- `call_id` (String) Function call identifier for `function_call` and `function_call_output` items.
- `func_name` (String) Function name for `function_call` items.
- `arguments` (String) JSON-encoded arguments for `function_call` items.
- `output` (String) Output content for `function_call_output` items.

## Import

```sh
terraform import testllm_test.example <org-id>/<suite-id>/<test-id>
```
