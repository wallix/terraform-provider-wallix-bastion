package bastion_test

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"golang.org/x/mod/semver"

	"github.com/wallix/terraform-provider-wallix-bastion/bastion"
)

type apiKeyChange int

const (
	apiKeyNew apiKeyChange = iota
	apiKeySame
	apiKeyRotated
)

// testAccCheckAPIKey checks that apikey holds a key and not the ******** of GET, compares it with the
// previous one and records it, without printing it.
func testAccCheckAPIKey(name string, key *string, change apiKeyChange) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("%s not found in the state", name)
		}
		v := rs.Primary.Attributes["apikey"]
		switch {
		case v == "":
			return errors.New("apikey is empty")
		case v == "********":
			return errors.New("apikey is the masked value returned by GET")
		case change == apiKeySame && v != *key:
			return errors.New("apikey changed, want the key stored at creation")
		case change == apiKeyRotated && v == *key:
			return errors.New("apikey unchanged, want a new key")
		}
		*key = v

		return nil
	}
}

// The key can't be read back from the API: an imported key has an empty apikey.
func testAccCheckImportedAPIKeyEmpty(states []*terraform.InstanceState) error {
	if len(states) != 1 {
		return fmt.Errorf("%d imported states, want 1", len(states))
	}
	if states[0].Attributes["apikey"] != "" {
		return errors.New("apikey is set after import, want empty")
	}

	return nil
}

func testAccAPIKeyV2Supported() bool {
	v := os.Getenv("WALLIX_BASTION_API_VERSION")

	return v == "" || semver.Compare(v, bastion.VersionWallixAPI312) >= 0
}

// apikeys-v2 doesn't exist before API v3.12; skip on older versions. Default (unset) is v3.12+.
func TestAccResourceAPIKeyV2_basic(t *testing.T) {
	if !testAccAPIKeyV2Supported() {
		return
	}
	var key string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAPIKeyV2Create(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"wallix-bastion_apikey_v2.testacc_APIKeyV2",
						"id"),
					resource.TestCheckResourceAttr(
						"wallix-bastion_apikey_v2.testacc_APIKeyV2",
						"profile", "user"),
					testAccCheckAPIKey("wallix-bastion_apikey_v2.testacc_APIKeyV2", &key, apiKeyNew),
				),
			},
			{
				Config: testAccResourceAPIKeyV2Update(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"wallix-bastion_apikey_v2.testacc_APIKeyV2",
						"description", "testacc updated description"),
					testAccCheckAPIKey("wallix-bastion_apikey_v2.testacc_APIKeyV2", &key, apiKeySame),
				),
			},
			{
				RefreshState: true,
				Check:        testAccCheckAPIKey("wallix-bastion_apikey_v2.testacc_APIKeyV2", &key, apiKeySame),
			},
			{
				ResourceName:            "wallix-bastion_apikey_v2.testacc_APIKeyV2",
				ImportState:             true,
				ImportStateId:           "testacc_APIKeyV2",
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apikey"},
				ImportStateCheck:        testAccCheckImportedAPIKeyEmpty,
			},
		},
		PreventPostDestroyRefresh: true,
	})
}

// A key needs a read-only profile; the Bastion refuses any change to one, so a change replaces the
// profile, and replace_triggered_by the key bound to it.
func TestAccResourceAPIKeyV2_readOnlyProfile(t *testing.T) {
	if !testAccAPIKeyV2Supported() {
		return
	}
	var key, profileID string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceAPIKeyV2ReadOnlyProfile("", "testacc"),
				ExpectError: regexp.MustCompile(`editable = false`),
			},
			{
				Config: testAccResourceAPIKeyV2ReadOnlyProfile("editable = false", "testacc"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"wallix-bastion_profile.testacc_APIKeyV2RO", "editable", "false"),
					testAccCheckAPIKey("wallix-bastion_apikey_v2.testacc_APIKeyV2RO", &key, apiKeyNew),
					testAccRecordID("wallix-bastion_profile.testacc_APIKeyV2RO", &profileID, false),
				),
			},
			{
				Config: testAccResourceAPIKeyV2ReadOnlyProfile("editable = false", "testacc changed"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"wallix-bastion_profile.testacc_APIKeyV2RO", "description", "testacc changed"),
					testAccCheckAPIKey("wallix-bastion_apikey_v2.testacc_APIKeyV2RO", &key, apiKeyRotated),
					testAccRecordID("wallix-bastion_profile.testacc_APIKeyV2RO", &profileID, true),
				),
			},
		},
	})
}

// testAccRecordID records the id of a resource; with replaced, it must differ from the recorded one.
func testAccRecordID(name string, id *string, replaced bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("%s not found in the state", name)
		}
		if replaced && rs.Primary.ID == *id {
			return fmt.Errorf("%s has the same id %s, want a replacement", name, *id)
		}
		*id = rs.Primary.ID

		return nil
	}
}

func testAccResourceAPIKeyV2Create() string {
	return `
resource "wallix-bastion_apikey_v2" "testacc_APIKeyV2" {
  apikey_name = "testacc_APIKeyV2"
  profile     = "user"
}
`
}

func testAccResourceAPIKeyV2Update() string {
	return `
resource "wallix-bastion_apikey_v2" "testacc_APIKeyV2" {
  apikey_name = "testacc_APIKeyV2"
  profile     = "user"
  description = "testacc updated description"
}
`
}

func testAccResourceAPIKeyV2ReadOnlyProfile(editable, description string) string {
	return fmt.Sprintf(`
resource "wallix-bastion_profile" "testacc_APIKeyV2RO" {
  profile_name = "testacc_APIKeyV2RO"
  description  = %q
  %s
  gui_features {
    wab_audit = "view"
  }
  gui_transmission {}
}

resource "wallix-bastion_apikey_v2" "testacc_APIKeyV2RO" {
  apikey_name = "testacc_APIKeyV2RO"
  profile     = wallix-bastion_profile.testacc_APIKeyV2RO.profile_name

  lifecycle {
    replace_triggered_by = [wallix-bastion_profile.testacc_APIKeyV2RO]
  }
}
`, description, editable)
}
