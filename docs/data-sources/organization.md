---
page_title: "testllm_organization Data Source"
description: "Look up TestLLM organizations by slug."
---

# testllm_organization Data Source

## Example Usage

```hcl
data "testllm_organization" "example" {
  slug = "acme-corp"
}

resource "testllm_test_suite" "example" {
  org_id = data.testllm_organization.example.id
  name   = "Smoke Tests"
}
```

## Schema

### Required

- `slug` (String) The unique slug of the organization to look up.

### Read-Only

- `id` (String) Unique identifier for the organization.
- `name` (String) Display name for the organization.
- `created_at` (String) Timestamp when the organization was created.
- `updated_at` (String) Timestamp when the organization was last updated.
