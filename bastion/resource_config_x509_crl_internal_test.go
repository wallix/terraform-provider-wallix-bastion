package bastion

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
	testCRLURL = "https://pki.example.com/crl.pem"
	testCRLPEM = "-----BEGIN X509 CRL-----\n" +
		"MIIB3DCBxQIBATANBgkqhkiG9w0BAQUFADCBgjELMAkGA1UEBhMCRlIx\n" +
		"-----END X509 CRL-----\n"
)

func TestConfigX509CRLCreateSendsAutoFetchForm(t *testing.T) {
	mock := newMockSingletonConfig(t, configX509CRLPath,
		map[string]interface{}{skCRLAutoFetch: "", skCRLCheck: false}, skCRLFile)
	r := resourceConfigX509CRL()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{
		skCRLCheck:     true,
		skCRLAutoFetch: testCRLURL,
	})

	if diags := r.CreateContext(t.Context(), d, mock.client(t)); diags.HasError() {
		t.Fatalf("create: %v", diags)
	}

	want := []map[string]interface{}{{skCRLAutoFetch: testCRLURL, skCRLCheck: true}}
	if got := mock.putBodies(); !reflect.DeepEqual(got, want) {
		t.Errorf("PUT bodies = %v, want %v", got, want)
	}
	if d.Id() != configX509CRLID {
		t.Errorf("id = %q, want x509CRLConfig", d.Id())
	}
	if got := d.Get(skCRLAutoFetch).(string); got != testCRLURL {
		t.Errorf("crl_auto_fetch = %q, want %q", got, testCRLURL)
	}
}

func TestConfigX509CRLCreateSendsFileForm(t *testing.T) {
	mock := newMockSingletonConfig(t, configX509CRLPath,
		map[string]interface{}{skCRLAutoFetch: "", skCRLCheck: false}, skCRLFile)
	r := resourceConfigX509CRL()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{
		skCRLCheck: true,
		skCRLFile:  testCRLPEM,
	})

	if diags := r.CreateContext(t.Context(), d, mock.client(t)); diags.HasError() {
		t.Fatalf("create: %v", diags)
	}

	want := []map[string]interface{}{{skCRLAutoFetch: "", skCRLFile: testCRLPEM, skCRLCheck: true}}
	if got := mock.putBodies(); !reflect.DeepEqual(got, want) {
		t.Errorf("PUT bodies = %v, want %v", got, want)
	}
	// The API never returns crl_file: the configured value must survive the read after create.
	if got := d.State().Attributes[skCRLFile]; got != testCRLPEM {
		t.Errorf("crl_file in state = %q, want the configured value", got)
	}
}

// applyConfigX509CRL plans then applies the configuration from a prior state, as Terraform does,
// and returns the resulting state.
func applyConfigX509CRL(
	t *testing.T, mock *mockSingletonConfig, prior map[string]string, config map[string]interface{},
) (*terraform.InstanceState, diag.Diagnostics) {
	t.Helper()
	r := resourceConfigX509CRL()
	state := &terraform.InstanceState{ID: configX509CRLID, Attributes: prior}
	state.Attributes["id"] = configX509CRLID
	instanceDiff, err := r.Diff(t.Context(), state, terraform.NewResourceConfigRaw(config), mock.client(t))
	if err != nil {
		t.Fatalf("diff: %v", err)
	}

	return r.Apply(t.Context(), state, instanceDiff, mock.client(t))
}

// Uploading a crl_file doesn't clear the URL on the Bastion (checked on a Bastion 12.4): the
// update must send an empty crl_auto_fetch, otherwise the URL read back is a permanent diff.
func TestConfigX509CRLUpdateFromURLToFileClearsURL(t *testing.T) {
	mock := newMockSingletonConfig(t, configX509CRLPath,
		map[string]interface{}{skCRLAutoFetch: testCRLURL, skCRLCheck: false}, skCRLFile)

	state, diags := applyConfigX509CRL(t, mock,
		map[string]string{skCRLAutoFetch: testCRLURL, skCRLCheck: "false"},
		map[string]interface{}{skCRLCheck: false, skCRLFile: testCRLPEM})
	if diags.HasError() {
		t.Fatalf("apply: %v", diags)
	}

	want := []map[string]interface{}{{skCRLAutoFetch: "", skCRLFile: testCRLPEM, skCRLCheck: false}}
	if got := mock.putBodies(); !reflect.DeepEqual(got, want) {
		t.Errorf("PUT bodies = %v, want %v", got, want)
	}
	if got := state.Attributes[skCRLAutoFetch]; got != "" {
		t.Errorf("crl_auto_fetch = %q after switching to crl_file, want empty", got)
	}
	if got := state.Attributes[skCRLFile]; got != testCRLPEM {
		t.Errorf("crl_file in state = %q, want the uploaded file", got)
	}
}

