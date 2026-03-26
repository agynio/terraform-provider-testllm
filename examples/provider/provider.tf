terraform {
  required_providers {
    testllm = {
      source = "agynio/testllm"
    }
  }
}

provider "testllm" {
  host  = "https://testllm.example.com"
  token = var.testllm_token
}
