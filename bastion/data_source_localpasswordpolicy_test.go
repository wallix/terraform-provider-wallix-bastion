package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceLocalpasswordpolicy_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceLocalpasswordpolicyData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.wallix-bastion_localpasswordpolicy.default",
						"id"),
				),
			},
		},
		PreventPostDestroyRefresh: true,
	})
}

func testAccDataSourceLocalpasswordpolicyData() string {
	return `
data "wallix-bastion_localpasswordpolicy" "default" {}
`
}
