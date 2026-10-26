package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceTimeframe_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceTimeframeCreate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"wallix-bastion_timeframe.testacc_Timeframe",
						"id"),
				),
			},
			{
				Config: testAccResourceTimeframeUpdate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"wallix-bastion_timeframe.testacc_Timeframe",
						"description", "testacc Timeframe"),
					resource.TestCheckResourceAttr(
						"wallix-bastion_timeframe.testacc_Timeframe",
						"periods.#", "2"),
					resource.TestCheckResourceAttrSet(
						"wallix-bastion_timeframe.testacc_Timeframe2",
						"id"),
				),
			},
			{
				ResourceName:      "wallix-bastion_timeframe.testacc_Timeframe",
				ImportState:       true,
				ImportStateId:     "testacc_Timeframe",
				ImportStateVerify: true,
			},
		},
		PreventPostDestroyRefresh: true,
	})
}

func testAccResourceTimeframeCreate() string {
	return `
resource "wallix-bastion_timeframe" "testacc_Timeframe" {
  timeframe_name = "testacc_Timeframe"
  periods {
    start_date = "2020-01-01"
    end_date   = "2020-02-02"
    start_time = "08:00"
    end_time   = "12:00"
    week_days  = ["monday"]
  }
}
resource "wallix-bastion_timeframe" "testacc_Timeframe2" {
  timeframe_name = "testacc_Timeframe2"
}
`
}

func testAccResourceTimeframeUpdate() string {
	return `
resource "wallix-bastion_timeframe" "testacc_Timeframe" {
  timeframe_name = "testacc_Timeframe"
  description    = "testacc Timeframe"
  is_overtimable = true
  periods {
    start_date = "2020-01-01"
    end_date   = "2020-02-02"
    start_time = "08:00"
    end_time   = "12:00"
    week_days  = ["monday"]
  }
  periods {
    start_date = "2020-02-01"
    end_date   = "2020-03-02"
    start_time = "10:00"
    end_time   = "16:00"
    week_days  = ["monday", "friday"]
  }
}
resource "wallix-bastion_timeframe" "testacc_Timeframe2" {
  timeframe_name = "testacc_Timeframe2"
}
`
}

// Renaming a timeframe dropped it from state ("Root object was present, but now absent") and
// left it on the Bastion under its new name.
func TestAccResourceTimeframe_rename(t *testing.T) {
	resourceName := "wallix-bastion_timeframe.testacc_TimeframeRename"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceTimeframeRename("testacc_TimeframeRename_a"),
			},
			{
				Config: testAccResourceTimeframeRename("testacc_TimeframeRename_b"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", "testacc_TimeframeRename_b"),
					resource.TestCheckResourceAttr(resourceName, "timeframe_name", "testacc_TimeframeRename_b"),
				),
			},
		},
	})
}

func testAccResourceTimeframeRename(name string) string {
	return `
resource "wallix-bastion_timeframe" "testacc_TimeframeRename" {
  timeframe_name = "` + name + `"
  periods {
    start_date = "2020-01-01"
    end_date   = "2020-02-02"
    start_time = "08:00"
    end_time   = "12:00"
    week_days  = ["monday"]
  }
}
`
}

// A state written by 0.16.0, the last release on SDKv2, gives an empty plan once the resource
// runs on the Framework: SDKv2 stored "" and false for the arguments left unset.
func TestAccResourceTimeframe_upgradeFromSDKv2(t *testing.T) {
	// The muxed provider is served as wallix/wallix-bastion, the address the state of 0.16.0 holds.
	t.Setenv(resource.EnvTfAccProviderNamespace, "wallix")
	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"wallix-bastion": {Source: "wallix/wallix-bastion", VersionConstraint: "0.16.0"},
				},
				Config: testAccResourceTimeframeUpgrade(),
			},
			{
				ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
				Config:                   testAccResourceTimeframeUpgradeSource + testAccResourceTimeframeUpgrade(),
				PlanOnly:                 true,
			},
			{
				ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
				Config:                   testAccResourceTimeframeUpgradeSource + testAccResourceTimeframeUpgrade(),
			},
		},
	})
}

const testAccResourceTimeframeUpgradeSource = `
terraform {
  required_providers {
    wallix-bastion = {
      source = "wallix/wallix-bastion"
    }
  }
}
`

func testAccResourceTimeframeUpgrade() string {
	return `
resource "wallix-bastion_timeframe" "testacc_TimeframeUpgrade" {
  timeframe_name = "testacc_TimeframeUpgrade"
  periods {
    start_date = "2020-01-01"
    end_date   = "2020-02-02"
    start_time = "08:00"
    end_time   = "12:00"
    week_days  = ["monday", "friday"]
  }
}
resource "wallix-bastion_timeframe" "testacc_TimeframeUpgrade2" {
  timeframe_name = "testacc_TimeframeUpgrade2"
  description    = "testacc Timeframe"
  is_overtimable = true
}
data "wallix-bastion_timeframe" "testacc_TimeframeUpgrade" {
  timeframe_name = wallix-bastion_timeframe.testacc_TimeframeUpgrade.timeframe_name
}
`
}
