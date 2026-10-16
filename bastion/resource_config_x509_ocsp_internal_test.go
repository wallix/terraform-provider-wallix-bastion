package bastion

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func defaultOCSPState() map[string]interface{} {
	return map[string]interface{}{
		skOCSPCheck:             false,
		skDefaultOCSPResponder:  "",
		skOverrideOCSPResponder: false,
		skOCSPProxy:             "",
	}
}

func TestConfigX509OCSPCreateSendsAllFields(t *testing.T) {
	mock := newMockSingletonConfig(t, configX509OCSPPath, defaultOCSPState())
	r := resourceConfigX509OCSP()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{
		skOCSPCheck:            true,
		skDefaultOCSPResponder: "http://ocsp.example.com",
	})

	if diags := r.CreateContext(t.Context(), d, mock.client(t)); diags.HasError() {
		t.Fatalf("create: %v", diags)
	}

	// Unset arguments are sent with their zero value, as in the documented PUT example, so that
	// removing one from the configuration resets it on the Bastion.
	want := []map[string]interface{}{{
		skOCSPCheck:             true,
		skDefaultOCSPResponder:  "http://ocsp.example.com",
		skOverrideOCSPResponder: false,
		skOCSPProxy:             "",
	}}
	if got := mock.putBodies(); !reflect.DeepEqual(got, want) {
		t.Errorf("PUT bodies = %v, want %v", got, want)
	}
	if d.Id() != configX509OCSPID {
		t.Errorf("id = %q, want x509OCSPConfig", d.Id())
	}
}

func TestConfigX509OCSPReadDetectsDrift(t *testing.T) {
	mock := newMockSingletonConfig(t, configX509OCSPPath, defaultOCSPState())
	r := resourceConfigX509OCSP()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{
		skOCSPCheck: true,
	})
	d.SetId(configX509OCSPID)

	mock.setState(map[string]interface{}{
		skOCSPCheck:             false,
		skDefaultOCSPResponder:  "http://other.example.com",
		skOverrideOCSPResponder: true,
		skOCSPProxy:             "http://proxy.example.com:3128",
	})

	if diags := r.ReadContext(t.Context(), d, mock.client(t)); diags.HasError() {
		t.Fatalf("read: %v", diags)
	}
	want := map[string]string{
		skOCSPCheck:             "false",
		skDefaultOCSPResponder:  "http://other.example.com",
		skOverrideOCSPResponder: "true",
		skOCSPProxy:             "http://proxy.example.com:3128",
	}
	attrs := d.State().Attributes
	for k, v := range want {
		if attrs[k] != v {
			t.Errorf("%s = %q, want %q", k, attrs[k], v)
		}
	}
}

func TestConfigX509OCSPDeleteDoesNotCallAPI(t *testing.T) {
	mock := newMockSingletonConfig(t, configX509OCSPPath, defaultOCSPState())
	r := resourceConfigX509OCSP()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{
		skOCSPCheck: true,
	})
	d.SetId(configX509OCSPID)

	if diags := r.DeleteContext(t.Context(), d, mock.client(t)); diags.HasError() {
		t.Fatalf("delete: %v", diags)
	}
	if d.Id() != "" {
		t.Errorf("id = %q after delete, want empty", d.Id())
	}
	if n := mock.requestCount(); n != 0 {
		t.Errorf("delete sent %d API requests, want 0 (the OCSP check must stay as is)", n)
	}
}

func TestDataSourceConfigX509OCSPRead(t *testing.T) {
	state := defaultOCSPState()
	state[skOCSPCheck] = true
	state[skDefaultOCSPResponder] = "http://ocsp.example.com"
	mock := newMockSingletonConfig(t, configX509OCSPPath, state)
	ds := dataSourceConfigX509OCSP()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{})

	if diags := ds.ReadContext(t.Context(), d, mock.client(t)); diags.HasError() {
		t.Fatalf("read: %v", diags)
	}
	if d.Id() != configX509OCSPID || !d.Get(skOCSPCheck).(bool) ||
		d.Get(skDefaultOCSPResponder).(string) != "http://ocsp.example.com" {
		t.Errorf("id, ocsp_check, default_ocsp_responder = %q, %v, %q, want the values of the API",
			d.Id(), d.Get(skOCSPCheck), d.Get(skDefaultOCSPResponder))
	}
}
