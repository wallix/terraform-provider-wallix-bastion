package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// The encryption resource is a singleton with no natural lookup key: the data source just
// reports whether encryption is enabled/ready on the Bastion (see data_source_encryption.go).
// We can't control that state from this test, so we only assert the attribute is readable.
func TestAccDataSourceEncryption_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		ProviderFactories:         testAccProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceEncryptionConfigData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.wallix-bastion_encryption.current", "enabled"),
				),
			},
		},
	})
}

func testAccDataSourceEncryptionConfigData() string {
	return `
data "wallix-bastion_encryption" "current" {}
`
}
