package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOrganizationDataSource_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-org")
	slug := acctest.RandomWithPrefix("tf-org")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationDataSourceConfig(name, slug),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.testllm_organization.test", "id", "testllm_organization.test", "id"),
					resource.TestCheckResourceAttrPair("data.testllm_organization.test", "name", "testllm_organization.test", "name"),
					resource.TestCheckResourceAttrPair("data.testllm_organization.test", "slug", "testllm_organization.test", "slug"),
					resource.TestCheckResourceAttrSet("data.testllm_organization.test", "created_at"),
					resource.TestCheckResourceAttrSet("data.testllm_organization.test", "updated_at"),
				),
			},
		},
	})
}

func TestAccOrganizationDataSource_notFound(t *testing.T) {
	slug := acctest.RandomWithPrefix("tf-org-missing")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccOrganizationDataSourceNotFoundConfig(slug),
				ExpectError: regexp.MustCompile("Organization not found"),
			},
		},
	})
}

func testAccOrganizationDataSourceConfig(name, slug string) string {
	return fmt.Sprintf(`
%s

resource "testllm_organization" "test" {
  name = %q
  slug = %q
}

data "testllm_organization" "test" {
  slug = testllm_organization.test.slug
}
`, testAccProviderConfig(), name, slug)
}

func testAccOrganizationDataSourceNotFoundConfig(slug string) string {
	return fmt.Sprintf(`
%s

data "testllm_organization" "test" {
  slug = %q
}
`, testAccProviderConfig(), slug)
}
