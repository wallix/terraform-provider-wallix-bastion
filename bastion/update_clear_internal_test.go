package bastion

import (
	"fmt"
	"io"
	"maps"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// updateRecorder is a Bastion API answering every GET with readBody and recording the PUT requests.
type updateRecorder struct {
	mu   sync.Mutex
	puts []string
}

func (rec *updateRecorder) client(t *testing.T, readBody string) *Client {
	t.Helper()

	return newTestAPIClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(contentTypeHeader, contentTypeJSON)
		if r.Method == http.MethodPut {
			body, _ := io.ReadAll(r.Body)
			rec.mu.Lock()
			rec.puts = append(rec.puts, strings.TrimPrefix(r.URL.RequestURI(), "/api/"+VersionWallixAPI312)+
				" "+strings.TrimSpace(string(body)))
			rec.mu.Unlock()
			w.WriteHeader(http.StatusNoContent)

			return
		}
		fmt.Fprint(w, readBody)
	})
}

func (rec *updateRecorder) put(t *testing.T) string {
	t.Helper()
	rec.mu.Lock()
	defer rec.mu.Unlock()
	if len(rec.puts) != 1 {
		t.Fatalf("PUT requests = %q, want one", rec.puts)
	}

	return rec.puts[0]
}

// planAndApplyUpdate plans the configuration over a state built from prior, then applies the
// update, as Terraform does.
func planAndApplyUpdate(
	t *testing.T, r *schema.Resource, prior, config map[string]interface{}, client *Client,
) {
	t.Helper()
	d := schema.TestResourceDataRaw(t, r.Schema, prior)
	d.SetId("obj-id")
	state := d.State()
	diff, err := r.Diff(t.Context(), state, terraform.NewResourceConfigRaw(config), client)
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	if diff == nil || diff.Empty() {
		t.Fatal("the plan has no change")
	}
	if _, diags := r.Apply(t.Context(), state, diff, client); diags.HasError() {
		t.Fatalf("apply: %v", diags)
	}
}

func deviceTags(tags ...string) []interface{} {
	set := make([]interface{}, len(tags))
	for i, tag := range tags {
		key, value, _ := strings.Cut(tag, "=")
		set[i] = map[string]interface{}{skKey: key, skValue: value}
	}

	return set
}

// Without force=true, the API adds the tags sent to the existing ones (checked on a Bastion 12.4):
// a tag removed from the configuration stayed on the device. The last tag removed is sent as an
// empty list, and a request without tags leaves the tags managed outside of Terraform as they are.
func TestDeviceUpdateReplacesTags(t *testing.T) {
	const read = `{"id":"obj-id","device_name":"tag-dev","host":"192.0.2.10","local_domains":[],"services":[]}`
	for name, tc := range map[string]struct {
		prior, config []interface{}
		description   string
		want          string
	}{
		"one tag removed": {
			prior:  deviceTags("env=test", "team=infra"),
			config: deviceTags("env=test"),
			want: `/devices/obj-id?force=true {"alias":"","description":"","device_name":"tag-dev",` +
				`"host":"192.0.2.10","tags":[{"key":"env","value":"test"}]}`,
		},
		"last tag removed": {
			prior: deviceTags("env=test"),
			want: `/devices/obj-id?force=true {"alias":"","description":"","device_name":"tag-dev",` +
				`"host":"192.0.2.10","tags":[]}`,
		},
		"no tags, description changed": {
			description: "updated",
			want: `/devices/obj-id?force=true {"alias":"","description":"updated","device_name":"tag-dev",` +
				`"host":"192.0.2.10"}`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			rec := &updateRecorder{}
			prior := map[string]interface{}{"device_name": "tag-dev", skHost: "192.0.2.10"}
			config := map[string]interface{}{"device_name": "tag-dev", skHost: "192.0.2.10", skDescription: tc.description}
			if tc.prior != nil {
				prior["tags"] = tc.prior
			}
			if tc.config != nil {
				config["tags"] = tc.config
			}
			planAndApplyUpdate(t, resourceDevice(), prior, config, rec.client(t, read))
			if got := rec.put(t); got != tc.want {
				t.Errorf("PUT = %s\nwant  %s", got, tc.want)
			}
		})
	}
}

