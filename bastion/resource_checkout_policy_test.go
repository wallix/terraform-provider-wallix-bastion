package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceCheckoutPolicy_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceCheckoutPolicyCreate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"wallix-bastion_checkout_policy.testacc_CheckoutPolicy",
						"id"),
				),
			},
			{
				Config: testAccResourceCheckoutPolicyUpdate(),
			},
			{
				ResourceName:  "wallix-bastion_checkout_policy.testacc_CheckoutPolicy",
				ImportState:   true,
				ImportStateId: "testacc_CheckoutPolicy",
			},
		},
		PreventPostDestroyRefresh: true,
	})
}

func testAccResourceCheckoutPolicyCreate() string {
	return `
resource "wallix-bastion_checkout_policy" "testacc_CheckoutPolicy" {
  checkout_policy_name = "testacc_CheckoutPolicy"
}
`
}

func testAccResourceCheckoutPolicyUpdate() string {
	return `
resource "wallix-bastion_checkout_policy" "testacc_CheckoutPolicy" {
  checkout_policy_name          = "testacc_CheckoutPolicy"
  description                   = "testacc CheckoutPolicy"
  enable_lock                   = true
  change_credentials_at_checkin = true
  duration                      = 60
  extension                     = 60
  max_duration                  = 180
}
`
}
