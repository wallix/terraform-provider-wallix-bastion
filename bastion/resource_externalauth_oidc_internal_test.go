package bastion

import (
	"encoding/json"
	"maps"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

const (
	externalAuthOIDCType = "wallix-bastion_externalauth_oidc"
	oidcRedirectURI      = "https://bastion.example.com/api/oidc"
	oidcFirstID          = "oidca"
	oidcSecretVersion    = "client_secret_wo_version"
	oidcRedirect         = "redirect_uri"
)

// writeJSON writes v as the response of a fake Bastion.
func writeJSON(w http.ResponseWriter, v any) {
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// fakeOIDCBastion serves /externalauths/ as API v3.12 does for OIDC (Bastion 12.4.3): PUT merges
// the top-level keys and refuses redirect_uri, GET masks the secret.
type fakeOIDCBastion struct {
	mu      sync.Mutex
	objects map[string]map[string]any
	secrets map[string]string
}

func newFakeOIDCBastion() *fakeOIDCBastion {
	return &fakeOIDCBastion{objects: map[string]map[string]any{}, secrets: map[string]string{}}
}

func (f *fakeOIDCBastion) handler(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	id := strings.TrimPrefix(r.URL.Path, "/api/"+VersionWallixAPI312+"/externalauths/")
	var body map[string]any
	_ = json.NewDecoder(r.Body).Decode(&body)
	switch {
	case r.Method == http.MethodPost:
		id = "oidc" + string(rune('a'+len(f.objects)))
		f.secrets[id], _ = body["client_secret"].(string)
		delete(body, "client_secret")
		if body[oidcRedirect] == nil {
			body[oidcRedirect] = oidcRedirectURI
		}
		body["id"] = id
		f.objects[id] = body
		w.Header().Set("X-Object-Id", id)
		w.WriteHeader(http.StatusNoContent)
	case r.Method == http.MethodGet && id == "":
		name := strings.TrimPrefix(r.URL.Query().Get("q"), skAuthenticationName+"=")
		list := []map[string]any{}
		for _, o := range f.objects {
			if o[skAuthenticationName] == name {
				list = append(list, o)
			}
		}
		writeJSON(w, list)
	case f.objects[id] == nil:
		w.WriteHeader(http.StatusNotFound)
	case r.Method == http.MethodGet:
		o := maps.Clone(f.objects[id])
		o["client_secret"] = ""
		if f.secrets[id] != "" {
			o["client_secret"] = "********"
		}
		writeJSON(w, o)
	case r.Method == http.MethodPut:
		if _, ok := body[oidcRedirect]; ok {
			http.Error(w, `{"error":"Bad Request","description":"'redirect_uri' was unexpected"}`, http.StatusBadRequest)

			return
		}
		if s, ok := body["client_secret"].(string); ok {
			f.secrets[id] = s
			delete(body, "client_secret")
		}
		maps.Copy(f.objects[id], body)
		w.WriteHeader(http.StatusNoContent)
	case r.Method == http.MethodDelete:
		delete(f.objects, id)
		w.WriteHeader(http.StatusNoContent)
	}
}

// oidcValues returns the attributes of an external auth named name with its required claims.
func oidcValues(t *testing.T, typ tftypes.Object, name string, extra map[string]interface{}) map[string]interface{} {
	t.Helper()
	claims, ok := typ.AttributeTypes["claim_customization"].(tftypes.Object)
	if !ok {
		t.Fatal("claim_customization is not an object")
	}
	values := map[string]interface{}{
		skAuthenticationName: name, skTimeout: 30, "client_id": "cid",
		"claim_customization": objectValue(claims, map[string]interface{}{"username": "sub", "group": "groups"}),
	}
	maps.Copy(values, extra)

	return values
}

func applyOIDC(
	t *testing.T, server tfprotov5.ProviderServer, typ tftypes.Object, prior, planned, config map[string]interface{},
) *tfprotov5.ApplyResourceChangeResponse {
	t.Helper()
	priorState := nullValue(t, typ)
	if prior != nil {
		priorState = dynamicValue(t, typ, prior)
	}
	resp, err := server.ApplyResourceChange(t.Context(), &tfprotov5.ApplyResourceChangeRequest{
		TypeName: externalAuthOIDCType, PriorState: priorState,
		PlannedState: dynamicValue(t, typ, planned), Config: dynamicValue(t, typ, config),
	})
	if err != nil || hasError(resp.Diagnostics) {
		t.Fatalf("apply: %v %v", err, resp.Diagnostics)
	}

	return resp
}

// The secret goes from the configuration to the Bastion and never into the state, a new version
// sends a new secret, and removing the secret and its version removes it from the Bastion.
func TestExternalAuthOIDCWriteOnlySecret(t *testing.T) {
	fake := newFakeOIDCBastion()
	server := newTestProviderServer(t, newTestAPIClient(t, fake.handler))
	typ := schemaType(t, server, externalAuthOIDCType)
	withSecret := map[string]interface{}{"client_secret_wo": "s1", oidcSecretVersion: 1}
	planned := oidcValues(t, typ, "oidc", map[string]interface{}{
		"id": tftypes.UnknownValue, oidcRedirect: tftypes.UnknownValue, oidcSecretVersion: 1,
	})

	resp := applyOIDC(t, server, typ, nil, planned, oidcValues(t, typ, "oidc", withSecret))
	state := stateAttributes(t, resp.NewState, typ)
	if !state["client_secret_wo"].IsNull() {
		t.Errorf("client_secret_wo = %v in state, want null", state["client_secret_wo"])
	}
	if fake.secrets[oidcFirstID] != "s1" {
		t.Errorf("secret on the Bastion = %q, want s1", fake.secrets[oidcFirstID])
	}

	prior := oidcValues(t, typ, "oidc", map[string]interface{}{
		"id": oidcFirstID, oidcRedirect: oidcRedirectURI, oidcSecretVersion: 1,
	})
	rotated := oidcValues(t, typ, "oidc", map[string]interface{}{
		"id": oidcFirstID, oidcRedirect: oidcRedirectURI, oidcSecretVersion: 2,
	})
	applyOIDC(t, server, typ, prior, rotated, oidcValues(t, typ, "oidc",
		map[string]interface{}{"client_secret_wo": "s2", oidcSecretVersion: 2}))
	if fake.secrets[oidcFirstID] != "s2" {
		t.Errorf("secret on the Bastion = %q after the rotation, want s2", fake.secrets[oidcFirstID])
	}

	cleared := oidcValues(t, typ, "oidc", map[string]interface{}{
		"id": oidcFirstID, oidcRedirect: oidcRedirectURI,
	})
	applyOIDC(t, server, typ, rotated, cleared, oidcValues(t, typ, "oidc", nil))
	if fake.secrets[oidcFirstID] != "" {
		t.Errorf("secret on the Bastion = %q after removing it, want none", fake.secrets[oidcFirstID])
	}
}

// A secret removed outside Terraform nulls the version, so the next plan sends it again.
func TestExternalAuthOIDCSecretRemovedOutside(t *testing.T) {
	fake := newFakeOIDCBastion()
	server := newTestProviderServer(t, newTestAPIClient(t, fake.handler))
	typ := schemaType(t, server, externalAuthOIDCType)
	planned := oidcValues(t, typ, "oidc", map[string]interface{}{
		"id": tftypes.UnknownValue, oidcRedirect: tftypes.UnknownValue, oidcSecretVersion: 1,
	})
	resp := applyOIDC(t, server, typ, nil, planned, oidcValues(t, typ, "oidc",
		map[string]interface{}{"client_secret_wo": "s1", oidcSecretVersion: 1}))
	fake.secrets[oidcFirstID] = ""

	read, err := server.ReadResource(t.Context(), &tfprotov5.ReadResourceRequest{
		TypeName: externalAuthOIDCType, CurrentState: resp.NewState,
	})
	if err != nil || hasError(read.Diagnostics) {
		t.Fatalf("read: %v %v", err, read.Diagnostics)
	}
	if v := stateAttributes(t, read.NewState, typ)[oidcSecretVersion]; !v.IsNull() {
		t.Errorf("client_secret_wo_version = %v, want null", v)
	}
}

// Optional values the API returns empty read back as null, "" set by the configuration stays "",
// and an update never sends redirect_uri.
func TestExternalAuthOIDCReadBackAndUpdate(t *testing.T) {
	fake := newFakeOIDCBastion()
	server := newTestProviderServer(t, newTestAPIClient(t, fake.handler))
	typ := schemaType(t, server, externalAuthOIDCType)
	planned := oidcValues(t, typ, "oidc", map[string]interface{}{
		"id": tftypes.UnknownValue, oidcRedirect: tftypes.UnknownValue,
	})
	resp := applyOIDC(t, server, typ, nil, planned, oidcValues(t, typ, "oidc", nil))
	state := stateAttributes(t, resp.NewState, typ)
	for _, attr := range []string{skDescription, "discovery_url"} {
		if !state[attr].IsNull() {
			t.Errorf("%s = %v, want null", attr, state[attr])
		}
	}
	if !state[oidcRedirect].Equal(tftypes.NewValue(tftypes.String, oidcRedirectURI)) {
		t.Errorf("redirect_uri = %v, want the one computed by the Bastion", state[oidcRedirect])
	}

	prior := oidcValues(t, typ, "oidc", map[string]interface{}{
		"id": oidcFirstID, oidcRedirect: oidcRedirectURI,
	})
	claims, ok := typ.AttributeTypes["claim_customization"].(tftypes.Object)
	if !ok {
		t.Fatal("claim_customization is not an object")
	}
	zeros := map[string]interface{}{
		skDescription: "", "discovery_url": "",
		"claim_customization": objectValue(claims, map[string]interface{}{skUsername: "sub", skGroup: ""}),
	}
	config := oidcValues(t, typ, "oidc2", zeros)
	zeros["id"], zeros[oidcRedirect] = oidcFirstID, oidcRedirectURI
	renamed := oidcValues(t, typ, "oidc2", zeros)
	resp = applyOIDC(t, server, typ, prior, renamed, config)
	// Terraform rejects a result that differs from the plan; ApplyResourceChange doesn't check it.
	want := stateAttributes(t, dynamicValue(t, typ, renamed), typ)
	for name, got := range stateAttributes(t, resp.NewState, typ) {
		if !got.Equal(want[name]) {
			t.Errorf("%s = %v after the update, want the planned %v", name, got, want[name])
		}
	}
}

// grant_type and response_type default to the only values the API supports, inside the block only;
// removing the block, which the Bastion cannot do, replaces the external auth.
func TestExternalAuthOIDCPlanAdvancedAttributes(t *testing.T) {
	server := newTestProviderServer(t, newTestAPIClient(t, newFakeOIDCBastion().handler))
	typ := schemaType(t, server, externalAuthOIDCType)
	advType, ok := typ.AttributeTypes["oidc_advanced_attributes"].(tftypes.Object)
	if !ok {
		t.Fatal("oidc_advanced_attributes is not an object")
	}
	strs := func(s ...string) []tftypes.Value {
		v := make([]tftypes.Value, len(s))
		for i := range s {
			v[i] = tftypes.NewValue(tftypes.String, s[i])
		}

		return v
	}
	advanced := objectValue(advType, map[string]interface{}{
		"issuer": "https://idp", "scope": strs("openid"), "authorization_endpoint": "https://idp/a",
		"token_endpoint": "https://idp/t", "jwks_uri": "https://idp/j",
		"id_token_signing_alg_values_supported": strs("RS256"),
	})
	plan := func(prior, config map[string]interface{}) *tfprotov5.PlanResourceChangeResponse {
		t.Helper()
		priorState := nullValue(t, typ)
		if prior != nil {
			priorState = dynamicValue(t, typ, prior)
		}
		resp, err := server.PlanResourceChange(t.Context(), &tfprotov5.PlanResourceChangeRequest{
			TypeName: externalAuthOIDCType, PriorState: priorState,
			ProposedNewState: dynamicValue(t, typ, config), Config: dynamicValue(t, typ, config),
		})
		if err != nil || hasError(resp.Diagnostics) {
			t.Fatalf("plan: %v %v", err, resp.Diagnostics)
		}

		return resp
	}

	withoutBlock := stateAttributes(t, plan(nil, oidcValues(t, typ, "oidc", nil)).PlannedState, typ)
	if adv := withoutBlock["oidc_advanced_attributes"]; !adv.IsNull() {
		t.Errorf("oidc_advanced_attributes = %v without the block, want null", adv)
	}
	planned := stateAttributes(t, plan(nil, oidcValues(t, typ, "oidc",
		map[string]interface{}{"oidc_advanced_attributes": advanced})).PlannedState, typ)
	var adv map[string]tftypes.Value
	if err := planned["oidc_advanced_attributes"].As(&adv); err != nil {
		t.Fatal(err)
	}
	if !adv["grant_type"].Equal(tftypes.NewValue(tftypes.String, "authorization_code")) ||
		!adv["response_type"].Equal(tftypes.NewValue(tftypes.String, "code")) {
		t.Errorf("grant_type = %v, response_type = %v, want the defaults", adv["grant_type"], adv["response_type"])
	}

	withBlock := objectValue(advType, map[string]interface{}{
		"issuer": "https://idp", "scope": strs("openid"), "authorization_endpoint": "https://idp/a",
		"token_endpoint": "https://idp/t", "jwks_uri": "https://idp/j", "grant_type": "authorization_code",
		"response_type": "code", "id_token_signing_alg_values_supported": strs("RS256"),
	})
	prior := oidcValues(t, typ, "oidc", map[string]interface{}{
		"id": oidcFirstID, oidcRedirect: oidcRedirectURI, "oidc_advanced_attributes": withBlock,
	})
	config := oidcValues(t, typ, "oidc", map[string]interface{}{
		"id": oidcFirstID, oidcRedirect: oidcRedirectURI,
	})
	resp := plan(prior, config)
	if len(resp.RequiresReplace) != 1 || resp.RequiresReplace[0].String() != `AttributeName("oidc_advanced_attributes")` {
		t.Errorf("requires replace = %v, want oidc_advanced_attributes", resp.RequiresReplace)
	}
}

// The fields the API requires in oidc_advanced_attributes are only required with the block.
func TestExternalAuthOIDCValidate(t *testing.T) {
	server := newTestProviderServer(t, newTestAPIClient(t, newFakeOIDCBastion().handler))
	typ := schemaType(t, server, externalAuthOIDCType)
	advType, ok := typ.AttributeTypes["oidc_advanced_attributes"].(tftypes.Object)
	if !ok {
		t.Fatal("oidc_advanced_attributes is not an object")
	}
	for name, tc := range map[string]struct {
		extra   map[string]interface{}
		wantErr string
	}{
		"without the block": {},
		"block without issuer": {
			extra: map[string]interface{}{
				"oidc_advanced_attributes": objectValue(advType, map[string]interface{}{"jwks_uri": "j"}),
			},
			wantErr: "issuer",
		},
		"empty redirect_uri": {
			extra:   map[string]interface{}{oidcRedirect: ""},
			wantErr: "redirect_uri",
		},
		"version without secret": {
			extra:   map[string]interface{}{oidcSecretVersion: 1},
			wantErr: "client_secret_wo",
		},
	} {
		t.Run(name, func(t *testing.T) {
			resp, err := server.ValidateResourceTypeConfig(t.Context(), &tfprotov5.ValidateResourceTypeConfigRequest{
				TypeName: externalAuthOIDCType, Config: dynamicValue(t, typ, oidcValues(t, typ, "oidc", tc.extra)),
				ClientCapabilities: &tfprotov5.ValidateResourceTypeConfigClientCapabilities{WriteOnlyAttributesAllowed: true},
			})
			if err != nil {
				t.Fatal(err)
			}
			got := make([]string, 0, len(resp.Diagnostics))
			for _, d := range resp.Diagnostics {
				got = append(got, d.Summary+": "+d.Detail)
			}
			if hasError(resp.Diagnostics) != (tc.wantErr != "") || !strings.Contains(strings.Join(got, "\n"), tc.wantErr) {
				t.Errorf("validate = %v, want error %q", got, tc.wantErr)
			}
		})
	}
}

func TestExternalAuthOIDCVersionCheck(t *testing.T) {
	if err := oidcVersionCheck("resource", externalAuthOIDCType, VersionWallixAPI38); err == nil {
		t.Error("v3.8: want an error, OIDC came with v3.12")
	}
	if err := oidcVersionCheck("resource", externalAuthOIDCType, VersionWallixAPI312); err != nil {
		t.Errorf("v3.12: %v", err)
	}
}
