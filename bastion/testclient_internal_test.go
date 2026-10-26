package bastion

import (
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync/atomic"
	"testing"
)

// newTestAPIClient starts a TLS Bastion API that authenticates any POST /api and passes every
// other request to handler, and returns a client of it.
func newTestAPIClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()

	return newCountingTestAPIClient(t, handler, new(atomic.Int32))
}

// newCountingTestAPIClient is newTestAPIClient that counts the logins.
func newCountingTestAPIClient(t *testing.T, handler http.HandlerFunc, logins *atomic.Int32) *Client {
	t.Helper()
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/api" {
			logins.Add(1)
			http.SetCookie(w, &http.Cookie{Name: "wab_session_id", Value: "test-session", Path: "/"})
			w.WriteHeader(http.StatusNoContent)

			return
		}
		handler(w, r)
	}))
	t.Cleanup(server.Close)

	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("parsing server URL: %v", err)
	}
	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		t.Fatalf("splitting server host: %v", err)
	}
	portNum, err := strconv.Atoi(port)
	if err != nil {
		t.Fatalf("parsing server port: %v", err)
	}

	return &Client{
		bastionIP:          host,
		bastionPort:        portNum,
		bastionAPIVersion:  VersionWallixAPI312,
		bastionUser:        "test-api-user",
		bastionPwd:         "test-api-password",
		insecureSkipVerify: true,
	}
}
