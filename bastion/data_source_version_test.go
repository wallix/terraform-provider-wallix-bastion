package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceVersion_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVersionData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.wallix-bastion_version.testacc_version",
						"version"),
				),
			},
		},
		PreventPostDestroyRefresh: true,
	})
}

func testAccDataSourceVersionData() string {
	return `
data "wallix-bastion_version" "testacc_version" {}
`
}
