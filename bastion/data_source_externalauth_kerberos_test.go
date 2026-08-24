// nolint: lll,nolintlint
package bastion_test

import (
	"encoding/hex"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	dataSourceKeytabDataHexStr = "0502000000320001000b4558414d504c452e434f4d00047573657200000001586aa82d01001700100c61039f010b2fbb88fe449fbf262477000000420001000b4558414d504c452e434f4d00047573657200000001586aa82d010012002053142f614ee6c39823710d9f31ff2984ed0bd9074d6e542e8468137f7b909c17000000320001000b4558414d504c452e434f4d00047573657200000001586beaad01001700100c61039f010b2fbb88fe449fbf262477000000420001000b4558414d504c452e434f4d00047573657200000001586beaae010012002053142f614ee6c39823710d9f31ff2984ed0bd9074d6e542e8468137f7b909c17000000430001000b4a544c414e2e434f2e554b000562696c6c7900000001586beaae1f00120020508dd2b209064e101bf209caef5fda236875706a5e9ad47c157db5907778785f" //nolint: lll
)

func TestAccDataSourceExternalAuthKerberos_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		ProviderFactories:         testAccProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				// Create the resource to be fetched by the datasource.
				Config: testAccDataSourceExternalAuthKerberosConfigCreate(),
			},
			{
				// Validate that the datasource correctly retrieves the resource.
				Config: testAccDataSourceExternalAuthKerberosConfigData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.wallix-bastion_externalauth_kerberos.testacc_dataExternalAuthKerberos",
						"authentication_name", "testacc_dataExternalAuthKerberos"),
					resource.TestCheckResourceAttr(
						"data.wallix-bastion_externalauth_kerberos.testacc_dataExternalAuthKerberos",
						"host", "192.168.100.20"),
					resource.TestCheckResourceAttr(
						"data.wallix-bastion_externalauth_kerberos.testacc_dataExternalAuthKerberos",
						"ker_dom_controller", "EXAMPLE.COM"),
					resource.TestCheckResourceAttr(
						"data.wallix-bastion_externalauth_kerberos.testacc_dataExternalAuthKerberos",
						"port", "188"),
					resource.TestCheckResourceAttr(
						"data.wallix-bastion_externalauth_kerberos.testacc_dataExternalAuthKerberos",
						"description", "testacc dataExternalAuthKerberos"),
					resource.TestCheckResourceAttr(
						"data.wallix-bastion_externalauth_kerberos.testacc_dataExternalAuthKerberos",
						"use_primary_auth_domain", "true"),
				),
			},
		},
	})
}

// Resource creation configuration.
func testAccDataSourceExternalAuthKerberosConfigCreate() string {
	k, _ := hex.DecodeString(dataSourceKeytabDataHexStr)
	os.WriteFile("/tmp/testacc_data_ds", k, 0o644) //nolint: all

	return `
data "wallix-bastion_version" "v_ds" {}
resource "wallix-bastion_externalauth_kerberos" "testacc_dataExternalAuthKerberos" {
  authentication_name     = "testacc_dataExternalAuthKerberos"
  host                    = "192.168.100.20"
  ker_dom_controller      = "EXAMPLE.COM"
  port                    = 188
  description             = "testacc dataExternalAuthKerberos"
  use_primary_auth_domain = true
  keytab                  = split(".", data.wallix-bastion_version.v_ds.wab_version)[0] == "8" ? "" : filebase64("/tmp/testacc_data_ds")
}
`
}

// Datasource configuration to retrieve the created resource.
func testAccDataSourceExternalAuthKerberosConfigData() string {
	return `
data "wallix-bastion_version" "v_ds" {}
resource "wallix-bastion_externalauth_kerberos" "testacc_dataExternalAuthKerberos" {
  authentication_name     = "testacc_dataExternalAuthKerberos"
  host                    = "192.168.100.20"
  ker_dom_controller      = "EXAMPLE.COM"
  port                    = 188
  description             = "testacc dataExternalAuthKerberos"
  use_primary_auth_domain = true
  keytab                  = split(".", data.wallix-bastion_version.v_ds.wab_version)[0] == "8" ? "" : filebase64("/tmp/testacc_data_ds")
}

data "wallix-bastion_externalauth_kerberos" "testacc_dataExternalAuthKerberos" {
  authentication_name = wallix-bastion_externalauth_kerberos.testacc_dataExternalAuthKerberos.authentication_name
}
`
}
