package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceDomain_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceDomainConfigCreate(),
			},
			{
				Config: testAccDataSourceDomainConfigData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.wallix-bastion_domain.testacc_dataDomain",
						"domain_real_name", "testacc-domain.local"),
				),
			},
		},
		PreventPostDestroyRefresh: true,
	})
}

func testAccDataSourceDomainConfigCreate() string {
	return `
resource "wallix-bastion_domain" "testacc_dataDomain" {
  domain_name      = "testacc-domain"
  domain_real_name = "testacc-domain.local"
}
`
}

func testAccDataSourceDomainConfigData() string {
	return `
resource "wallix-bastion_domain" "testacc_dataDomain" {
  domain_name      = "testacc-domain"
  domain_real_name = "testacc-domain.local"
}

data "wallix-bastion_domain" "testacc_dataDomain" {
  domain_name = "testacc-domain"
}
`
}
