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

func TestAccTestResource_messageFlags(t *testing.T) {
	orgName := acctest.RandomWithPrefix("tf-org")
	orgSlug := acctest.RandomWithPrefix("tf-org")
	suiteName := acctest.RandomWithPrefix("tf-suite")
	testName := acctest.RandomWithPrefix("tf-test")

	items := `[
  {
    type        = "message"
    role        = "user"
    content     = "Say hello"
    any_role    = true
    any_content = true
    repeat      = true
  }
]`

	config := testAccTestResourceConfig(orgName, orgSlug, suiteName, testName, "", items)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("testllm_test.test", "items.0.any_role", "true"),
					resource.TestCheckResourceAttr("testllm_test.test", "items.0.any_content", "true"),
					resource.TestCheckResourceAttr("testllm_test.test", "items.0.repeat", "true"),
				),
			},
			{
				Config:             config,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccTestResource_anthropicSimple(t *testing.T) {
	orgName := acctest.RandomWithPrefix("tf-org")
	orgSlug := acctest.RandomWithPrefix("tf-org")
	suiteName := acctest.RandomWithPrefix("tf-suite")
	testName := acctest.RandomWithPrefix("tf-test")

	items := `[
  {
    type = "anthropic_system"
    text = "You are a helpful assistant."
  },
  {
    type    = "anthropic_message"
    role    = "user"
    content = "Hello"
  },
  {
    type           = "anthropic_message"
    role           = "assistant"
    content_blocks = jsonencode([{ type = "text", text = "Hi there!" }])
  },
]`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAnthropicTestResourceConfig(orgName, orgSlug, suiteName, testName, "", items),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("testllm_test.test", "items.0.type", "anthropic_system"),
					resource.TestCheckResourceAttr("testllm_test.test", "items.0.text", "You are a helpful assistant."),
					resource.TestCheckResourceAttr("testllm_test.test", "items.1.role", "user"),
					resource.TestCheckResourceAttr("testllm_test.test", "items.1.content", "Hello"),
					resource.TestCheckResourceAttrSet("testllm_test.test", "items.2.content_blocks"),
				),
			},
		},
	})
}

func TestAccTestResource_anthropicToolUse(t *testing.T) {
	orgName := acctest.RandomWithPrefix("tf-org")
	orgSlug := acctest.RandomWithPrefix("tf-org")
	suiteName := acctest.RandomWithPrefix("tf-suite")
	testName := acctest.RandomWithPrefix("tf-test")

	items := `[
  {
    type           = "anthropic_system"
    content_blocks = jsonencode([{ type = "text", text = "Use tools when needed." }])
  },
  {
    type    = "anthropic_message"
    role    = "user"
    content = "Lookup weather"
  },
  {
    type           = "anthropic_message"
    role           = "assistant"
    content_blocks = jsonencode([
      {
        type  = "tool_use"
        id    = "tool-1"
        name  = "get_weather"
        input = { location = "Portland" }
      }
    ])
  },
  {
    type           = "anthropic_message"
    role           = "assistant"
    content_blocks = jsonencode([
      {
        type        = "tool_result"
        tool_use_id = "tool-1"
        content     = "Sunny"
      }
    ])
  },
]`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAnthropicTestResourceConfig(orgName, orgSlug, suiteName, testName, "", items),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("testllm_test.test", "items.0.type", "anthropic_system"),
					resource.TestCheckResourceAttr("testllm_test.test", "items.1.role", "user"),
					resource.TestCheckResourceAttrSet("testllm_test.test", "items.2.content_blocks"),
					resource.TestCheckResourceAttrSet("testllm_test.test", "items.3.content_blocks"),
				),
			},
		},
	})
}

func TestAccTestResource_crossProtocolAnthropicInOpenAI(t *testing.T) {
	orgName := acctest.RandomWithPrefix("tf-org")
	orgSlug := acctest.RandomWithPrefix("tf-org")
	suiteName := acctest.RandomWithPrefix("tf-suite")
	testName := acctest.RandomWithPrefix("tf-test")

	items := `[
  {
    type = "anthropic_system"
    text = "You are a helpful assistant."
  }
]`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccTestResourceConfig(orgName, orgSlug, suiteName, testName, "", items),
				ExpectError: regexp.MustCompile("(?s).+"),
			},
		},
	})
}

func TestAccTestResource_crossProtocolOpenAIInAnthropic(t *testing.T) {
	orgName := acctest.RandomWithPrefix("tf-org")
	orgSlug := acctest.RandomWithPrefix("tf-org")
	suiteName := acctest.RandomWithPrefix("tf-suite")
	testName := acctest.RandomWithPrefix("tf-test")

	items := `[
  {
    type    = "message"
    role    = "user"
    content = "Say hello"
  }
]`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccAnthropicTestResourceConfig(orgName, orgSlug, suiteName, testName, "", items),
				ExpectError: regexp.MustCompile("(?s).+"),
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

	anthropicSystemMissingItems := `[
  {
    type = "anthropic_system"
  }
]`

	anthropicSystemBothItems := `[
  {
    type           = "anthropic_system"
    text           = "Hello"
    content_blocks = jsonencode([{ type = "text", text = "Hi" }])
  }
]`

	anthropicMessageMissingItems := `[
  {
    type = "anthropic_message"
    role = "user"
  }
]`

	anthropicMessageUnexpectedItems := `[
  {
    type    = "anthropic_message"
    role    = "user"
    content = "Hello"
    call_id = "nope"
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
			{
				Config:      testAccTestResourceConfig(orgName, orgSlug, suiteName, testName, "", anthropicSystemMissingItems),
				ExpectError: regexp.MustCompile("content_blocks|text"),
			},
			{
				Config:      testAccTestResourceConfig(orgName, orgSlug, suiteName, testName, "", anthropicSystemBothItems),
				ExpectError: regexp.MustCompile("content_blocks|text"),
			},
			{
				Config:      testAccTestResourceConfig(orgName, orgSlug, suiteName, testName, "", anthropicMessageMissingItems),
				ExpectError: regexp.MustCompile("content_blocks|content"),
			},
			{
				Config:      testAccTestResourceConfig(orgName, orgSlug, suiteName, testName, "", anthropicMessageUnexpectedItems),
				ExpectError: regexp.MustCompile("call_id"),
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

func testAccAnthropicTestResourceConfig(orgName, orgSlug, suiteName, testName, description, items string) string {
	return fmt.Sprintf(`
%s

resource "testllm_organization" "test" {
  name = %q
  slug = %q
}

resource "testllm_test_suite" "test" {
  org_id   = testllm_organization.test.id
  name     = %q
  protocol = "anthropic"
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
