package bastion

import (
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

const timeframeDataSourceType = "data." + timeframeType

func timeframeAPI(w http.ResponseWriter, r *http.Request) {
	if !strings.HasSuffix(r.URL.Path, "/timeframes/tf") {
		w.WriteHeader(http.StatusNotFound)

		return
	}
	w.Header().Set(contentTypeHeader, contentTypeJSON)
	fmt.Fprint(w, `{"timeframe_name":"tf","description":"","is_overtimable":false,"periods":[`+
		`{"start_date":"2020-01-01","end_date":"2020-02-02","start_time":"08:00","end_time":"12:00",`+
		`"week_days":["monday"]}]}`)
}

func readTimeframeDataSource(
	t *testing.T, server tfprotov5.ProviderServer, name string,
) *tfprotov5.ReadDataSourceResponse {
	t.Helper()
	typ := schemaType(t, server, timeframeDataSourceType)
	resp, err := server.ReadDataSource(t.Context(), &tfprotov5.ReadDataSourceRequest{
		TypeName: timeframeType,
		Config:   dynamicValue(t, typ, map[string]interface{}{"timeframe_name": name}),
	})
	if err != nil {
		t.Fatal(err)
	}

	return resp
}

// Computed values keep the zero values SDKv2 stored, so that references to them don't change.
func TestTimeframeDataSourceRead(t *testing.T) {
	server := newTestProviderServer(t, newTestAPIClient(t, timeframeAPI))
	typ := schemaType(t, server, timeframeDataSourceType)

	resp := readTimeframeDataSource(t, server, "tf")
	if hasError(resp.Diagnostics) {
		t.Fatalf("read: %v", resp.Diagnostics)
	}
	attrs := stateAttributes(t, resp.State, typ)
	for attr, want := range map[string]tftypes.Value{
		"id":             tftypes.NewValue(tftypes.String, "tf"),
		skDescription:    tftypes.NewValue(tftypes.String, ""),
		"is_overtimable": tftypes.NewValue(tftypes.Bool, false),
	} {
		if !attrs[attr].Equal(want) {
			t.Errorf("%s = %v, want %v", attr, attrs[attr], want)
		}
	}
	var periods []tftypes.Value
	if err := attrs["periods"].As(&periods); err != nil || len(periods) != 1 {
		t.Errorf("periods = %v, want one period", attrs["periods"])
	}

	resp = readTimeframeDataSource(t, server, "missing")
	if !hasError(resp.Diagnostics) || !strings.Contains(resp.Diagnostics[0].Summary, "doesn't exists") {
		t.Errorf("read of a missing timeframe: %v, want doesn't exists", resp.Diagnostics)
	}
}

// The SDKv2 and Framework servers are configured separately: they must share one Bastion session.
func TestMuxedServersShareOneSession(t *testing.T) {
	var logins atomic.Int32
	server := newTestProviderServer(t, newCountingTestAPIClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/"+VersionWallixAPI312+"/version" {
			w.Header().Set(contentTypeHeader, contentTypeJSON)
			fmt.Fprint(w, `{}`)

			return
		}
		timeframeAPI(w, r)
	}, &logins))

	versionType := schemaType(t, server, "data.wallix-bastion_version")
	sdkResp, err := server.ReadDataSource(t.Context(), &tfprotov5.ReadDataSourceRequest{
		TypeName: "wallix-bastion_version",
		Config:   dynamicValue(t, versionType, nil),
	})
	if err != nil || hasError(sdkResp.Diagnostics) {
		t.Fatalf("SDKv2 read: %v %v", err, sdkResp.Diagnostics)
	}
	if resp := readTimeframeDataSource(t, server, "tf"); hasError(resp.Diagnostics) {
		t.Fatalf("Framework read: %v", resp.Diagnostics)
	}
	if n := logins.Load(); n != 1 {
		t.Errorf("%d logins, want 1", n)
	}
}
