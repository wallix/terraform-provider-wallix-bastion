package bastion_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// The Bastion downloads the CRL as soon as crl_auto_fetch is set and rejects a URL it can't
// download, so the URL mode is only tested with the URL of a CRL that the test Bastion can
// download (resolvable and reachable from it, with a certificate it trusts for HTTPS).
const envTestCRLURL = "WALLIX_TEST_CRL_URL"

// testAccCRL returns an empty CRL in PEM format, signed by a throwaway CA. The Bastion accepts it
// with crl_check = false: the CRL doesn't have to come from a CA it knows.
func testAccCRL(t *testing.T, number int64) string {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generating the CA key: %v", err)
	}
	now := time.Now()
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "testacc-crl-ca"},
		NotBefore:             now.Add(-time.Hour),
		NotAfter:              now.Add(48 * time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		SubjectKeyId:          []byte{1, 2, 3, 4},
	}
	caDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("creating the CA certificate: %v", err)
	}
	ca, err := x509.ParseCertificate(caDER)
	if err != nil {
		t.Fatalf("parsing the CA certificate: %v", err)
	}
	crlDER, err := x509.CreateRevocationList(rand.Reader, &x509.RevocationList{
		Number:     big.NewInt(number),
		ThisUpdate: now.Add(-time.Hour),
		NextUpdate: now.Add(24 * time.Hour),
	}, ca, key)
	if err != nil {
		t.Fatalf("creating the CRL: %v", err)
	}

	return string(pem.EncodeToMemory(&pem.Block{Type: "X509 CRL", Bytes: crlDER}))
}

func TestTestAccCRLIsValid(t *testing.T) {
	block, _ := pem.Decode([]byte(testAccCRL(t, 1)))
	if block == nil || block.Type != "X509 CRL" {
		t.Fatal("testAccCRL doesn't return a PEM X509 CRL")
	}
	crl, err := x509.ParseRevocationList(block.Bytes)
	if err != nil {
		t.Fatalf("parsing the CRL: %v", err)
	}
	if !crl.NextUpdate.After(time.Now()) {
		t.Errorf("CRL nextUpdate %s is already past", crl.NextUpdate)
	}
}

// crl_check stays false in every step so that the test CRLs never reject the X509
// authentications of the test Bastion. The test leaves the CRL configuration of the Bastion
// changed (a test CRL uploaded, no URL): run it on a dedicated Bastion.
func TestAccResourceConfigX509CRL_basic(t *testing.T) {
	resourceName := "wallix-bastion_config_x509_crl.test"
	firstCRL := testAccCRL(t, 1)
	secondCRL := testAccCRL(t, 2)

	var steps []resource.TestStep
	if crlURL := os.Getenv(envTestCRLURL); crlURL != "" {
		// URL first: the next step then checks that switching to a file clears the URL.
		steps = append(steps, resource.TestStep{
			Config: testAccResourceConfigX509CRLURL(crlURL),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(resourceName, "id", "x509CRLConfig"),
				resource.TestCheckResourceAttr(resourceName, "crl_check", "false"),
				resource.TestCheckResourceAttr(resourceName, "crl_auto_fetch", crlURL),
			),
		})
	} else {
		t.Logf("%s isn't set: crl_auto_fetch and the switch from a URL to a file aren't tested", envTestCRLURL)
	}
	steps = append(steps,
		resource.TestStep{
			Config: testAccResourceConfigX509CRLFile(firstCRL),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(resourceName, "id", "x509CRLConfig"),
				resource.TestCheckResourceAttr(resourceName, "crl_check", "false"),
				resource.TestCheckResourceAttr(resourceName, "crl_auto_fetch", ""),
				resource.TestCheckResourceAttr(resourceName, "crl_file", firstCRL),
			),
		},
		resource.TestStep{
			Config: testAccResourceConfigX509CRLFile(secondCRL),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(resourceName, "crl_file", secondCRL),
			),
		},
		resource.TestStep{
			ResourceName:      resourceName,
			ImportState:       true,
			ImportStateVerify: true,
			ImportStateId:     "x509CRLConfig",
			// The API never returns the CRL file.
			ImportStateVerifyIgnore: []string{"crl_file"},
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories:  testAccProtoV5ProviderFactories,
		Steps:                     steps,
		PreventPostDestroyRefresh: true, // Config always exists on the Bastion; nothing to destroy.
	})
}

func testAccResourceConfigX509CRLURL(crlURL string) string {
	return fmt.Sprintf(`
resource "wallix-bastion_config_x509_crl" "test" {
  crl_check      = false
  crl_auto_fetch = %q
}
`, crlURL)
}

func testAccResourceConfigX509CRLFile(crl string) string {
	return fmt.Sprintf(`
resource "wallix-bastion_config_x509_crl" "test" {
  crl_check = false
  crl_file  = <<EOT
%sEOT
}
`, crl)
}
