package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceConnectionPolicy_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				// Create the resource to be fetched by the datasource.
				Config: testAccDataSourceConnectionPolicyConfigCreate(),
			},
			{
				// Validate that the datasource correctly retrieves the resource.
				Config: testAccDataSourceConnectionPolicyConfigData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.wallix-bastion_connection_policy.testacc_dataConnectionPolicy",
						"id"),
					resource.TestCheckResourceAttr("data.wallix-bastion_connection_policy.testacc_dataConnectionPolicy",
						"connection_policy_name", "testacc_ConnectionPolicy_ds"),
					resource.TestCheckResourceAttr("data.wallix-bastion_connection_policy.testacc_dataConnectionPolicy",
						"protocol", "RAWTCPIP"),
					resource.TestCheckResourceAttrSet("data.wallix-bastion_connection_policy.testacc_dataConnectionPolicy",
						"options"),
				),
			},
		},
	})
}

// Resource creation configuration (simplest variant: RAWTCPIP protocol, no
// authentication_methods / description needed).
func testAccDataSourceConnectionPolicyConfigCreate() string {
	return `
locals {
  optionsRAWTCPIP_ds = {
    nat_redirection = {
      enable = false
      host   = ""
      port   = 0
    }
    seamless_connection = {
      ipredir_apps = ""
      mode         = "iploop"
    }
  }
}

resource "wallix-bastion_connection_policy" "testacc_ConnectionPolicy_ds" {
  connection_policy_name = "testacc_ConnectionPolicy_ds"
  protocol                = "RAWTCPIP"
  options                 = jsonencode(local.optionsRAWTCPIP_ds)
}
`
}

// Datasource configuration to retrieve the created resource.
func testAccDataSourceConnectionPolicyConfigData() string {
	return `
locals {
  optionsRAWTCPIP_ds = {
    nat_redirection = {
      enable = false
      host   = ""
      port   = 0
    }
    seamless_connection = {
      ipredir_apps = ""
      mode         = "iploop"
    }
  }
}

resource "wallix-bastion_connection_policy" "testacc_ConnectionPolicy_ds" {
  connection_policy_name = "testacc_ConnectionPolicy_ds"
  protocol                = "RAWTCPIP"
  options                 = jsonencode(local.optionsRAWTCPIP_ds)
}

data "wallix-bastion_connection_policy" "testacc_dataConnectionPolicy" {
  connection_policy_name = wallix-bastion_connection_policy.testacc_ConnectionPolicy_ds.connection_policy_name
}
`
}
