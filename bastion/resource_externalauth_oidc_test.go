package bastion_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/wallix/terraform-provider-wallix-bastion/bastion"
)

// testAccSkipBeforeOIDC skips a test of the OIDC resources, which came with API v3.12.
func testAccSkipBeforeOIDC(t *testing.T) {
	t.Helper()
	if os.Getenv("WALLIX_BASTION_API_VERSION") == bastion.VersionWallixAPI38 {
		t.Skip("OIDC requires API v3.12")
	}
}

func TestAccResourceExternalAuthOIDC_basic(t *testing.T) {
	testAccSkipBeforeOIDC(t)
	resourceName := "wallix-bastion_externalauth_oidc.testacc_ExternalAuthOIDC"
	var firstID string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceExternalAuthOIDCConfig("testacc_ExternalAuthOIDC", "s1", 1, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "timeout", "12.5"),
					resource.TestCheckResourceAttr(resourceName, "claim_customization.group", "groups"),
					resource.TestCheckNoResourceAttr(resourceName, "claim_customization.email"),
					resource.TestCheckNoResourceAttr(resourceName, "client_secret_wo"),
					resource.TestCheckResourceAttr(resourceName, "client_secret_wo_version", "1"),
					resource.TestMatchResourceAttr(resourceName, "redirect_uri", regexp.MustCompile(`^https://.+/api/oidc$`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateId:           "testacc_ExternalAuthOIDC",
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"client_secret_wo_version"},
			},
			{
				// Rename, new secret, advanced attributes.
				Config: testAccResourceExternalAuthOIDCConfig("testacc_ExternalAuthOIDC-u", "s2", 2,
					testAccResourceExternalAuthOIDCAdvanced),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "authentication_name", "testacc_ExternalAuthOIDC-u"),
					resource.TestCheckResourceAttr(resourceName, "client_secret_wo_version", "2"),
					resource.TestCheckResourceAttr(resourceName, "oidc_advanced_attributes.grant_type", "authorization_code"),
					resource.TestCheckResourceAttr(resourceName, "oidc_advanced_attributes.scope.#", "2"),
					testAccSaveID(resourceName, &firstID),
				),
			},
			{
				// The Bastion cannot remove the advanced attributes: replaced.
				Config: testAccResourceExternalAuthOIDCConfig("testacc_ExternalAuthOIDC-u", "s2", 2, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(resourceName, "oidc_advanced_attributes.issuer"),
					testAccCheckIDChanged(resourceName, &firstID),
				),
			},
		},
	})
}

const testAccResourceExternalAuthOIDCAdvanced = `
  oidc_advanced_attributes {
    issuer                                = "https://idp.example.com"
    scope                                 = ["openid", "profile"]
    authorization_endpoint                = "https://idp.example.com/authorize"
    token_endpoint                        = "https://idp.example.com/token"
    jwks_uri                              = "https://idp.example.com/jwks"
    id_token_signing_alg_values_supported = ["RS256"]
  }
`

func testAccResourceExternalAuthOIDCConfig(name, secret string, version int, advanced string) string {
	return fmt.Sprintf(`
resource "wallix-bastion_externalauth_oidc" "testacc_ExternalAuthOIDC" {
  authentication_name      = %q
  description              = "testacc OIDC"
  timeout                  = 12.5
  client_id                = "testacc"
  client_secret_wo         = %q
  client_secret_wo_version = %d
  discovery_url            = "https://idp.example.com/.well-known/openid-configuration"
  claim_customization {
    username = "preferred_username"
    group    = "groups"
  }
%s}
`, name, secret, version, advanced)
}

// testAccSaveID stores the id of name in id.
func testAccSaveID(name string, id *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("%s not found in state", name)
		}
		*id = rs.Primary.ID

		return nil
	}
}

// testAccCheckIDChanged fails when name still has the id saved by testAccSaveID.
func testAccCheckIDChanged(name string, id *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("%s not found in state", name)
		}
		if rs.Primary.ID == *id {
			return fmt.Errorf("%s kept the id %s, want a replacement", name, *id)
		}

		return nil
	}
}
