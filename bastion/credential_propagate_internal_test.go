package bastion

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
	testPropagateDomain   = "dom"
	testPropagateAccount  = "acc"
	testPropagatePassword = "Secret-1"
	testStateTrue         = "true"
	testPrivateKey        = "-----BEGIN OPENSSH PRIVATE KEY-----\ntest\n-----END OPENSSH PRIVATE KEY-----"
)

// propagateMock is a Bastion API with the account "acc" of the domain "dom", which records the
// requests, answers the propagation with propagateStatus and the creation of a credential with
// the id "cred".
type propagateMock struct {
	mu              sync.Mutex
	requests        []string
	propagateStatus int
}

func (mock *propagateMock) client(t *testing.T) *Client {
	t.Helper()

	return newTestAPIClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		mock.mu.Lock()
		mock.requests = append(mock.requests, r.Method+" "+strings.TrimPrefix(r.URL.Path, "/api/"+VersionWallixAPI312)+
			" "+strings.TrimSpace(string(body)))
		mock.mu.Unlock()
		const credentials = "/api/" + VersionWallixAPI312 + "/domains/dom/accounts/acc/credentials/"
		w.Header().Set(contentTypeHeader, contentTypeJSON)
		switch {
		case strings.HasPrefix(r.URL.Path, "/api/"+VersionWallixAPI312+"/accountchangepassword/"):
			w.WriteHeader(mock.propagateStatus)
			if mock.propagateStatus != http.StatusNoContent {
				fmt.Fprint(w, `{"error":"Service Unavailable","description":"Failed to change the password"}`)
			}
		case r.Method == http.MethodPost && r.URL.Path == credentials:
			w.Header().Set("X-Object-Id", "cred")
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodGet && r.URL.Path == credentials:
			fmt.Fprint(w, `[]`)
		case r.Method == http.MethodGet && r.URL.Path == credentials+"cred":
			fmt.Fprint(w, `{"id":"cred","type":"password"}`)
		case r.Method == http.MethodGet && r.URL.Path == "/api/"+VersionWallixAPI312+"/domains/dom":
			fmt.Fprint(w, `{"id":"dom","domain_name":"dom"}`)
		case r.Method == http.MethodGet && r.URL.Path == "/api/"+VersionWallixAPI312+"/domains/dom/accounts/acc":
			fmt.Fprint(w, `{"id":"acc","account_name":"acc"}`)
		default:
			w.WriteHeader(http.StatusNoContent)
		}
	})
}

// propagations returns the requests sent to /accountchangepassword.
func (mock *propagateMock) propagations() []string {
	mock.mu.Lock()
	defer mock.mu.Unlock()
	var found []string
	for _, request := range mock.requests {
		if strings.Contains(request, "/accountchangepassword/") {
			found = append(found, request)
		}
	}

	return found
}

func credentialRaw(extra map[string]interface{}) map[string]interface{} {
	raw := map[string]interface{}{
		skDomainID: testPropagateDomain, skAccountID: testPropagateAccount, skPropagateCredentialChange: true,
	}
	for k, v := range extra {
		raw[k] = v
	}

	return raw
}

// Per the API, an empty body on /accountchangepassword makes the Bastion generate a new value and
// change it on the target host. The creation always sends the configured value, to the endpoint
// of the credential type: the SSH key propagation used to send {} to /password, which rotated the
// password of the account on the target host.
func TestCredentialCreatePropagatesTheConfiguredValue(t *testing.T) {
	for name, tc := range map[string]struct {
		raw  map[string]interface{}
		want string
	}{
		"password": {
			raw:  credentialRaw(map[string]interface{}{skType: skPassword, skPassword: testPropagatePassword}),
			want: `PUT /accountchangepassword/acc/password {"password":"Secret-1"}`,
		},
		"ssh_key": {
			raw: credentialRaw(map[string]interface{}{
				skType: "ssh_key", skPrivateKey: testPrivateKey, skPassphrase: "phrase",
			}),
			want: `PUT /accountchangepassword/acc/ssh_key {"private_key":"` +
				strings.ReplaceAll(testPrivateKey, "\n", `\n`) + `","passphrase":"phrase"}`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			mock := &propagateMock{propagateStatus: http.StatusNoContent}
			r := resourceDomainAccountCredential()
			d := schema.TestResourceDataRaw(t, r.Schema, tc.raw)
			if diags := r.CreateContext(t.Context(), d, mock.client(t)); diags.HasError() {
				t.Fatalf("create: %v", diags)
			}
			if got := mock.propagations(); len(got) != 1 || got[0] != tc.want {
				t.Errorf("propagation requests = %q, want [%q]", got, tc.want)
			}
		})
	}
}

