---
page_title: "TestLLM Provider"
description: "Manage TestLLM organizations, test suites, and tests."
---

# TestLLM Provider

The TestLLM provider manages TestLLM organizations, test suites, and tests.

## Example Usage

```hcl
terraform {
  required_providers {
    testllm = {
      source = "agynio/testllm"
    }
  }
}

provider "testllm" {
  token = var.testllm_token
}
```

## Authentication

TestLLM supports two token types:

- `tlp_` personal access tokens for user-level access.
- `tlo_` organization tokens scoped to a specific organization.

## Schema

### Optional

- `host` (String) TestLLM API base URL. Defaults to `https://testllm.dev`.
- `token` (String, Sensitive) API authentication token.

~> **Note:** You can also set `TESTLLM_HOST` and `TESTLLM_TOKEN` environment variables to configure the provider.
