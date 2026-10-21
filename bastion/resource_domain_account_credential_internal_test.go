package bastion

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// readRotatedDomainCredential reads the credential "old" of the account "acc" from an API that
// answers 404 for it, and lists the credentials of the account with listStatus and listBody.
func readRotatedDomainCredential(t *testing.T, listStatus int, listBody string) *schema.ResourceData {
	t.Helper()
	const credentials = "/api/" + VersionWallixAPI312 + "/domains/dom/accounts/acc/credentials/"
	client := newTestAPIClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(contentTypeHeader, contentTypeJSON)
		switch r.URL.Path {
		case credentials:
			w.WriteHeader(listStatus)
			fmt.Fprint(w, listBody)
		case credentials + "rotated":
			fmt.Fprint(w, `{"id":"rotated","type":"password"}`)
		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"error":"Resource not found"}`)
		}
	})

	r := resourceDomainAccountCredential()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{
		skDomainID:  "dom",
		skAccountID: "acc",
		skType:      "password",
		skPassword:  "secret",
	})
	d.SetId("old")
	if diags := r.ReadContext(t.Context(), d, client); diags.HasError() {
		t.Fatalf("read: %v", diags)
	}

	return d
}

// A rotation recreates the credential with a new id: the read must follow it instead of dropping
// it from state, which made the next apply fail with "already exists".
func TestDomainAccountCredentialReadFollowsRotation(t *testing.T) {
	d := readRotatedDomainCredential(t, http.StatusOK, `[{"id":"rotated","type":"password"}]`)
	if d.Id() != "rotated" {
		t.Errorf("id = %q, want the id of the rotated credential", d.Id())
	}
}

func TestDomainAccountCredentialReadDropsDeletedCredential(t *testing.T) {
	d := readRotatedDomainCredential(t, http.StatusOK, `[]`)
	if d.Id() != "" {
		t.Errorf("id = %q, want the credential dropped from state", d.Id())
	}
}

// When the account was deleted, listing its credentials returns 404: the credential is gone too,
// and the read must not fail.
func TestDomainAccountCredentialReadDropsCredentialOfDeletedAccount(t *testing.T) {
	d := readRotatedDomainCredential(t, http.StatusNotFound, `{"error":"Resource not found"}`)
	if d.Id() != "" {
		t.Errorf("id = %q, want the credential dropped from state", d.Id())
	}
}
