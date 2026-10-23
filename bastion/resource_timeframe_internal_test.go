package bastion

import (
	"cmp"
	"fmt"
	"net/http"
	"path"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

const timeframeType = "wallix-bastion_timeframe"

// newTestProviderServer returns the muxed provider server, configured for the test API of api.
func newTestProviderServer(t *testing.T, api *Client) tfprotov5.ProviderServer { //nolint:ireturn // unexported mux type
	t.Helper()
	clearProviderEnv(t)
	factory, err := NewProviderServer(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	server := factory()
	resp, err := server.ConfigureProvider(t.Context(), &tfprotov5.ConfigureProviderRequest{
		Config: dynamicValue(t, schemaType(t, server, ""), map[string]interface{}{
			"ip": api.bastionIP, skUser: api.bastionUser, skPassword: api.bastionPwd,
			skPort: api.bastionPort, "insecure_skip_verify": true,
		}),
	})
	if err != nil || hasError(resp.Diagnostics) {
		t.Fatalf("configure: %v %v", err, resp.Diagnostics)
	}

	return server
}

// schemaType returns the type of the provider schema, or of a resource or "data." data source schema.
func schemaType(t *testing.T, server tfprotov5.ProviderServer, resourceType string) tftypes.Object {
	t.Helper()
	resp, err := server.GetProviderSchema(t.Context(), &tfprotov5.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatal(err)
	}
	s := resp.Provider
	switch dataSource, ok := strings.CutPrefix(resourceType, "data."); {
	case ok:
		s = resp.DataSourceSchemas[dataSource]
	case resourceType != "":
		s = resp.ResourceSchemas[resourceType]
	}
	typ, ok := s.ValueType().(tftypes.Object)
	if !ok {
		t.Fatalf("schema type of %q is not an object", resourceType)
	}

	return typ
}

func dynamicValue(t *testing.T, typ tftypes.Object, values map[string]interface{}) *tfprotov5.DynamicValue {
	t.Helper()
	v, err := tfprotov5.NewDynamicValue(typ, objectValue(typ, values))
	if err != nil {
		t.Fatal(err)
	}

	return &v
}

func nullValue(t *testing.T, typ tftypes.Object) *tfprotov5.DynamicValue {
	t.Helper()
	v, err := tfprotov5.NewDynamicValue(typ, tftypes.NewValue(typ, nil))
	if err != nil {
		t.Fatal(err)
	}

	return &v
}

func hasError(diags []*tfprotov5.Diagnostic) bool {
	return slices.ContainsFunc(diags, func(d *tfprotov5.Diagnostic) bool {
		return d.Severity == tfprotov5.DiagnosticSeverityError
	})
}

func stateAttributes(t *testing.T, state *tfprotov5.DynamicValue, typ tftypes.Object) map[string]tftypes.Value {
	t.Helper()
	v, err := state.Unmarshal(typ)
	if err != nil {
		t.Fatal(err)
	}
	var attrs map[string]tftypes.Value
	if err := v.As(&attrs); err != nil {
		t.Fatal(err)
	}

	return attrs
}

// timeframeValues returns the attributes of a timeframe named name, without periods by default.
func timeframeValues(name string, extra map[string]interface{}) map[string]interface{} {
	values := map[string]interface{}{"id": name, "timeframe_name": name, "periods": []tftypes.Value{}}
	for k, v := range extra {
		values[k] = v
	}

	return values
}

// SDKv2 (0.16) stored "" and false for unset description and is_overtimable: read back as null,
// the plan of a configuration without them is empty. In a Framework state they come from the
// configuration and stay.
func TestTimeframeReadBackOfZeroValues(t *testing.T) {
	server := newTestProviderServer(t, newTestAPIClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(contentTypeHeader, contentTypeJSON)
		fmt.Fprintf(w, `{"timeframe_name":%q,"description":"","is_overtimable":false,"periods":[]}`,
			path.Base(r.URL.Path))
	}))
	typ := schemaType(t, server, timeframeType)
	zero := map[string]interface{}{skDescription: "", "is_overtimable": false}
	frameworkPrivate := []byte(`{"` + privateKeyFrameworkState + `":"dHJ1ZQ=="}`) // base64 of true

	for name, tc := range map[string]struct {
		prior    map[string]interface{}
		private  []byte
		wantNull bool
	}{
		"SDKv2 state":                  {prior: zero, private: []byte("null"), wantNull: true},
		"imported state":               {wantNull: true},
		"Framework state, zero values": {prior: zero, private: frameworkPrivate},
		"Framework state, null values": {private: frameworkPrivate, wantNull: true},
	} {
		t.Run(name, func(t *testing.T) {
			resp, err := server.ReadResource(t.Context(), &tfprotov5.ReadResourceRequest{
				TypeName:     timeframeType,
				CurrentState: dynamicValue(t, typ, timeframeValues("tf", tc.prior)),
				Private:      tc.private,
			})
			if err != nil || hasError(resp.Diagnostics) {
				t.Fatalf("read: %v %v", err, resp.Diagnostics)
			}
			attrs := stateAttributes(t, resp.NewState, typ)
			for _, attr := range []string{skDescription, "is_overtimable"} {
				if attrs[attr].IsNull() != tc.wantNull {
					t.Errorf("%s = %v, want null: %t", attr, attrs[attr], tc.wantNull)
				}
			}
			if !strings.Contains(string(resp.Private), privateKeyFrameworkState) {
				t.Errorf("private = %s, want the %s key", resp.Private, privateKeyFrameworkState)
			}
		})
	}
}

