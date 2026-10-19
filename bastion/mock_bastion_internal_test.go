package bastion

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync"
	"testing"
)

// mockSingletonConfig is a minimal Bastion API serving one singleton configuration endpoint
// (GET/PUT, as /config/smtp or /config/x509/crl): any POST /api authenticates, a GET returns the
// current state and a PUT is recorded then merged into the state, except the write-only keys
// that the real API never returns. failPuts makes the next PUTs fail without changing the state.
type mockSingletonConfig struct {
	server    *httptest.Server
	path      string
	writeOnly map[string]bool

	mu        sync.Mutex
	state     map[string]interface{}
	puts      []map[string]interface{}
	requests  []string
	putStatus int
}

func newMockSingletonConfig(
	t *testing.T, uri string, state map[string]interface{}, writeOnly ...string,
) *mockSingletonConfig {
	t.Helper()
	mock := &mockSingletonConfig{
		path:      "/api/" + VersionWallixAPI312 + uri,
		writeOnly: make(map[string]bool),
		state:     state,
	}
	for _, k := range writeOnly {
		mock.writeOnly[k] = true
	}
	mock.server = httptest.NewTLSServer(http.HandlerFunc(mock.serveHTTP))
	t.Cleanup(mock.server.Close)

	return mock
}

func (mock *mockSingletonConfig) serveHTTP(w http.ResponseWriter, r *http.Request) {
	mock.mu.Lock()
	defer mock.mu.Unlock()

	if r.Method == http.MethodPost && r.URL.Path == "/api" {
		http.SetCookie(w, &http.Cookie{Name: "wab_session_id", Value: "mock-session", Path: "/"})
		w.WriteHeader(http.StatusNoContent)

		return
	}
	mock.requests = append(mock.requests, r.Method+" "+r.URL.Path)
	if r.URL.Path != mock.path {
		w.WriteHeader(http.StatusNotFound)

		return
	}
	switch r.Method {
	case http.MethodGet:
		body, err := json.Marshal(mock.state)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)

			return
		}
		w.Header().Set(contentTypeHeader, contentTypeJSON)
		_, _ = w.Write(body)
	case http.MethodPut:
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)

			return
		}
		mock.puts = append(mock.puts, body)
		if mock.putStatus != 0 {
			w.WriteHeader(mock.putStatus)
			_, _ = w.Write([]byte(`{"error":"Bad Request","description":"mock failure"}`))

			return
		}
		for k, v := range body {
			if !mock.writeOnly[k] {
				mock.state[k] = v
			}
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (mock *mockSingletonConfig) client(t *testing.T) *Client {
	t.Helper()
	u, err := url.Parse(mock.server.URL)
	if err != nil {
		t.Fatalf("parsing mock server URL: %v", err)
	}
	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		t.Fatalf("splitting mock server host: %v", err)
	}
	portNum, err := strconv.Atoi(port)
	if err != nil {
		t.Fatalf("parsing mock server port: %v", err)
	}

	return &Client{
		bastionIP:          host,
		bastionPort:        portNum,
		bastionAPIVersion:  VersionWallixAPI312,
		bastionUser:        "mock-user",
		bastionPwd:         "mock-password",
		insecureSkipVerify: true,
	}
}

func (mock *mockSingletonConfig) putBodies() []map[string]interface{} {
	mock.mu.Lock()
	defer mock.mu.Unlock()

	return append([]map[string]interface{}(nil), mock.puts...)
}

func (mock *mockSingletonConfig) setState(state map[string]interface{}) {
	mock.mu.Lock()
	defer mock.mu.Unlock()

	mock.state = state
}

func (mock *mockSingletonConfig) failPuts(status int) {
	mock.mu.Lock()
	defer mock.mu.Unlock()

	mock.putStatus = status
}

func (mock *mockSingletonConfig) requestCount() int {
	mock.mu.Lock()
	defer mock.mu.Unlock()

	return len(mock.requests)
}
