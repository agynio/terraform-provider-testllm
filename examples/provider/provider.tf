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
  token = var.testllm_token
}
