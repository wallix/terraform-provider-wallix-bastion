package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceLocalPasswordPolicy_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceLocalPasswordPolicyData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.wallix-bastion_local_password_policy.default",
						"id"),
				),
			},
		},
		PreventPostDestroyRefresh: true,
	})
}

func testAccDataSourceLocalPasswordPolicyData() string {
	return `
data "wallix-bastion_local_password_policy" "default" {}
`
}
