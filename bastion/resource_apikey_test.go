package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceAPIKey_basic(t *testing.T) {
	var key string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAPIKeyCreate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"wallix-bastion_apikey.testacc_APIKey",
						"id"),
					testAccCheckAPIKey("wallix-bastion_apikey.testacc_APIKey", &key, apiKeyNew),
				),
			},
			{
				Config: testAccResourceAPIKeyUpdate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"wallix-bastion_apikey.testacc_APIKey",
						"ip_limitation", "10.10.11.22,10.10.33.44"),
					testAccCheckAPIKey("wallix-bastion_apikey.testacc_APIKey", &key, apiKeySame),
				),
			},
			{
				RefreshState: true,
				Check:        testAccCheckAPIKey("wallix-bastion_apikey.testacc_APIKey", &key, apiKeySame),
			},
			{
				ResourceName:            "wallix-bastion_apikey.testacc_APIKey",
				ImportState:             true,
				ImportStateId:           "testacc_APIKey",
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apikey"},
				ImportStateCheck:        testAccCheckImportedAPIKeyEmpty,
			},
		},
		PreventPostDestroyRefresh: true,
	})
}

func testAccResourceAPIKeyCreate() string {
	return `
resource "wallix-bastion_apikey" "testacc_APIKey" {
  apikey_name = "testacc_APIKey"
}
`
}

func testAccResourceAPIKeyUpdate() string {
	return `
resource "wallix-bastion_apikey" "testacc_APIKey" {
  apikey_name   = "testacc_APIKey"
  ip_limitation = "10.10.11.22,10.10.33.44"
}
`
}
