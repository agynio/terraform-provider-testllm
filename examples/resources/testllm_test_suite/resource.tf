resource "testllm_organization" "example" {
  name = "Acme Corp"
  slug = "acme-corp"
}

resource "testllm_test_suite" "example" {
  org_id      = testllm_organization.example.id
  name        = "Smoke Tests"
  description = "Basic sanity checks"
}
