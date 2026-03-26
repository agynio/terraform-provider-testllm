package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"testllm": providerserver.NewProtocol6WithError(New()),
}

func testAccPreCheck(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set; skipping acceptance tests")
	}
	if value := os.Getenv("TESTLLM_HOST"); value == "" {
		t.Fatal("TESTLLM_HOST must be set for acceptance tests")
	}
	if value := os.Getenv("TESTLLM_TOKEN"); value == "" {
		t.Fatal("TESTLLM_TOKEN must be set for acceptance tests")
	}
}

func testAccProviderConfig() string {
	return `provider "testllm" {}`
}
