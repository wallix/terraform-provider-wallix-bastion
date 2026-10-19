package bastion

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync"
	"testing"
)

// connCountingServer is a TLS Bastion API that authenticates any POST /api, answers {} to
// everything else, and counts the TCP connections opened by the clients.
type connCountingServer struct {
	server   *httptest.Server
	mu       sync.Mutex
	newConns int
}

func newConnCountingServer(t *testing.T) *connCountingServer {
	t.Helper()
	ccs := &connCountingServer{}
	ccs.server = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/api" {
			http.SetCookie(w, &http.Cookie{Name: "wab_session_id", Value: "test-session", Path: "/"})
			w.WriteHeader(http.StatusNoContent)

			return
		}
		w.Header().Set(contentTypeHeader, contentTypeJSON)
		fmt.Fprint(w, `{}`)
	}))
	ccs.server.Config.ConnState = func(_ net.Conn, state http.ConnState) {
		if state == http.StateNew {
			ccs.mu.Lock()
			ccs.newConns++
			ccs.mu.Unlock()
		}
	}
	ccs.server.StartTLS()
	t.Cleanup(ccs.server.Close)

	return ccs
}

func (ccs *connCountingServer) client(t *testing.T) *Client {
	t.Helper()
	u, err := url.Parse(ccs.server.URL)
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
		bastionUser:        "conn-test-user",
		bastionPwd:         "conn-test-password",
		insecureSkipVerify: true,
	}
}

func (ccs *connCountingServer) connections() int {
	ccs.mu.Lock()
	defer ccs.mu.Unlock()

	return ccs.newConns
}

// A Bastion 12.4 stopped accepting new connections from a client with 30 of them open: the
// provider must reuse its connections instead of opening one per request.
func TestNewRequestReusesConnection(t *testing.T) {
	ccs := newConnCountingServer(t)
	client := ccs.client(t)

	for i := range 20 {
		_, code, err := client.newRequest(t.Context(), "/version", http.MethodGet, nil)
		if err != nil || code != http.StatusOK {
			t.Fatalf("request %d: code %d, error %v", i, code, err)
		}
	}

	if n := ccs.connections(); n != 1 {
		t.Errorf("authentication and 20 requests opened %d connections, want 1", n)
	}
}

// Terraform calls the provider concurrently: the requests share one client and never open more
// than httpMaxConnsPerHost connections.
func TestNewRequestConcurrentRequestsShareCappedConnections(t *testing.T) {
	ccs := newConnCountingServer(t)
	client := ccs.client(t)

	var wg sync.WaitGroup
	errs := make(chan error, 20)
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 5 {
				_, code, err := client.newRequest(t.Context(), "/version", http.MethodGet, nil)
				if err != nil || code != http.StatusOK {
					errs <- fmt.Errorf("code %d, error %w", code, err)

					return
				}
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}

	if n := ccs.connections(); n > httpMaxConnsPerHost {
		t.Errorf("100 concurrent requests opened %d connections, want at most %d", n, httpMaxConnsPerHost)
	}
}

func TestGetHTTPClientIsSharedAndConfigured(t *testing.T) {
	client := &Client{insecureSkipVerify: true}

	httpClient := client.getHTTPClient()
	if httpClient != client.getHTTPClient() {
		t.Fatal("getHTTPClient must return the same client on every call")
	}
	if httpClient.Jar == nil {
		t.Error("the shared client must use the cookie jar of the Client")
	}
	transport, ok := httpClient.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("transport is %T, want *http.Transport", httpClient.Transport)
	}
	if !transport.TLSClientConfig.InsecureSkipVerify {
		t.Error("InsecureSkipVerify of the Client must be applied to the transport")
	}
	if transport.MaxConnsPerHost != httpMaxConnsPerHost || transport.IdleConnTimeout != httpIdleConnTimeout {
		t.Errorf("MaxConnsPerHost, IdleConnTimeout = %d, %s, want %d, %s",
			transport.MaxConnsPerHost, transport.IdleConnTimeout, httpMaxConnsPerHost, httpIdleConnTimeout)
	}
}
