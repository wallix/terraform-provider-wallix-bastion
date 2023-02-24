package bastion_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

//nolint:wrapcheck
func TestAccDataSourceConfigoption_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfigoptionData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.wallix-bastion_configoption.global",
						"date"),
					resource.TestCheckResourceAttr("data.wallix-bastion_configoption.global",
						"name", "Global"),
					resource.TestCheckResourceAttr("data.wallix-bastion_configoption.global",
						"config_name", "wabengine"),
					resource.TestCheckResourceAttrWith("data.wallix-bastion_configoption.global",
						"options.0", func(value string) error {
							jsonData := make(map[string]interface{})
							if err := json.Unmarshal([]byte(value), &jsonData); err != nil {
								return err
							}
							if jsonData["name"].(string) != "one_time_password_ttl" {
								return fmt.Errorf("can't find name=one_time_password_ttl in json")
							}

							return nil
						},
					),
				),
			},
		},
		PreventPostDestroyRefresh: true,
	})
}

func testAccDataSourceConfigoptionData() string {
	return `
data "wallix-bastion_configoption" "global" {
  config_id = "wabengine"
  options_list = [
    "one_time_password_ttl",
  ]
}
`
}
