package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// ocsp_check stays false: enabling it with an unreachable responder could reject the X509
// authentications of the test Bastion.
func TestAccResourceConfigX509OCSP_basic(t *testing.T) {
	resourceName := "wallix-bastion_config_x509_ocsp.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceConfigX509OCSPCreate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", "x509OCSPConfig"),
					resource.TestCheckResourceAttr(resourceName, "ocsp_check", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_ocsp_responder",
						"http://testacc-ocsp.example.com"),
					resource.TestCheckResourceAttr(resourceName, "override_ocsp_responder", "false"),
					resource.TestCheckResourceAttr(resourceName, "ocsp_proxy", ""),
				),
			},
			{
				Config: testAccResourceConfigX509OCSPUpdate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "override_ocsp_responder", "true"),
					resource.TestCheckResourceAttr(resourceName, "ocsp_proxy",
						"http://testacc-proxy.example.com:3128"),
				),
			},
			{
				// Arguments removed from the configuration are reset on the Bastion.
				Config: testAccResourceConfigX509OCSPCreate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "override_ocsp_responder", "false"),
					resource.TestCheckResourceAttr(resourceName, "ocsp_proxy", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "x509OCSPConfig",
			},
		},
		PreventPostDestroyRefresh: true, // Config always exists on the Bastion; nothing to destroy.
	})
}

func testAccResourceConfigX509OCSPCreate() string {
	return `
resource "wallix-bastion_config_x509_ocsp" "test" {
  ocsp_check             = false
  default_ocsp_responder = "http://testacc-ocsp.example.com"
}
`
}

func testAccResourceConfigX509OCSPUpdate() string {
	return `
resource "wallix-bastion_config_x509_ocsp" "test" {
  ocsp_check              = false
  default_ocsp_responder  = "http://testacc-ocsp.example.com"
  override_ocsp_responder = true
  ocsp_proxy              = "http://testacc-proxy.example.com:3128"
}
`
}
