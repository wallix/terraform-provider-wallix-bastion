package bastion

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
	propagateTestDomain  = "dom"
	propagateTestAccount = "acc"
)

// credentialPropagateMock is a Bastion with the account "acc" of the domain "dom". It records the
// writes, and answers propagateCode to the propagation requests.
type credentialPropagateMock struct {
	mu            sync.Mutex
	credType      string
	created       bool
	propagateCode int
	writes        []string
}

func (mock *credentialPropagateMock) serveHTTP(w http.ResponseWriter, r *http.Request) {
	mock.mu.Lock()
	defer mock.mu.Unlock()
	body, _ := io.ReadAll(r.Body)
	path := strings.TrimPrefix(r.URL.Path, "/api/"+VersionWallixAPI312)
	if r.Method != http.MethodGet {
		mock.writes = append(mock.writes, r.Method+" "+path+" "+string(body))
	}
	w.Header().Set(contentTypeHeader, contentTypeJSON)
	switch {
	case r.Method == http.MethodGet && path == "/domains/dom":
		fmt.Fprint(w, `{"id":"dom","domain_name":"dom"}`)
	case r.Method == http.MethodGet && path == "/domains/dom/accounts/acc":
		fmt.Fprint(w, `{"id":"acc","account_name":"acc"}`)
	case r.Method == http.MethodGet && path == "/domains/dom/accounts/acc/credentials/":
		if mock.created {
			fmt.Fprintf(w, `[{"id":"cred","type":%q}]`, mock.credType)
		} else {
			fmt.Fprint(w, `[]`)
		}
	case r.Method == http.MethodGet && path == "/domains/dom/accounts/acc/credentials/cred":
		fmt.Fprintf(w, `{"id":"cred","type":%q,"public_key":""}`, mock.credType)
	case r.Method == http.MethodPost && path == "/domains/dom/accounts/acc/credentials/":
		mock.created = true
		w.Header().Set("X-Object-Id", "cred")
		w.WriteHeader(http.StatusNoContent)
	case r.Method == http.MethodPut && strings.HasPrefix(path, "/accountchangepassword/"):
		w.WriteHeader(mock.propagateCode)
		if mock.propagateCode != http.StatusNoContent {
			fmt.Fprint(w, `{"error":"Service Unavailable","description":"Network timeout"}`)
		}
	case r.Method == http.MethodPut && path == "/domains/dom/accounts/acc/credentials/cred":
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (mock *credentialPropagateMock) recordedWrites() []string {
	mock.mu.Lock()
	defer mock.mu.Unlock()

	return append([]string(nil), mock.writes...)
}

// applyDomainAccountCredential plans then applies the configuration from a prior state (nil for a
// creation), as Terraform does.
func applyDomainAccountCredential(
	t *testing.T, mock *credentialPropagateMock, prior map[string]string, config map[string]interface{},
) (*terraform.InstanceState, diag.Diagnostics, error) {
	t.Helper()
	client := newTestAPIClient(t, mock.serveHTTP)
	r := resourceDomainAccountCredential()
	var state *terraform.InstanceState
	if prior != nil {
		state = &terraform.InstanceState{ID: "cred", Attributes: prior}
		state.Attributes["id"] = "cred"
	}
	config[skDomainID] = propagateTestDomain
	config[skAccountID] = propagateTestAccount
	instanceDiff, err := r.Diff(t.Context(), state, terraform.NewResourceConfigRaw(config), client)
	if err != nil {
		return nil, nil, fmt.Errorf("diff: %w", err)
	}
	newState, diags := r.Apply(t.Context(), state, instanceDiff, client)

	return newState, diags, nil
}

func propagateBody(t *testing.T, write string) map[string]string {
	t.Helper()
	parts := strings.SplitN(write, " ", 3)
	var body map[string]string
	if err := json.Unmarshal([]byte(parts[2]), &body); err != nil {
		t.Fatalf("propagation body of %q: %v", write, err)
	}

	return body
}

// An empty propagation body makes the Bastion generate a new password or SSH key and set it on
// the target host (checked on a Bastion 12.4.1): the secret of the configuration must be sent, on
// the route of the credential type.
func TestDomainAccountCredentialPropagatesTheSecret(t *testing.T) {
	for name, tc := range map[string]struct {
		config    map[string]interface{}
		wantRoute string
		wantBody  map[string]string
	}{
		skPassword: {
			config:    map[string]interface{}{skType: skPassword, skPassword: "New-Pass1"},
			wantRoute: "PUT /accountchangepassword/acc/password",
			wantBody:  map[string]string{skPassword: "New-Pass1"},
		},
		"ssh_key": {
			config:    map[string]interface{}{skType: skSSHKey, skPrivateKey: "KEY", skPassphrase: "phrase"},
			wantRoute: "PUT /accountchangepassword/acc/ssh_key",
			wantBody:  map[string]string{skPrivateKey: "KEY", skPassphrase: "phrase"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			mock := &credentialPropagateMock{credType: tc.config[skType].(string), propagateCode: http.StatusNoContent}
			tc.config[skPropagateCredentialChange] = true
			_, diags, err := applyDomainAccountCredential(t, mock, nil, tc.config)
			if err != nil || diags.HasError() {
				t.Fatalf("create: %v %v", err, diags)
			}
			writes := mock.recordedWrites()
			if len(writes) != 2 {
				t.Fatalf("writes = %v, want the creation then the propagation", writes)
			}
			if create := propagateBody(t, writes[0]); create[skType] != tc.config[skType] {
				t.Errorf("creation body = %v, want the credential with its type", create)
			}
			if !strings.HasPrefix(writes[1], tc.wantRoute+" ") {
				t.Errorf("propagation = %q, want %s", writes[1], tc.wantRoute)
			}
			got := propagateBody(t, writes[1])
			if len(got) != len(tc.wantBody) {
				t.Errorf("propagation body = %v, want %v", got, tc.wantBody)
			}
			for k, v := range tc.wantBody {
				if got[k] != v {
					t.Errorf("propagation body = %v, want %v", got, tc.wantBody)
				}
			}
		})
	}
}

// Without the secret, the propagation would rotate the credential on the target, and a generated
// key can't be propagated: refused at plan.
func TestDomainAccountCredentialPropagateRequiresTheSecret(t *testing.T) {
	for name, config := range map[string]map[string]interface{}{
		skPassword:  {skType: skPassword},
		"ssh_key":   {skType: skSSHKey},
		"generated": {skType: skSSHKey, skPrivateKey: "generate:RSA_4096"},
	} {
		t.Run(name, func(t *testing.T) {
			mock := &credentialPropagateMock{credType: config[skType].(string), propagateCode: http.StatusNoContent}
			config[skPropagateCredentialChange] = true
			_, _, err := applyDomainAccountCredential(t, mock, nil, config)
			if err == nil || !strings.Contains(err.Error(), "propagate_credential_change = true") {
				t.Fatalf("plan error = %v, want the propagation refused", err)
			}
			if writes := mock.recordedWrites(); len(writes) != 0 {
				t.Errorf("writes = %v, want none", writes)
			}
		})
	}
}

// The credential is created before it is propagated: when the propagation fails, it must stay in
// state (tainted) instead of being left on the Bastion out of state.
func TestDomainAccountCredentialFailedPropagationKeepsTheCredential(t *testing.T) {
	mock := &credentialPropagateMock{credType: skPassword, propagateCode: http.StatusServiceUnavailable}
	state, diags, err := applyDomainAccountCredential(t, mock, nil, map[string]interface{}{
		skType: skPassword, skPassword: "New-Pass1", skPropagateCredentialChange: true,
	})
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	if !diags.HasError() || !strings.Contains(fmt.Sprint(diags), "Network timeout") {
		t.Fatalf("diags = %v, want the propagation error", diags)
	}
	if state == nil || state.ID != "cred" {
		t.Fatalf("state = %v, want the created credential kept", state)
	}
}

// Enabling the propagation on an existing credential changes nothing on the target: only a new
// secret is propagated.
func TestDomainAccountCredentialPropagatesOnlyANewSecret(t *testing.T) {
	prior := map[string]string{
		skDomainID: propagateTestDomain, skAccountID: propagateTestAccount, skType: skPassword, skPassword: "Pass1",
		skPropagateCredentialChange: strconv.FormatBool(false),
	}
	mock := &credentialPropagateMock{credType: skPassword, created: true, propagateCode: http.StatusNoContent}
	if _, diags, err := applyDomainAccountCredential(t, mock, prior, map[string]interface{}{
		skType: skPassword, skPassword: "Pass1", skPropagateCredentialChange: true,
	}); err != nil || diags.HasError() {
		t.Fatalf("update: %v %v", err, diags)
	}
	if writes := mock.recordedWrites(); len(writes) != 0 {
		t.Errorf("writes = %v, want none when only propagate_credential_change changes", writes)
	}

	prior[skPropagateCredentialChange] = strconv.FormatBool(true)
	if _, diags, err := applyDomainAccountCredential(t, mock, prior, map[string]interface{}{
		skType: skPassword, skPassword: "Pass2", skPropagateCredentialChange: true,
	}); err != nil || diags.HasError() {
		t.Fatalf("update: %v %v", err, diags)
	}
	writes := mock.recordedWrites()
	if len(writes) != 1 || !strings.HasPrefix(writes[0], "PUT /accountchangepassword/acc/password ") ||
		propagateBody(t, writes[0])[skPassword] != "Pass2" {
		t.Errorf("writes = %v, want the propagation of the new password", writes)
	}
}
