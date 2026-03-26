terraform {
  required_providers {
    testllm = {
      source = "agynio/testllm"
    }
  }
}

variable "testllm_token" {
  type = string
}

provider "testllm" {
  host  = "https://testllm.example.com"
  token = var.testllm_token
}
