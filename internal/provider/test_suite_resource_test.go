package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccTestSuiteResource_basic(t *testing.T) {
	orgName := acctest.RandomWithPrefix("tf-org")
	orgSlug := acctest.RandomWithPrefix("tf-org")
	suiteName := acctest.RandomWithPrefix("tf-suite")
	updatedName := acctest.RandomWithPrefix("tf-suite-update")
	updatedDescription := "updated description"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTestSuiteResourceConfig(orgName, orgSlug, suiteName, "", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("testllm_test_suite.test", "name", suiteName),
					resource.TestCheckResourceAttr("testllm_test_suite.test", "description", ""),
					resource.TestCheckResourceAttrSet("testllm_test_suite.test", "created_at"),
					resource.TestCheckResourceAttrSet("testllm_test_suite.test", "updated_at"),
				),
			},
			{
				Config: testAccTestSuiteResourceConfig(orgName, orgSlug, updatedName, updatedDescription, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("testllm_test_suite.test", "name", updatedName),
					resource.TestCheckResourceAttr("testllm_test_suite.test", "description", updatedDescription),
				),
			},
			{
				ResourceName:      "testllm_test_suite.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					resourceState, ok := state.RootModule().Resources["testllm_test_suite.test"]
					if !ok {
						return "", fmt.Errorf("resource not found in state")
					}
					orgID := resourceState.Primary.Attributes["org_id"]
					suiteID := resourceState.Primary.ID
					return fmt.Sprintf("%s/%s", orgID, suiteID), nil
				},
			},
		},
	})
}

func testAccTestSuiteResourceConfig(orgName, orgSlug, suiteName, description string, includeDescription bool) string {
	descriptionLine := ""
	if includeDescription {
		descriptionLine = fmt.Sprintf("  description = %q\n", description)
	}

	return fmt.Sprintf(`
%s

resource "testllm_organization" "test" {
  name = %q
  slug = %q
}

resource "testllm_test_suite" "test" {
  org_id = testllm_organization.test.id
  name   = %q
%s}
`, testAccProviderConfig(), orgName, orgSlug, suiteName, descriptionLine)
}
