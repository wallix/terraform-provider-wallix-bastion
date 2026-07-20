package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourcePasswordChangePolicy_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePasswordChangePolicyCreate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"wallix-bastion_passwordchangepolicy.testacc_PCP",
						"id"),
					resource.TestCheckResourceAttr(
						"wallix-bastion_passwordchangepolicy.testacc_PCP",
						"password_length", "12"),
					resource.TestCheckResourceAttr(
						"wallix-bastion_passwordchangepolicy.testacc_PCP",
						"lower_chars", "1"),
					// special_chars is left unconfigured: the API leaves it null (character
					// class disallowed entirely), which reads back into Terraform as 0.
					resource.TestCheckResourceAttr(
						"wallix-bastion_passwordchangepolicy.testacc_PCP",
						"special_chars", "0"),
				),
			},
			{
				Config: testAccResourcePasswordChangePolicyUpdate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"wallix-bastion_passwordchangepolicy.testacc_PCP",
						"description", "testacc updated description"),
					// special_chars is now explicitly set to 0 (allowed, no minimum enforced) -
					// exercises the tri-state handling distinguishing "unset" from "explicit 0".
					resource.TestCheckResourceAttr(
						"wallix-bastion_passwordchangepolicy.testacc_PCP",
						"special_chars", "0"),
					resource.TestCheckResourceAttr(
						"wallix-bastion_passwordchangepolicy.testacc_PCP",
						"upper_chars", "2"),
					resource.TestCheckResourceAttr(
						"wallix-bastion_passwordchangepolicy.testacc_PCP",
						"exclude_chars", "!@#$"),
				),
			},
			{
				ResourceName:  "wallix-bastion_passwordchangepolicy.testacc_PCP",
				ImportState:   true,
				ImportStateId: "testacc_PCP",
			},
		},
		PreventPostDestroyRefresh: true,
	})
}

func testAccResourcePasswordChangePolicyCreate() string {
	return `
resource "wallix-bastion_passwordchangepolicy" "testacc_PCP" {
  password_change_policy_name = "testacc_PCP"
  password_length             = 12
  lower_chars                 = 1
}
`
}

func testAccResourcePasswordChangePolicyUpdate() string {
	return `
resource "wallix-bastion_passwordchangepolicy" "testacc_PCP" {
  password_change_policy_name = "testacc_PCP"
  description                 = "testacc updated description"
  password_length             = 12
  lower_chars                 = 1
  upper_chars                 = 2
  special_chars               = 0
  exclude_chars               = "!@#$"
}
`
}
