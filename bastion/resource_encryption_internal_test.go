package bastion

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// encryptionMock is a Bastion whose GET /encryption returns sealState, and which records the PUT
// bodies.
type encryptionMock struct {
	mu        sync.Mutex
	sealState string
	puts      []string
}

func (mock *encryptionMock) client(t *testing.T) *Client {
	t.Helper()

	return newTestAPIClient(t, func(w http.ResponseWriter, r *http.Request) {
		mock.mu.Lock()
		defer mock.mu.Unlock()
		w.Header().Set(contentTypeHeader, contentTypeJSON)
		if r.URL.Path != "/api/"+VersionWallixAPI312+"/encryption" {
			w.WriteHeader(http.StatusNotFound)

			return
		}
		if r.Method == http.MethodPut {
			body, _ := io.ReadAll(r.Body)
			mock.puts = append(mock.puts, string(body))
			w.WriteHeader(http.StatusNoContent)

			return
		}
		// The field names of a Bastion 12.4.1, on API v3.8 and v3.12.
		fmt.Fprintf(w, `{"seal_state":%q,"encryption_mode":"unprotected"}`, mock.sealState)
	})
}

// The state was read from fields that don't exist ("encryption", "sealed_state") after a switch
// on versions without their "v" prefix, so the encryption was always reported ready.
func TestDataSourceEncryptionReadsTheSealState(t *testing.T) {
	for state, wantEnabled := range map[string]bool{
		encryptionUnsealed: true, encryptionSealed: false, encryptionNeedSetup: false,
	} {
		t.Run(state, func(t *testing.T) {
			mock := &encryptionMock{sealState: state}
			r := dataSourceEncryption()
			d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{})
			if diags := r.ReadContext(t.Context(), d, mock.client(t)); diags.HasError() {
				t.Fatalf("read: %v", diags)
			}
			if got := d.Get("enabled").(bool); got != wantEnabled {
				t.Errorf("enabled = %v, want %v", got, wantEnabled)
			}
			if got := d.Get("seal_state").(string); got != state {
				t.Errorf("seal_state = %q, want %q", got, state)
			}
			if got := d.Get("encryption_mode").(string); got != "unprotected" {
				t.Errorf("encryption_mode = %q, want unprotected", got)
			}
		})
	}
}

// A sealed Bastion must not be dropped from state: the next apply would send it a setup request.
func TestEncryptionReadOfASealedBastionFails(t *testing.T) {
	mock := &encryptionMock{sealState: encryptionSealed}
	r := resourceEncryption()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{"new_passphrase": "p"})
	d.SetId("encryption")
	diags := r.ReadContext(t.Context(), d, mock.client(t))
	if !diags.HasError() || !strings.Contains(fmt.Sprint(diags), "sealed") {
		t.Fatalf("diags = %v, want the sealed error", diags)
	}
	if d.Id() != "encryption" {
		t.Errorf("id = %q, want the encryption kept in state", d.Id())
	}

	mock.sealState = encryptionNeedSetup
	if diags := r.ReadContext(t.Context(), d, mock.client(t)); diags.HasError() || d.Id() != "" {
		t.Errorf("read of a Bastion to set up = (%v, id %q), want it dropped from state", diags, d.Id())
	}
}

// "Setup without a passphrase" is new_passphrase = "": it was dropped by omitempty, which sent {}.
// A Bastion already set up gets no setup request.
func TestEncryptionCreate(t *testing.T) {
	mock := &encryptionMock{sealState: encryptionNeedSetup}
	r := resourceEncryption()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{"new_passphrase": ""})
	if diags := r.CreateContext(t.Context(), d, mock.client(t)); diags.HasError() {
		t.Fatalf("create: %v", diags)
	}
	if len(mock.puts) != 1 {
		t.Fatalf("PUT bodies = %v, want the setup", mock.puts)
	}
	var body map[string]interface{}
	if err := json.Unmarshal([]byte(mock.puts[0]), &body); err != nil {
		t.Fatal(err)
	}
	if v, ok := body["new_passphrase"]; !ok || v != "" || len(body) != 1 {
		t.Errorf("setup body = %s, want {\"new_passphrase\":\"\"}", mock.puts[0])
	}

	mock = &encryptionMock{sealState: encryptionUnsealed}
	d = schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{"new_passphrase": "p"})
	diags := r.CreateContext(t.Context(), d, mock.client(t))
	if !diags.HasError() || !strings.Contains(fmt.Sprint(diags), "already set up") {
		t.Fatalf("diags = %v, want the already set up error", diags)
	}
	if len(mock.puts) != 0 {
		t.Errorf("PUT bodies = %v, want none on a Bastion already set up", mock.puts)
	}
}

// A change of the passphrase sends the current one and the new one.
func TestPrepareEncryptionJSONForAChange(t *testing.T) {
	r := resourceEncryption()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{
		"current_passphrase": "old-passphrase", "new_passphrase": "new-passphrase",
	})
	body, err := json.Marshal(prepareEncryptionJSON(d, true))
	if err != nil {
		t.Fatal(err)
	}
	if want := `{"passphrase":"old-passphrase","new_passphrase":"new-passphrase"}`; string(body) != want {
		t.Errorf("body = %s, want %s", body, want)
	}
}
