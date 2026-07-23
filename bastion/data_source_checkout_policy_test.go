package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceCheckoutPolicy_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				// Create the resource to be fetched by the datasource.
				Config: testAccDataSourceCheckoutPolicyConfigCreate(),
			},
			{
				// Validate that the datasource correctly retrieves the resource.
				Config: testAccDataSourceCheckoutPolicyConfigData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.wallix-bastion_checkout_policy.testacc_dataCheckoutPolicy",
						"id"),
					resource.TestCheckResourceAttr("data.wallix-bastion_checkout_policy.testacc_dataCheckoutPolicy",
						"checkout_policy_name", "testacc_CheckoutPolicy_ds"),
				),
			},
		},
	})
}

// Resource creation configuration.
func testAccDataSourceCheckoutPolicyConfigCreate() string {
	return `
resource "wallix-bastion_checkout_policy" "testacc_CheckoutPolicy_ds" {
  checkout_policy_name = "testacc_CheckoutPolicy_ds"
}
`
}

// Datasource configuration to retrieve the created resource.
func testAccDataSourceCheckoutPolicyConfigData() string {
	return `
resource "wallix-bastion_checkout_policy" "testacc_CheckoutPolicy_ds" {
  checkout_policy_name = "testacc_CheckoutPolicy_ds"
}

data "wallix-bastion_checkout_policy" "testacc_dataCheckoutPolicy" {
  checkout_policy_name = wallix-bastion_checkout_policy.testacc_CheckoutPolicy_ds.checkout_policy_name
}
`
}
