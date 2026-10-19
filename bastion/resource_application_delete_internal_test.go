package bastion

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// appDeleteMock is a Bastion API with an application "app" in the global domain "dom", whose
// account "acc" has the resources "app:WEBAPP" (when attached) and "other:SSH". Like a Bastion 12.4,
// it refuses to delete the application while an account is attached to it. With deviceAccount, the
// account "dev-acc", listed first, has the service SSH of a device also named "app".
type appDeleteMock struct {
	mu             sync.Mutex
	attached       bool
	alwaysConflict bool
	otherConflict  bool
	deviceAccount  bool
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
		if mock.otherConflict {
			w.WriteHeader(http.StatusConflict)
			fmt.Fprint(w, `{"error":"Constraint violation","description":"Another constraint."}`)

			return
		}
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
		accounts := `{"id":"acc-id","resources":["other:SSH"]}`
		if mock.attached {
			accounts = `{"id":"acc-id","resources":["app:WEBAPP","other:SSH"]}`
		}
		if mock.deviceAccount {
			accounts = `{"id":"dev-acc-id","resources":["app:SSH"]},` + accounts
		}
		fmt.Fprint(w, "["+accounts+"]")
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

// A device with the same name as the application ("app:SSH") is not the application: its account
// is not detached, and its 404 doesn't hide the application's own account.
func TestDeleteApplicationIgnoresDeviceWithSameName(t *testing.T) {
	mock := &appDeleteMock{attached: true, deviceAccount: true}
	if err := deleteTestApplication(t, mock); err != nil {
		t.Fatalf("delete: %v", err)
	}
	for _, request := range mock.requests {
		if strings.Contains(request, "/accounts/dev-acc-id/") {
			t.Errorf("request %q, want the device account left alone", request)
		}
	}
}

// A conflict that isn't caused by an attached account is returned as is, without detaching
// accounts or retrying.
func TestDeleteApplicationReturnsOtherConflicts(t *testing.T) {
	mock := &appDeleteMock{attached: true, otherConflict: true}
	err := deleteTestApplication(t, mock)
	if err == nil || !strings.Contains(err.Error(), "Another constraint.") {
		t.Fatalf("delete error = %v, want the original 409", err)
	}
	for _, request := range mock.requests {
		if strings.Contains(request, "/accounts/") {
			t.Errorf("request %q, want no detach for another conflict", request)
		}
	}
}

// When the deletion still fails after detaching, the error says that accounts were detached.
func TestDeleteApplicationReportsDetachOnSecondConflict(t *testing.T) {
	mock := &appDeleteMock{attached: true, alwaysConflict: true}
	err := deleteTestApplication(t, mock)
	if err == nil || !strings.Contains(err.Error(), "detached from 1 global domain account(s)") {
		t.Fatalf("delete error = %v, want the 409 and the detached accounts", err)
	}
}

// Without any attached account, a "still associated" conflict is returned without retrying.
func TestDeleteApplicationDoesNotRetryWithoutDetachedAccount(t *testing.T) {
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
