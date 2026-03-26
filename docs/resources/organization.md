---
page_title: "testllm_organization Resource"
description: "Manage TestLLM organizations."
---

# testllm_organization Resource

## Example Usage

```hcl
resource "testllm_organization" "example" {
  name = "Acme Corp"
  slug = "acme-corp"
}
```

## Schema

### Required

- `name` (String) Display name for the organization.
- `slug` (String) Unique slug for the organization.

### Read-Only

- `id` (String) Unique identifier for the organization.
- `created_at` (String) Timestamp when the organization was created.
- `updated_at` (String) Timestamp when the organization was last updated.

## Import

```sh
terraform import testllm_organization.example <org-id>
```