// crl_file is never read back: a failed upload must leave the previous file in state, otherwise
// the next plan shows no change and the new CRL is never uploaded.
func TestConfigX509CRLFailedUpdateKeepsPreviousFile(t *testing.T) {
	mock := newMockSingletonConfig(t, configX509CRLPath,
		map[string]interface{}{skCRLAutoFetch: "", skCRLCheck: false}, skCRLFile)
	mock.failPuts(http.StatusBadRequest)

	state, diags := applyConfigX509CRL(t, mock,
		map[string]string{skCRLAutoFetch: "", skCRLCheck: "false", skCRLFile: "OLD-CRL"},
		map[string]interface{}{skCRLCheck: false, skCRLFile: testCRLPEM})
	if !diags.HasError() {
		t.Fatal("expected the update to fail")
	}
	if got := state.Attributes[skCRLFile]; got != "OLD-CRL" {
		t.Errorf("crl_file in state after a failed update = %q, want the previous file", got)
	}
}

func TestConfigX509CRLCreateFailureReturnsError(t *testing.T) {
	mock := newMockSingletonConfig(t, configX509CRLPath,
		map[string]interface{}{skCRLAutoFetch: "", skCRLCheck: false}, skCRLFile)
	mock.failPuts(http.StatusBadRequest)
	r := resourceConfigX509CRL()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{
		skCRLCheck:     false,
		skCRLAutoFetch: testCRLURL,
	})

	if diags := r.CreateContext(t.Context(), d, mock.client(t)); !diags.HasError() {
		t.Fatal("expected the create to fail when the API rejects the PUT")
	}
	if d.Id() != "" {
		t.Errorf("id = %q after a failed create, want empty", d.Id())
	}
}

func TestDataSourceConfigX509CRLRead(t *testing.T) {
	mock := newMockSingletonConfig(t, configX509CRLPath,
		map[string]interface{}{skCRLAutoFetch: testCRLURL, skCRLCheck: true}, skCRLFile)
	ds := dataSourceConfigX509CRL()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{})

	if diags := ds.ReadContext(t.Context(), d, mock.client(t)); diags.HasError() {
		t.Fatalf("read: %v", diags)
	}
	if d.Id() != configX509CRLID || d.Get(skCRLAutoFetch).(string) != testCRLURL || !d.Get(skCRLCheck).(bool) {
		t.Errorf("id, crl_auto_fetch, crl_check = %q, %q, %v, want the values of the API",
			d.Id(), d.Get(skCRLAutoFetch), d.Get(skCRLCheck))
	}
}

func TestConfigX509CRLReadDetectsDrift(t *testing.T) {
	mock := newMockSingletonConfig(t, configX509CRLPath,
		map[string]interface{}{skCRLAutoFetch: testCRLURL, skCRLCheck: true}, skCRLFile)
	r := resourceConfigX509CRL()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{
		skCRLCheck:     true,
		skCRLAutoFetch: testCRLURL,
	})
	d.SetId(configX509CRLID)

	// Changed outside of Terraform, e.g. in the GUI.
	mock.setState(map[string]interface{}{skCRLAutoFetch: "https://other.example.com/crl.pem", skCRLCheck: false})

	if diags := r.ReadContext(t.Context(), d, mock.client(t)); diags.HasError() {
		t.Fatalf("read: %v", diags)
	}
	if got := d.Get(skCRLAutoFetch).(string); got != "https://other.example.com/crl.pem" {
		t.Errorf("crl_auto_fetch = %q, want the value read from the API", got)
	}
	if got := d.Get(skCRLCheck).(bool); got {
		t.Error("crl_check = true, want the value read from the API (false)")
	}
	if len(mock.putBodies()) != 0 {
		t.Errorf("read sent PUT requests: %v", mock.putBodies())
	}
}

func TestConfigX509CRLDeleteDoesNotCallAPI(t *testing.T) {
	mock := newMockSingletonConfig(t, configX509CRLPath,
		map[string]interface{}{skCRLAutoFetch: testCRLURL, skCRLCheck: true}, skCRLFile)
	r := resourceConfigX509CRL()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{
		skCRLCheck:     true,
		skCRLAutoFetch: testCRLURL,
	})
	d.SetId(configX509CRLID)

	if diags := r.DeleteContext(t.Context(), d, mock.client(t)); diags.HasError() {
		t.Fatalf("delete: %v", diags)
	}
	if d.Id() != "" {
		t.Errorf("id = %q after delete, want empty", d.Id())
	}
	if n := mock.requestCount(); n != 0 {
		t.Errorf("delete sent %d API requests, want 0 (the CRL check must stay as is)", n)
	}
}

func TestConfigX509CRLRejectsURLAndFileTogether(t *testing.T) {
	r := resourceConfigX509CRL()
	diags := r.Validate(terraform.NewResourceConfigRaw(map[string]interface{}{
		skCRLCheck:     true,
		skCRLAutoFetch: testCRLURL,
		skCRLFile:      testCRLPEM,
	}))
	if !diags.HasError() {
		t.Error("crl_auto_fetch and crl_file set together must be rejected")
	}
}
