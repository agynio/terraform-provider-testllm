---
page_title: "testllm_test_suite Resource"
description: "Manage TestLLM test suites."
---

# testllm_test_suite Resource

## Example Usage

```hcl
resource "testllm_organization" "example" {
  name = "Acme Corp"
  slug = "acme-corp"
}

resource "testllm_test_suite" "example" {
  org_id      = testllm_organization.example.id
  name        = "Smoke Tests"
  description = "Basic sanity checks"
}
```

## Schema

### Required

- `org_id` (String) Organization ID that owns the test suite.
- `name` (String) Display name for the test suite.

### Optional

- `description` (String) Description of the test suite.

### Read-Only

- `id` (String) Unique identifier for the test suite.
- `created_at` (String) Timestamp when the test suite was created.
- `updated_at` (String) Timestamp when the test suite was last updated.

## Import

```sh
terraform import testllm_test_suite.example <org-id>/<suite-id>
```
