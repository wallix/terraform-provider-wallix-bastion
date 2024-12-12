package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAuthDomainAD_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				// Create the resource to be fetched by the datasource.
				Config: testAccDataSourceAuthDomainADConfigCreate(),
			},
			{
				// Validate that the datasource correctly retrieves the resource.
				Config: testAccDataSourceAuthDomainADConfigData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.wallix-bastion_authdomain.testacc_dataDomain",
						"domain_name", "testacc-domain"),
					resource.TestCheckResourceAttr("data.wallix-bastion_authdomain.testacc_dataDomain",
						"auth_domain_name", "testacc-auth-domain"),
					resource.TestCheckResourceAttr("data.wallix-bastion_authdomain.testacc_dataDomain",
						"default_language", "en"),
					resource.TestCheckResourceAttr("data.wallix-bastion_authdomain.testacc_dataDomain",
						"default_email_domain", "example.com"),
				),
			},
		},
	})
}

// Resource creation configuration.
func testAccDataSourceAuthDomainADConfigCreate() string {
	return `
resource "wallix-bastion_authdomain_ad" "testacc_dataAuthDomain" {
  domain_name         = "testacc-domain"
  auth_domain_name    = "testacc-auth-domain"
  default_language    = "en"
  default_email_domain = "example.com"
  external_auths      = ["auth1", "auth2"]
}
`
}

// Datasource configuration to retrieve the created resource.
func testAccDataSourceAuthDomainADConfigData() string {
	return `
resource "wallix-bastion_authdomain_ad" "testacc_dataAuthDomain" {
  domain_name         = "testacc-domain"
  auth_domain_name    = "testacc-auth-domain"
  default_language    = "en"
  default_email_domain = "example.com"
  external_auths      = ["auth1", "auth2"]
}

data "wallix-bastion_authdomain" "testacc_dataDomain" {
  domain_name = wallix-bastion_domain.testacc_dataDomain.domain_name
  auth_domain_name = wallix-bastion_domain.testacc_dataDomain.domain_real_name
}
`
}