// An emptied list was omitted from the update, so the Bastion kept its values and the plan showed
// the same change after every apply (checked on a Bastion 12.4). It is now sent empty, which
// clears the profile dashboards, the application tags and the subprotocols of an RDP service
// (checked on a Bastion 12.4.1).
func TestUpdateSendsEmptiedLists(t *testing.T) {
	gui := []interface{}{map[string]interface{}{"wab_audit": "view"}}
	for name, tc := range map[string]struct {
		resource      *schema.Resource
		prior, config map[string]interface{}
		read, want    string
	}{
		"profile dashboards": {
			resource: resourceProfile(),
			prior: map[string]interface{}{
				"profile_name": "p", "gui_features": gui, "gui_transmission": gui,
				"dashboards": []interface{}{"audit"},
			},
			config: map[string]interface{}{"profile_name": "p", "gui_features": gui, "gui_transmission": gui},
			read: `{"id":"obj-id","profile_name":"p","dashboards":[],` +
				`"target_groups_limitation":{"enabled":false},"user_groups_limitation":{"enabled":false}}`,
			want: `"dashboards":[]`,
		},
		"application tags": {
			resource: resourceApplication(),
			prior:    webApplicationRaw(map[string]interface{}{"tags": deviceTags("env=test")}),
			config:   webApplicationRaw(nil),
			read: `{"id":"obj-id","application_name":"web-app","category":"web_application",` +
				`"connection_policy":"WEBAPP","application_url":"https://example.com","tags":[]}`,
			want: `"tags":[]`,
		},
		"RDP device service subprotocols": {
			resource: resourceDeviceService(),
			prior: map[string]interface{}{
				skDeviceID: "svc-dev", skServiceName: "RDP", skConnectionPolicy: skProtoRDP, skPort: 3389,
				skProtocol: skProtoRDP, skSubprotocols: []interface{}{"RDP_CLIPBOARD_UP"},
			},
			config: map[string]interface{}{
				skDeviceID: "svc-dev", skServiceName: "RDP", skConnectionPolicy: skProtoRDP, skPort: 3389,
				skProtocol: skProtoRDP,
			},
			read: `{"id":"obj-id","service_name":"RDP","protocol":"RDP","port":3389,` +
				`"connection_policy":"RDP","subprotocols":[]}`,
			want: `"subprotocols":[]`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			rec := &updateRecorder{}
			planAndApplyUpdate(t, tc.resource, tc.prior, tc.config, rec.client(t, tc.read))
			if got := rec.put(t); !strings.Contains(got, tc.want) {
				t.Errorf("PUT = %s, want %s in the body", got, tc.want)
			}
		})
	}
}

// A value omitted from the update keeps its value on the Bastion (checked on a Bastion 12.4.3):
// the emptied parameters of an application and a changed preferred_language or force_change_pwd
// of a user, only sent at creation, were never applied.
func TestUpdateSendsChangedValues(t *testing.T) {
	stdApplication := func(parameters string) map[string]interface{} {
		return map[string]interface{}{
			skConnectionPolicy: skProtoRDP, skTarget: "params-cluster",
			"paths": []interface{}{map[string]interface{}{
				skTarget: "Interactive@params-dev:RDP", "program": "params-prog", "working_dir": "",
			}},
			"parameters": parameters,
		}
	}
	user := func(language string, forceChangePwd bool) map[string]interface{} {
		return map[string]interface{}{
			"user_name": "lang-user", skEmail: "lang-user@example.com", skProfile: "lang-profile",
			"user_auths": []interface{}{"local_password"}, "preferred_language": language,
			"force_change_pwd": forceChangePwd,
		}
	}
	const userRead = `{"user_name":"lang-user","email":"lang-user@example.com","profile":"lang-profile",` +
		`"user_auths":["local_password"],"preferred_language":"fr"}`
	for name, tc := range map[string]struct {
		resource      *schema.Resource
		prior, config map[string]interface{}
		read, want    string
	}{
		"application parameters emptied": {
			resource: resourceApplication(),
			prior:    stdApplication("params-value"),
			config:   stdApplication(""),
			read: `{"id":"obj-id","application_name":"params-app","category":"standard",` +
				`"connection_policy":"RDP","target":"params-cluster","parameters":""}`,
			want: `"parameters":""`,
		},
		"user preferred_language changed": {
			resource: resourceUser(),
			prior:    user("en", false),
			config:   user("fr", false),
			read:     userRead,
			want:     `"preferred_language":"fr"`,
		},
		"user force_change_pwd enabled": {
			resource: resourceUser(),
			prior:    user("fr", false),
			config:   user("fr", true),
			read:     userRead,
			want:     `"force_change_pwd":true`,
		},
		"user force_change_pwd disabled": {
			resource: resourceUser(),
			prior:    user("fr", true),
			config:   user("fr", false),
			read:     userRead,
			want:     `"force_change_pwd":false`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			rec := &updateRecorder{}
			planAndApplyUpdate(t, tc.resource, tc.prior, tc.config, rec.client(t, tc.read))
			if got := rec.put(t); !strings.Contains(got, tc.want) {
				t.Errorf("PUT = %s, want %s in the body", got, tc.want)
			}
		})
	}
}

