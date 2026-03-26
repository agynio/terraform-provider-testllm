data "testllm_organization" "example" {
  slug = "acme-corp"
}

resource "testllm_test_suite" "example" {
  org_id = data.testllm_organization.example.id
  name   = "Smoke Tests"
}
