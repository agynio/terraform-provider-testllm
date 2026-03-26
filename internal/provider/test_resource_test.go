package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccTestResource_basic(t *testing.T) {
	orgName := acctest.RandomWithPrefix("tf-org")
	orgSlug := acctest.RandomWithPrefix("tf-org")
	suiteName := acctest.RandomWithPrefix("tf-suite")
	testName := acctest.RandomWithPrefix("tf-test")
	updatedName := acctest.RandomWithPrefix("tf-test-update")

	items := `[
  {
    type    = "message"
    role    = "user"
    content = "Say hello"
  },
  {
    type    = "message"
    role    = "assistant"
    content = "Hello"
  },
]`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTestResourceConfig(orgName, orgSlug, suiteName, testName, "", items),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("testllm_test.test", "name", testName),
					resource.TestCheckResourceAttr("testllm_test.test", "items.0.type", "message"),
					resource.TestCheckResourceAttr("testllm_test.test", "items.0.role", "user"),
					resource.TestCheckResourceAttr("testllm_test.test", "items.1.role", "assistant"),
					resource.TestCheckResourceAttrSet("testllm_test.test", "created_at"),
					resource.TestCheckResourceAttrSet("testllm_test.test", "updated_at"),
				),
			},
			{
				Config: testAccTestResourceConfig(orgName, orgSlug, suiteName, updatedName, "", items),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("testllm_test.test", "name", updatedName),
				),
			},
			{
				ResourceName:      "testllm_test.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					resourceState, ok := state.RootModule().Resources["testllm_test.test"]
					if !ok {
						return "", fmt.Errorf("resource not found in state")
					}
					orgID := resourceState.Primary.Attributes["org_id"]
					suiteID := resourceState.Primary.Attributes["suite_id"]
					testID := resourceState.Primary.ID
					return fmt.Sprintf("%s/%s/%s", orgID, suiteID, testID), nil
				},
			},
		},
	})
}

func TestAccTestResource_functionCallItems(t *testing.T) {
	orgName := acctest.RandomWithPrefix("tf-org")
	orgSlug := acctest.RandomWithPrefix("tf-org")
	suiteName := acctest.RandomWithPrefix("tf-suite")
	testName := acctest.RandomWithPrefix("tf-test")

	items := `[
  {
    type      = "function_call"
    call_id   = "call-1"
    func_name = "get_data"
    arguments = jsonencode({ foo = "bar" })
  },
  {
    type    = "function_call_output"
    call_id = "call-1"
    output  = "done"
  },
]`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTestResourceConfig(orgName, orgSlug, suiteName, testName, "", items),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("testllm_test.test", "items.0.type", "function_call"),
					resource.TestCheckResourceAttr("testllm_test.test", "items.0.func_name", "get_data"),
					resource.TestCheckResourceAttr("testllm_test.test", "items.1.type", "function_call_output"),
					resource.TestCheckResourceAttr("testllm_test.test", "items.1.output", "done"),
				),
			},
		},
	})
}

func TestAccTestResource_validateConfig(t *testing.T) {
	orgName := acctest.RandomWithPrefix("tf-org")
	orgSlug := acctest.RandomWithPrefix("tf-org")
	suiteName := acctest.RandomWithPrefix("tf-suite")
	testName := acctest.RandomWithPrefix("tf-test")

	missingRoleItems := `[
  {
    type    = "message"
    content = "Say hello"
  }
]`

	unexpectedFieldItems := `[
  {
    type    = "message"
    role    = "user"
    content = "Say hello"
    call_id = "nope"
  }
]`

	invalidTypeItems := `[
  {
    type    = "unknown"
    role    = "user"
    content = "Say hello"
  }
]`

	invalidBoolItems := `[
  {
		type        = "message"
		role        = "assistant"
		content     = "Hello"
		any_role    = true
		any_content = true
		repeat      = true
  }
]`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccTestResourceConfig(orgName, orgSlug, suiteName, testName, "", missingRoleItems),
				ExpectError: regexp.MustCompile("role"),
			},
			{
				Config:      testAccTestResourceConfig(orgName, orgSlug, suiteName, testName, "", unexpectedFieldItems),
				ExpectError: regexp.MustCompile("call_id"),
			},
			{
				Config:      testAccTestResourceConfig(orgName, orgSlug, suiteName, testName, "", invalidTypeItems),
				ExpectError: regexp.MustCompile("type"),
			},
			{
				Config:      testAccTestResourceConfig(orgName, orgSlug, suiteName, testName, "", invalidBoolItems),
				ExpectError: regexp.MustCompile("any_role"),
			},
		},
	})
}

func testAccTestResourceConfig(orgName, orgSlug, suiteName, testName, description, items string) string {
	return fmt.Sprintf(`
%s

resource "testllm_organization" "test" {
  name = %q
  slug = %q
}

resource "testllm_test_suite" "test" {
  org_id = testllm_organization.test.id
  name   = %q
}

resource "testllm_test" "test" {
  org_id      = testllm_organization.test.id
  suite_id    = testllm_test_suite.test.id
  name        = %q
  description = %q
  items       = %s
}
`, testAccProviderConfig(), orgName, orgSlug, suiteName, testName, description, items)
}
