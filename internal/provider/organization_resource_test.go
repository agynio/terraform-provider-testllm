package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func TestAccOrganizationResource_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-org")
	slug := acctest.RandomWithPrefix("tf-org")
	updatedName := acctest.RandomWithPrefix("tf-org-update")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationResourceConfig(name, slug),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("testllm_organization.test", "name", name),
					resource.TestCheckResourceAttr("testllm_organization.test", "slug", slug),
					resource.TestCheckResourceAttrSet("testllm_organization.test", "created_at"),
					resource.TestCheckResourceAttrSet("testllm_organization.test", "updated_at"),
				),
			},
			{
				Config: testAccOrganizationResourceConfig(updatedName, slug),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("testllm_organization.test", "name", updatedName),
					resource.TestCheckResourceAttr("testllm_organization.test", "slug", slug),
				),
			},
			{
				ResourceName:      "testllm_organization.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccOrganizationResource_slugRequiresReplace(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-org")
	slug := acctest.RandomWithPrefix("tf-org")
	replaceSlug := acctest.RandomWithPrefix("tf-org-new")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationResourceConfig(name, slug),
			},
			{
				Config: testAccOrganizationResourceConfig(name, replaceSlug),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("testllm_organization.test", plancheck.ResourceActionReplace),
					},
				},
			},
		},
	})
}

func testAccOrganizationResourceConfig(name, slug string) string {
	return fmt.Sprintf(`
%s

resource "testllm_organization" "test" {
  name = %q
  slug = %q
}
`, testAccProviderConfig(), name, slug)
}
