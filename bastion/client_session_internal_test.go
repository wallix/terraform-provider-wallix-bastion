package bastion

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"
)

// sessionServer is a TLS Bastion API whose sessions can be expired. Like Bastion 12.4.3, a login
// ends the session its request presents, and a request with an ended session gets 401.
type sessionServer struct {
	server *httptest.Server
	mu     sync.Mutex
	valid  map[string]bool
	logins int
}

func newSessionServer(t *testing.T) *sessionServer {
	t.Helper()
	ss := &sessionServer{valid: make(map[string]bool)}
	ss.server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, _ := r.Cookie("wab_session_id")
		if r.Method == http.MethodPost && r.URL.Path == "/api" {
			time.Sleep(20 * time.Millisecond)
			ss.mu.Lock()
			if cookie != nil {
				delete(ss.valid, cookie.Value)
			}
			ss.logins++
			id := fmt.Sprintf("session-%d", ss.logins)
			ss.valid[id] = true
			ss.mu.Unlock()
			http.SetCookie(w, &http.Cookie{Name: "wab_session_id", Value: id, Path: "/"})
			w.WriteHeader(http.StatusNoContent)

			return
		}
		time.Sleep(40 * time.Millisecond)
		ss.mu.Lock()
		ok := cookie != nil && ss.valid[cookie.Value]
		ss.mu.Unlock()
		if !ok {
			// Spread the 401s of the requests in flight, as their processing times differ.
			i, _ := strconv.Atoi(r.URL.Query().Get("i"))
			time.Sleep(time.Duration(i) * 30 * time.Millisecond)
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, `{"error":"Authentication required"}`)

			return
		}
		w.Header().Set(contentTypeHeader, contentTypeJSON)
		fmt.Fprint(w, `{}`)
	}))
	t.Cleanup(ss.server.Close)

	return ss
}

func (ss *sessionServer) expireSessions() {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	clear(ss.valid)
}

func (ss *sessionServer) loginCount() int {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	return ss.logins
}

// When the session expires, every request in flight gets a 401: one login must serve them all,
// since each extra login ended the session that the retries of the others were using.
func TestNewRequestConcurrent401LogInOnce(t *testing.T) {
	ss := newSessionServer(t)
	client := newTestClient(t, ss.server.URL)
	client.csrfEnabled = true

	if _, code, err := client.newRequest(t.Context(), "/x", http.MethodGet, nil); err != nil || code != http.StatusOK {
		t.Fatalf("first request: code %d, error %v", code, err)
	}
	ss.expireSessions()

	var wg sync.WaitGroup
	for i := range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, code, err := client.newRequest(t.Context(), "/x?i="+strconv.Itoa(i), http.MethodGet, nil)
			if err != nil || code != http.StatusOK {
				t.Errorf("request %d: code %d, error %v", i, code, err)
			}
		}()
	}
	wg.Wait()

	if n := ss.loginCount(); n != 2 {
		t.Errorf("%d logins, want 2: the first one and one after the session expired", n)
	}
}
