resource "testllm_organization" "example" {
  name = "Acme Corp"
  slug = "acme-corp"
}

resource "testllm_test_suite" "example" {
  org_id      = testllm_organization.example.id
  name        = "Smoke Tests"
  description = "Basic sanity checks"
}

resource "testllm_test_suite" "anthropic" {
  org_id   = testllm_organization.example.id
  name     = "Anthropic Tests"
  protocol = "anthropic"
}

resource "testllm_test" "example" {
  org_id   = testllm_organization.example.id
  suite_id = testllm_test_suite.example.id
  name     = "greeting-flow"

  items = [
    {
      type    = "message"
      role    = "user"
      content = "Say hello"
    },
    {
      type    = "message"
      role    = "assistant"
      content = "Hello! How can I help you?"
    },
  ]
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
      content_blocks = jsonencode([{ type = "text", text = "Hi there!" }])
    },
  ]
}
