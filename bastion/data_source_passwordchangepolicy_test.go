package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourcePasswordChangePolicy_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				// Create the resource to be fetched by the datasource.
				Config: testAccDataSourcePasswordChangePolicyConfigCreate(),
			},
			{
				// Validate that the datasource correctly retrieves the resource.
				Config: testAccDataSourcePasswordChangePolicyConfigData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.wallix-bastion_passwordchangepolicy.testacc_dataPCP",
						"password_change_policy_name", "testacc_dataPCP"),
					resource.TestCheckResourceAttr("data.wallix-bastion_passwordchangepolicy.testacc_dataPCP",
						"description", "testacc data PCP"),
					resource.TestCheckResourceAttr("data.wallix-bastion_passwordchangepolicy.testacc_dataPCP",
						"password_length", "12"),
					resource.TestCheckResourceAttr("data.wallix-bastion_passwordchangepolicy.testacc_dataPCP",
						"lower_chars", "1"),
					resource.TestCheckResourceAttr("data.wallix-bastion_passwordchangepolicy.testacc_dataPCP",
						"upper_chars", "2"),
					resource.TestCheckResourceAttr("data.wallix-bastion_passwordchangepolicy.testacc_dataPCP",
						"special_chars", "0"),
					resource.TestCheckResourceAttr("data.wallix-bastion_passwordchangepolicy.testacc_dataPCP",
						"exclude_chars", "!@#$"),
				),
			},
		},
	})
}

// Resource creation configuration.
func testAccDataSourcePasswordChangePolicyConfigCreate() string {
	return `
resource "wallix-bastion_passwordchangepolicy" "testacc_dataPCP" {
  password_change_policy_name = "testacc_dataPCP"
  description                 = "testacc data PCP"
  password_length             = 12
  lower_chars                 = 1
  upper_chars                 = 2
  special_chars               = 0
  exclude_chars               = "!@#$"
}
`
}

// Datasource configuration to retrieve the created resource.
func testAccDataSourcePasswordChangePolicyConfigData() string {
	return `
resource "wallix-bastion_passwordchangepolicy" "testacc_dataPCP" {
  password_change_policy_name = "testacc_dataPCP"
  description                 = "testacc data PCP"
  password_length             = 12
  lower_chars                 = 1
  upper_chars                 = 2
  special_chars               = 0
  exclude_chars               = "!@#$"
}

data "wallix-bastion_passwordchangepolicy" "testacc_dataPCP" {
  password_change_policy_name = wallix-bastion_passwordchangepolicy.testacc_dataPCP.password_change_policy_name
}
`
}
