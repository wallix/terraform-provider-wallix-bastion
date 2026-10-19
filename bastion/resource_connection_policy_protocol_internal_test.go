package bastion

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// protocol is only sent on creation: changing it was a silent no-op update with a permanent diff.
func TestConnectionPolicyProtocolChangeReplacesThePolicy(t *testing.T) {
	const nameAttr = "connection_policy_name"
	state := &terraform.InstanceState{ID: "policy-id", Attributes: map[string]string{
		"id":       "policy-id",
		nameAttr:   "policy",
		skProtocol: skProtoSSH,
		skType:     skProtoSSH,
	}}
	config := terraform.NewResourceConfigRaw(map[string]interface{}{
		nameAttr:   "policy",
		skProtocol: "RAWTCPIP",
	})

	diff, err := schema.InternalMap(resourceConnectionPolicy().Schema).Diff(t.Context(), state, config, nil, nil, true)
	if err != nil {
		t.Fatalf("unexpected error computing the diff: %v", err)
	}
	if !diff.RequiresNew() {
		t.Error("changing protocol must replace the connection policy")
	}
}