// When the propagation fails after the credential was created, the credential stays in state, so
// that Terraform saves it as tainted and replaces it, instead of leaving it on the Bastion out of
// state (the next creation then failed with "already exists").
func TestCredentialCreateKeepsTheCredentialWhenThePropagationFails(t *testing.T) {
	mock := &propagateMock{propagateStatus: http.StatusServiceUnavailable}
	r := resourceDomainAccountCredential()
	d := schema.TestResourceDataRaw(t, r.Schema,
		credentialRaw(map[string]interface{}{skType: skPassword, skPassword: testPropagatePassword}))
	if diags := r.CreateContext(t.Context(), d, mock.client(t)); !diags.HasError() {
		t.Fatal("create succeeded, want the propagation error")
	}
	if d.Id() != "cred" {
		t.Errorf("id = %q, want the created credential kept in state", d.Id())
	}
}

func credentialUpdateData(
	t *testing.T, state map[string]string, config map[string]interface{},
) *schema.ResourceData {
	t.Helper()
	attributes := map[string]string{
		"id": "cred", skDomainID: testPropagateDomain, skAccountID: testPropagateAccount, skType: skPassword,
	}
	for k, v := range state {
		attributes[k] = v
	}
	prior := &terraform.InstanceState{ID: "cred", Attributes: attributes}
	sm := schema.InternalMap(resourceDomainAccountCredential().Schema)
	raw := map[string]interface{}{skDomainID: testPropagateDomain, skAccountID: testPropagateAccount, skType: skPassword}
	for k, v := range config {
		raw[k] = v
	}
	diff, err := sm.Diff(t.Context(), prior, terraform.NewResourceConfigRaw(raw), nil, nil, true)
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	d, err := sm.Data(prior, diff)
	if err != nil {
		t.Fatalf("data: %v", err)
	}

	return d
}

// Only a new value is propagated: enabling propagate_credential_change alone used to send the
// password again to the target host. A new password is sent with its value.
func TestCredentialUpdatePropagatesOnlyANewValue(t *testing.T) {
	for name, tc := range map[string]struct {
		state map[string]string
		want  []string
	}{
		"propagation enabled, same password": {
			state: map[string]string{skPassword: testPropagatePassword, skPropagateCredentialChange: "false"},
		},
		"new password": {
			state: map[string]string{skPassword: "Secret-0", skPropagateCredentialChange: testStateTrue},
			want:  []string{`PUT /accountchangepassword/acc/password {"password":"Secret-1"}`},
		},
	} {
		t.Run(name, func(t *testing.T) {
			mock := &propagateMock{propagateStatus: http.StatusNoContent}
			d := credentialUpdateData(t, tc.state,
				map[string]interface{}{skPassword: testPropagatePassword, skPropagateCredentialChange: true})
			if err := updateDomainAccountCredential(t.Context(), d, mock.client(t)); err != nil {
				t.Fatalf("update: %v", err)
			}
			mock.mu.Lock()
			requests := mock.requests
			mock.mu.Unlock()
			if got := mock.propagations(); fmt.Sprint(got) != fmt.Sprint(tc.want) || len(requests) != len(tc.want) {
				t.Errorf("requests = %q, want %q", requests, tc.want)
			}
		})
	}
}

// A propagation without the value would let the Bastion generate one and change it on the target
// host: it is refused at plan time.
func TestCredentialPropagateRequiresTheValue(t *testing.T) {
	for name, tc := range map[string]struct {
		raw     map[string]interface{}
		wantErr bool
	}{
		"password missing": {
			raw: credentialRaw(map[string]interface{}{skType: skPassword}), wantErr: true,
		},
		"generated ssh key": {
			raw: credentialRaw(map[string]interface{}{skType: "ssh_key", skPrivateKey: "generate:RSA_4096"}), wantErr: true,
		},
		"password set": {
			raw: credentialRaw(map[string]interface{}{skType: skPassword, skPassword: testPropagatePassword}),
		},
		"ssh key set": {
			raw: credentialRaw(map[string]interface{}{skType: "ssh_key", skPrivateKey: testPrivateKey}),
		},
		"no propagation": {
			raw: map[string]interface{}{skDomainID: testPropagateDomain, skAccountID: testPropagateAccount, skType: skPassword},
		},
	} {
		t.Run(name, func(t *testing.T) {
			r := resourceDomainAccountCredential()
			_, err := r.Diff(t.Context(), nil, terraform.NewResourceConfigRaw(tc.raw), nil)
			if (err != nil) != tc.wantErr {
				t.Errorf("plan error = %v, want an error: %v", err, tc.wantErr)
			}
		})
	}
}
