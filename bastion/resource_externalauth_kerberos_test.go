// nolint: lll,nolintlint
package bastion_test

import (
	"encoding/hex"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	keytabDataHexStr = "0502000000320001000b4558414d504c452e434f4d00047573657200000001586aa82d01001700100c61039f010b2fbb88fe449fbf262477000000420001000b4558414d504c452e434f4d00047573657200000001586aa82d010012002053142f614ee6c39823710d9f31ff2984ed0bd9074d6e542e8468137f7b909c17000000320001000b4558414d504c452e434f4d00047573657200000001586beaad01001700100c61039f010b2fbb88fe449fbf262477000000420001000b4558414d504c452e434f4d00047573657200000001586beaae010012002053142f614ee6c39823710d9f31ff2984ed0bd9074d6e542e8468137f7b909c17000000430001000b4a544c414e2e434f2e554b000562696c6c7900000001586beaae1f00120020508dd2b209064e101bf209caef5fda236875706a5e9ad47c157db5907778785f" //nolint: lll
)

func TestAccResourceExternalAuthKerberos_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceExternalAuthKerberosCreate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"wallix-bastion_externalauth_kerberos.testacc_ExternalAuthKerberos",
						"id"),
				),
			},
			{
				Config: testAccResourceExternalAuthKerberosUpdate(),
			},
			{
				ResourceName:  "wallix-bastion_externalauth_kerberos.testacc_ExternalAuthKerberos",
				ImportState:   true,
				ImportStateId: "testacc_ExternalAuthKerberos",
			},
			{
				Config: testAccResourceExternalAuthKerberosCreate2(),
			},
			{
				Config: testAccResourceExternalAuthKerberosUpdate2(),
			},
		},
		PreventPostDestroyRefresh: true,
	})
}

func testAccResourceExternalAuthKerberosCreate() string {
	k, _ := hex.DecodeString(keytabDataHexStr)
	os.WriteFile("/tmp/testacc_data", k, 0644) //nolint: all

	return `
data "wallix-bastion_version" "v" {}
resource "wallix-bastion_externalauth_kerberos" "testacc_ExternalAuthKerberos" {
  authentication_name = "testacc_ExternalAuthKerberos"
  host                = "server1"
  ker_dom_controller  = "EXAMPLE.COM"
  port                = 88
  keytab              = split(".", data.wallix-bastion_version.v.wab_version)[0] == "8" ? "" : filebase64("/tmp/testacc_data")
}
`
}

func testAccResourceExternalAuthKerberosUpdate() string {
	return `
data "wallix-bastion_version" "v" {}
resource "wallix-bastion_externalauth_kerberos" "testacc_ExternalAuthKerberos" {
  authentication_name     = "testacc_ExternalAuthKerberos"
  host                    = "server1"
  ker_dom_controller      = "EXAMPLE.COM"
  port                    = 188
  description             = "testacc ExternalAuthKerberos"
  use_primary_auth_domain = true
  keytab                  = split(".", data.wallix-bastion_version.v.wab_version)[0] == "8" ? "" : filebase64("/tmp/testacc_data")
}
`
}

func testAccResourceExternalAuthKerberosCreate2() string {
	return `
data "wallix-bastion_version" "v" {}
resource "wallix-bastion_externalauth_kerberos" "testacc_ExternalAuthKerberosPassword" {
  authentication_name = "testacc_ExternalAuthKerberosPassword"
  host                = "server2"
  ker_dom_controller  = "EXAMPLE.COM"
  kerberos_password   = true
  port                = 88
  keytab              = split(".", data.wallix-bastion_version.v.wab_version)[0] == "8" ? "" : filebase64("/tmp/testacc_data")
}
`
}

func testAccResourceExternalAuthKerberosUpdate2() string {
	return `
data "wallix-bastion_version" "v" {}
resource "wallix-bastion_externalauth_kerberos" "testacc_ExternalAuthKerberosPassword" {
  authentication_name     = "testacc_ExternalAuthKerberosPassword"
  host                    = "server2"
  ker_dom_controller      = "EXAMPLE.COM"
  kerberos_password       = true
  port                    = 188
  description             = "testacc ExternalAuthKerberosPassword"
  use_primary_auth_domain = true
  keytab                  = split(".", data.wallix-bastion_version.v.wab_version)[0] == "8" ? "" : filebase64("/tmp/testacc_data")
}
`
}
