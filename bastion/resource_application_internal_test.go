package bastion

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestFillApplicationHandlesNilLocalDomains(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceApplication().Schema, map[string]interface{}{})

	fillApplication(d, jsonApplication{
		ApplicationName:  "test-app",
		ConnectionPolicy: skProtoRDP,
		Category:         skStandard,
	})

	localDomains := d.Get("local_domains")
	if localDomains == nil {
		t.Fatalf("expected local_domains to be initialized")
	}
}
