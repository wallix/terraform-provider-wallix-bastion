package bastion_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"golang.org/x/mod/semver"

	"github.com/wallix/terraform-provider-wallix-bastion/bastion"
)

// config/wsm doesn't exist before API v3.12; skip on older versions. Default (unset) is v3.12+.
func TestAccResourceConfigWSM_basic(t *testing.T) {
	if v := os.Getenv("WALLIX_BASTION_API_VERSION"); v == "" || semver.Compare(v, bastion.VersionWallixAPI312) >= 0 {
		resourceName := "wallix-bastion_config_wsm.test"

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { testAccPreCheck(t) },
			ProviderFactories: testAccProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccResourceConfigWSMCreate(),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(resourceName, "id"),
						resource.TestCheckResourceAttr(resourceName, "hostname", "wsm.testacc.example.com"),
						resource.TestCheckResourceAttrSet(resourceName, "jws_public"),
					),
				},
				{
					Config: testAccResourceConfigWSMUpdate(),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "hostname", "wsm2.testacc.example.com"),
					),
				},
				{
					ResourceName:      resourceName,
					ImportState:       true,
					ImportStateVerify: false,
					ImportStateId:     "wsmConfig",
				},
			},
			PreventPostDestroyRefresh: true, // Config always exists on the Bastion; nothing to destroy.
		})
	}
}

func testAccResourceConfigWSMCreate() string {
	return `
resource "wallix-bastion_config_wsm" "test" {
  hostname = "wsm.testacc.example.com"
}
`
}

func testAccResourceConfigWSMUpdate() string {
	return `
resource "wallix-bastion_config_wsm" "test" {
  hostname = "wsm2.testacc.example.com"
}
`
}
