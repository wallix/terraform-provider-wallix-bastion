package bastion

import (
	"maps"
	"net/http"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

const authDomainOIDCType = "wallix-bastion_authdomain_oidc"

// fakeAuthDomains serves one domain of type typ, as API v3.12 returns it (Bastion 12.4.3).
func fakeAuthDomains(typ string) http.HandlerFunc {
	domain := map[string]any{
		"id": "ad1", "domain_name": "corp", "type": typ, "description": "", "is_default": false,
		"auth_domain_name": "corp.example.com", "external_auths": []string{"oidc"}, "secondary_auth": []string{},
		"default_language": "en", "default_email_domain": "example.com", "mappings": []any{}, "label": "Corp",
	}

	return func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/authdomains/"):
			found := []any{}
			if r.URL.Query().Get("q") == skDomainName+"=corp" {
				found = append(found, domain)
			}
			writeJSON(w, found)
		case r.Method == http.MethodGet:
			writeJSON(w, domain)
		default:
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

// An unset description, is_default or secondary_auth reads back as null, their zero value set by
// the configuration stays; import refuses another type.
func TestAuthDomainOIDC(t *testing.T) {
	server := newTestProviderServer(t, newTestAPIClient(t, fakeAuthDomains(typeOIDC)))
	typ := schemaType(t, server, authDomainOIDCType)
	values := map[string]interface{}{
		"id": "ad1", skDomainName: "corp", skAuthDomainName: "corp.example.com",
		skExternalAuths: []tftypes.Value{tftypes.NewValue(tftypes.String, "oidc")}, skDefaultLanguage: "en",
		skDefaultEmailDomain: "example.com", "label": "Corp",
	}
	resp, err := server.ReadResource(t.Context(), &tfprotov5.ReadResourceRequest{
		TypeName: authDomainOIDCType, CurrentState: dynamicValue(t, typ, values),
	})
	if err != nil || hasError(resp.Diagnostics) {
		t.Fatalf("read: %v %v", err, resp.Diagnostics)
	}
	state := stateAttributes(t, resp.NewState, typ)
	for _, attr := range []string{skDescription, skIsDefault, skSecondaryAuth} {
		if !state[attr].IsNull() {
			t.Errorf("%s = %v, want null", attr, state[attr])
		}
	}

	planned := maps.Clone(values)
	planned[skDescription], planned[skIsDefault], planned[skSecondaryAuth] = "", false, []tftypes.Value{}
	update, err := server.ApplyResourceChange(t.Context(), &tfprotov5.ApplyResourceChangeRequest{
		TypeName: authDomainOIDCType, PriorState: dynamicValue(t, typ, values),
		PlannedState: dynamicValue(t, typ, planned), Config: dynamicValue(t, typ, planned),
	})
	if err != nil || hasError(update.Diagnostics) {
		t.Fatalf("update: %v %v", err, update.Diagnostics)
	}
	state = stateAttributes(t, update.NewState, typ)
	for _, attr := range []string{skDescription, skIsDefault, skSecondaryAuth} {
		if want := stateAttributes(t, dynamicValue(t, typ, planned), typ)[attr]; !state[attr].Equal(want) {
			t.Errorf("%s = %v after the update, want the planned %v", attr, state[attr], want)
		}
	}

	for name, tc := range map[string]struct {
		domainType string
		wantErr    string
	}{
		"OIDC": {domainType: typeOIDC},
		"SAML": {domainType: "SAML", wantErr: "not OIDC"},
	} {
		t.Run("import "+name, func(t *testing.T) {
			server := newTestProviderServer(t, newTestAPIClient(t, fakeAuthDomains(tc.domainType)))
			resp, err := server.ImportResourceState(t.Context(), &tfprotov5.ImportResourceStateRequest{
				TypeName: authDomainOIDCType, ID: "corp",
			})
			if err != nil {
				t.Fatal(err)
			}
			got := make([]string, 0, len(resp.Diagnostics))
			for _, d := range resp.Diagnostics {
				got = append(got, d.Summary)
			}
			if hasError(resp.Diagnostics) != (tc.wantErr != "") || !strings.Contains(strings.Join(got, "\n"), tc.wantErr) {
				t.Errorf("import = %v, want error %q", got, tc.wantErr)
			}
		})
	}
}