// The API renames a timeframe on PUT; the provider kept the old name as id, so the read after
// the update didn't find it and dropped it from state ("Root object was present, but now absent").
func TestTimeframeRenameFollowsTheNewName(t *testing.T) {
	var mu sync.Mutex
	var requests []string
	server := newTestProviderServer(t, newTestAPIClient(t, func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requests = append(requests, r.Method+" "+r.URL.Path)
		mu.Unlock()
		if r.Method == http.MethodGet {
			w.Header().Set(contentTypeHeader, contentTypeJSON)
			fmt.Fprintf(w, `{"timeframe_name":%q,"description":"","is_overtimable":false,"periods":[]}`,
				path.Base(r.URL.Path))

			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	typ := schemaType(t, server, timeframeType)

	resp, err := server.ApplyResourceChange(t.Context(), &tfprotov5.ApplyResourceChangeRequest{
		TypeName:     timeframeType,
		PriorState:   dynamicValue(t, typ, timeframeValues("old", nil)),
		PlannedState: dynamicValue(t, typ, timeframeValues("new", nil)),
		Config:       dynamicValue(t, typ, timeframeValues("new", map[string]interface{}{"id": nil})),
	})
	if err != nil || hasError(resp.Diagnostics) {
		t.Fatalf("update: %v %v", err, resp.Diagnostics)
	}
	if id := stateAttributes(t, resp.NewState, typ)["id"]; !id.Equal(tftypes.NewValue(tftypes.String, "new")) {
		t.Errorf("id = %v after the rename, want new", id)
	}
	const base = "/api/" + VersionWallixAPI312
	for _, want := range []string{"PUT " + base + "/timeframes/old", "GET " + base + "/timeframes/new"} {
		if !slices.Contains(requests, want) {
			t.Errorf("requests = %v, want %q", requests, want)
		}
	}
}

// A period without week_days, or with a day the API doesn't know, failed only at apply (400 from the API).
func TestTimeframeWeekDaysValidation(t *testing.T) {
	server := newTestProviderServer(t, newTestAPIClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	typ := schemaType(t, server, timeframeType)
	periodsType, ok := typ.AttributeTypes["periods"].(tftypes.Set)
	if !ok {
		t.Fatal("periods is not a set")
	}
	periodType, ok := periodsType.ElementType.(tftypes.Object)
	if !ok {
		t.Fatal("periods is not a set of objects")
	}
	days := func(d ...string) []tftypes.Value {
		values := make([]tftypes.Value, len(d))
		for i, day := range d {
			values[i] = tftypes.NewValue(tftypes.String, day)
		}

		return values
	}
	for name, tc := range map[string]struct {
		weekDays  []tftypes.Value
		startDate string
		wantErr   string
	}{
		"no week_days":    {wantErr: "Missing Configuration for Required Attribute"},
		"empty week_days": {weekDays: days(), wantErr: "at least 1"},
		"capitalized day": {weekDays: days("Monday"), wantErr: "value must be one of"},
		"bad date":        {weekDays: days("monday"), startDate: "2020-1-1", wantErr: "yyyy-mm-dd"},
		"lower case days": {weekDays: days("monday", "sunday")},
	} {
		t.Run(name, func(t *testing.T) {
			values := map[string]interface{}{
				"start_date": cmp.Or(tc.startDate, "2020-01-01"), "end_date": "2020-02-02",
				"start_time": "08:00", "end_time": "12:00",
			}
			if tc.weekDays != nil {
				values["week_days"] = tc.weekDays
			}
			period := objectValue(periodType, values)
			config := timeframeValues("tf-plan", map[string]interface{}{"id": nil, "periods": []tftypes.Value{period}})
			resp, err := server.ValidateResourceTypeConfig(t.Context(), &tfprotov5.ValidateResourceTypeConfigRequest{
				TypeName: timeframeType, Config: dynamicValue(t, typ, config),
			})
			if err != nil {
				t.Fatal(err)
			}
			got := make([]string, 0, len(resp.Diagnostics))
			for _, d := range resp.Diagnostics {
				got = append(got, d.Summary+": "+d.Detail)
			}
			if hasError(resp.Diagnostics) != (tc.wantErr != "") {
				t.Fatalf("validate = %v, want an error: %t", got, tc.wantErr != "")
			}
			if tc.wantErr != "" && !strings.Contains(strings.Join(got, "\n"), tc.wantErr) {
				t.Errorf("validate = %v, want %q", got, tc.wantErr)
			}
		})
	}
}
