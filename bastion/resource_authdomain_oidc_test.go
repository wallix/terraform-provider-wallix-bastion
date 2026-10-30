package bastion_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceAuthDomainOIDC_basic(t *testing.T) {
	testAccSkipBeforeOIDC(t)
	resourceName := "wallix-bastion_authdomain_oidc.testacc_AuthDomainOIDC"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAuthDomainOIDCConfig("testacc.AuthDomainOIDC", ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "external_auths.0", "testacc_AuthDomainOIDC"),
					resource.TestCheckNoResourceAttr(resourceName, "description"),
					resource.TestCheckNoResourceAttr(resourceName, "is_default"),
				),
			},
			{
				Config: testAccResourceAuthDomainOIDCConfig("testacc.AuthDomainOIDC-u", `
  description = "testacc OIDC domain"
`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "domain_name", "testacc.AuthDomainOIDC-u"),
					resource.TestCheckResourceAttr(resourceName, "description", "testacc OIDC domain"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     "testacc.AuthDomainOIDC-u",
				ImportStateVerify: true,
			},
		},
	})
}

func testAccResourceAuthDomainOIDCConfig(domainName, extra string) string {
	return fmt.Sprintf(`
resource "wallix-bastion_externalauth_oidc" "testacc_AuthDomainOIDC" {
  authentication_name = "testacc_AuthDomainOIDC"
  timeout             = 30
  client_id           = "testacc"
  claim_customization {
    username = "preferred_username"
  }
}
resource "wallix-bastion_authdomain_oidc" "testacc_AuthDomainOIDC" {
  domain_name          = %q
  auth_domain_name     = "testacc-oidc.example.com"
  external_auths       = [wallix-bastion_externalauth_oidc.testacc_AuthDomainOIDC.authentication_name]
  default_email_domain = "example.com"
  default_language     = "en"
  label                = "testacc OIDC"
%s}
`, domainName, extra)
}