// The Bastion can't clear the subprotocols of an authorization (400 "At least one subprotocol is
// required"), even with authorize_sessions = false: removing the last ones replaces it, and adding
// or changing some is still an update.
func TestAuthorizationEmptiedSubprotocolsReplaceIt(t *testing.T) {
	base := map[string]interface{}{
		"authorization_name": "a", "user_group": "ug", "target_group": "tg", "authorize_password_retrieval": true,
	}
	withSubprotocols := func(subprotocols ...interface{}) map[string]interface{} {
		raw := maps.Clone(base)
		if len(subprotocols) > 0 {
			raw[skSubprotocols] = subprotocols
		}

		return raw
	}
	for name, tc := range map[string]struct {
		prior, config map[string]interface{}
		replace       bool
	}{
		"last removed": {withSubprotocols("SSH_SHELL_SESSION"), withSubprotocols(), true},
		"one removed":  {withSubprotocols("SSH_SHELL_SESSION", "SSH_SCP_UP"), withSubprotocols("SSH_SCP_UP"), false},
		"first added":  {withSubprotocols(), withSubprotocols("SSH_SHELL_SESSION"), false},
	} {
		t.Run(name, func(t *testing.T) {
			r := resourceAuthorization()
			d := schema.TestResourceDataRaw(t, r.Schema, tc.prior)
			d.SetId("obj-id")
			diff, err := r.Diff(t.Context(), d.State(), terraform.NewResourceConfigRaw(tc.config), nil)
			if err != nil {
				t.Fatalf("plan: %v", err)
			}
			if diff == nil || diff.Empty() {
				t.Fatal("the plan has no change")
			}
			if got := diff.RequiresNew(); got != tc.replace {
				t.Errorf("replacement = %v, want %v", got, tc.replace)
			}
		})
	}
}

// The Bastion refuses an SSH service without subprotocols, at creation and on update (checked on
// a Bastion 12.4.1): refused at plan. RDP accepts none.
func TestDeviceServiceSSHRequiresSubprotocols(t *testing.T) {
	for name, tc := range map[string]struct {
		protocol string
		wantErr  bool
	}{
		"SSH": {skProtoSSH, true},
		"RDP": {skProtoRDP, false},
	} {
		t.Run(name, func(t *testing.T) {
			r := resourceDeviceService()
			config := map[string]interface{}{
				skDeviceID: "svc-dev", skServiceName: "svc", skConnectionPolicy: tc.protocol, skPort: 22,
				skProtocol: tc.protocol,
			}
			_, err := r.Diff(t.Context(), nil, terraform.NewResourceConfigRaw(config), nil)
			if gotErr := err != nil; gotErr != tc.wantErr {
				t.Errorf("plan error = %v, want an error: %v", err, tc.wantErr)
			}
		})
	}
}

// A creation without these lists sends none of them: an empty list would clear what the Bastion
// sets by default.
func TestCreateSendsNoEmptiedLists(t *testing.T) {
	gui := []interface{}{map[string]interface{}{"wab_audit": "view"}}
	device := schema.TestResourceDataRaw(t, resourceDevice().Schema,
		map[string]interface{}{"device_name": "dev", skHost: "192.0.2.10"})
	if tags := prepareDeviceJSON(device).Tags; tags != nil {
		t.Errorf("device tags = %v, want none sent", *tags)
	}
	profile := schema.TestResourceDataRaw(t, resourceProfile().Schema,
		map[string]interface{}{"profile_name": "p", "gui_features": gui, "gui_transmission": gui})
	if dashboards := prepareProfileJSON(profile, true).Dashboards; dashboards != nil {
		t.Errorf("profile dashboards = %v, want none sent", *dashboards)
	}
	application, err := prepareApplicationJSON(
		schema.TestResourceDataRaw(t, resourceApplication().Schema, webApplicationRaw(nil)), true, VersionWallixAPI312)
	if err != nil {
		t.Fatal(err)
	}
	if application.Tags != nil {
		t.Errorf("application tags = %v, want none sent", *application.Tags)
	}
	authorization := schema.TestResourceDataRaw(t, resourceAuthorization().Schema, map[string]interface{}{
		"authorization_name": "a", "user_group": "ug", "target_group": "tg", "authorize_password_retrieval": true,
	})
	if subprotocols := prepareAuthorizationJSON(authorization, true).SubProtocols; subprotocols != nil {
		t.Errorf("authorization subprotocols = %v, want none sent", *subprotocols)
	}
}
