package bastion

import (
	"fmt"
	"net/http"
	"slices"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// appDeleteMock is a Bastion API with an application "app" in the global domain "dom", whose
// account "acc" has the resources "app:WEBAPP" (when attached) and "other:SSH". Like a Bastion 12.4,
// it refuses to delete the application while an account is attached to it.
type appDeleteMock struct {
	mu             sync.Mutex
	attached       bool
	alwaysConflict bool
	requests       []string
}

func (mock *appDeleteMock) serveHTTP(w http.ResponseWriter, r *http.Request) {
	mock.mu.Lock()
	defer mock.mu.Unlock()
	mock.requests = append(mock.requests, r.Method+" "+r.URL.Path)
	w.Header().Set(contentTypeHeader, contentTypeJSON)
	const base = "/api/" + VersionWallixAPI312
	switch {
	case r.Method == http.MethodDelete && r.URL.Path == base+"/applications/app-id":
		if mock.attached || mock.alwaysConflict {
			w.WriteHeader(http.StatusConflict)
			fmt.Fprint(w, `{"error":"Constraint violation",`+
				`"description":"The given application is still associated with one target. Unable to delete it."}`)

			return
		}
		w.WriteHeader(http.StatusNoContent)
	case r.Method == http.MethodGet && r.URL.Path == base+"/applications/app-id":
		fmt.Fprint(w, `{"id":"app-id","application_name":"app","connection_policy":"RDP","global_domains":["dom"]}`)
	case r.Method == http.MethodGet && r.URL.Path == base+"/domains/" && r.URL.RawQuery == "q=domain_name=dom":
		fmt.Fprint(w, `[{"id":"dom-id","domain_name":"dom"}]`)
	case r.Method == http.MethodGet && r.URL.Path == base+"/domains/dom-id/accounts/":
		if mock.attached {
			fmt.Fprint(w, `[{"id":"acc-id","resources":["app:WEBAPP","other:SSH"]}]`)
		} else {
			fmt.Fprint(w, `[{"id":"acc-id","resources":["other:SSH"]}]`)
		}
	case r.Method == http.MethodDelete && r.URL.Path == base+"/domains/dom-id/accounts/acc-id/resource/app:APP":
		mock.attached = false
		w.WriteHeader(http.StatusNoContent)
	case r.Method == http.MethodDelete && r.URL.Path == base+"/domains/dom-id/accounts/acc-id/resource/app:WEBAPP":
		// The detach request reads "<name>:WEBAPP" as the service WEBAPP of a device <name>.
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"error":"Resource not found","description":"Device 'app' not found"}`)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (mock *appDeleteMock) client(t *testing.T) *Client {
	t.Helper()

	return newTestAPIClient(t, mock.serveHTTP)
}

func deleteTestApplication(t *testing.T, mock *appDeleteMock) error {
	t.Helper()
	r := resourceApplication()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{"application_name": "app"})
	d.SetId("app-id")

	return deleteApplication(t.Context(), d, mock.client(t))
}

func TestDeleteApplicationDetachesGlobalDomainAccounts(t *testing.T) {
	mock := &appDeleteMock{attached: true}
	if err := deleteTestApplication(t, mock); err != nil {
		t.Fatalf("delete: %v", err)
	}

	const base = "/api/" + VersionWallixAPI312
	detach := "DELETE " + base + "/domains/dom-id/accounts/acc-id/resource/app:APP"
	if !slices.Contains(mock.requests, detach) {
		t.Errorf("requests = %v, want the detach request %q", mock.requests, detach)
	}
	if slices.Contains(mock.requests, "DELETE "+base+"/domains/dom-id/accounts/acc-id/resource/other:SSH") {
		t.Error("the resources of other applications or devices must stay attached")
	}
	if last := mock.requests[len(mock.requests)-1]; last != "DELETE "+base+"/applications/app-id" {
		t.Errorf("last request = %q, want the second DELETE of the application", last)
	}
}

// A conflict that isn't caused by an attached account is returned as is, without retrying.
func TestDeleteApplicationReturnsOtherConflicts(t *testing.T) {
	mock := &appDeleteMock{alwaysConflict: true}
	if err := deleteTestApplication(t, mock); err == nil {
		t.Fatal("expected the 409 to be returned")
	}

	deletes := 0
	for _, request := range mock.requests {
		if request == "DELETE /api/"+VersionWallixAPI312+"/applications/app-id" {
			deletes++
		}
	}
	if deletes != 1 {
		t.Errorf("the application was deleted %d times, want 1 (no retry without detached account)", deletes)
	}
}
