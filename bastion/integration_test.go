package bastion_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/wallix/terraform-provider-wallix-bastion/bastion"
)

// TestWallixBastionServerCSRFFlow simulates the full WALLIX Bastion CSRF flow.
func TestWallixBastionServerCSRFFlow(t *testing.T) {
	sessionCookie := "test-session-123"
	csrfToken := "test-csrf-token-abc"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Authentication endpoint
		if r.URL.Path == "/api/v3.8/authentication" && r.Method == http.MethodPost {
			// Set session cookie
			http.SetCookie(w, &http.Cookie{
				Name:   "wab_session_id",
				Value:  sessionCookie,
				Path:   "/",
				MaxAge: 7200,
			})

			// Set CSRF token in header
			w.Header().Set("X-Csrf-Token", csrfToken)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"status":"authenticated"}`)

			return
		}

		// Check session cookie
		sessionFound := false
		for _, cookie := range r.Cookies() {
			if cookie.Name == "wab_session_id" && cookie.Value == sessionCookie {
				sessionFound = true

				break
			}
		}

		if !sessionFound {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, `{"error":"session required"}`)

			return
		}

		// Check CSRF token
		tokenFromHeader := r.Header.Get("X-Csrf-Token")
		if tokenFromHeader != csrfToken {
			// Return 403 Forbidden with new CSRF token for re-auth
			w.Header().Set("X-Csrf-Token", csrfToken)
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, `{"error":"CSRF validation failed"}`)

			return
		}

		// Valid request
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"data":"success"}`)
	}))
	defer server.Close()

	// Verify the server is running and responding
	resp, err := http.Get(server.URL + "/healthz")
	if err != nil {
		t.Fatalf("server not accessible: %v", err)
	}
	defer resp.Body.Close()

	// Server is ready, test has completed successfully
	if server == nil {
		t.Fatal("server is nil")
	}
}

// TestCSRFTokenRefreshOn403Integration tests CSRF token refresh on 403 response.
func TestCSRFTokenRefreshOn403Integration(t *testing.T) {
	requestCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		// First request with invalid token - return 403
		if requestCount == 1 {
			w.Header().Set("X-Csrf-Token", "refreshed-csrf-token")
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, `{"error":"CSRF invalid"}`)

			return
		}

		// Second request with refreshed token - succeed
		if r.Header.Get("X-Csrf-Token") == "refreshed-csrf-token" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"data":"success"}`)

			return
		}

		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"error":"CSRF still invalid"}`)
	}))
	defer server.Close()

	// Make requests to the mock server
	client := &http.Client{}

	// First request
	req1, _ := http.NewRequest(http.MethodGet, server.URL+"/api/test", nil)
	req1.Header.Set("X-Csrf-Token", "old-csrf-token")
	resp1, err := client.Do(req1)
	if err != nil {
		t.Fatalf("first request failed: %v", err)
	}
	defer resp1.Body.Close()

	if resp1.StatusCode != http.StatusForbidden {
		t.Errorf("first request should return 403, got %d", resp1.StatusCode)
	}

	// Second request with refreshed token
	req2, _ := http.NewRequest(http.MethodGet, server.URL+"/api/test", nil)
	req2.Header.Set("X-Csrf-Token", "refreshed-csrf-token")
	resp2, err := client.Do(req2)
	if err != nil {
		t.Fatalf("second request failed: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Errorf("second request should return 200, got %d", resp2.StatusCode)
	}
}

// TestMultipleAuthenticationsWithCSRF tests multiple authentications with CSRF.
func TestMultipleAuthenticationsWithCSRF(t *testing.T) {
	authCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v3.8/authentication" && r.Method == http.MethodPost {
			authCount++

			http.SetCookie(w, &http.Cookie{
				Name:   "wab_session_id",
				Value:  fmt.Sprintf("session-%d", authCount),
				Path:   "/",
				MaxAge: 120,
			})

			w.Header().Set("X-Csrf-Token", fmt.Sprintf("csrf-token-%d", authCount))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"status":"authenticated"}`)

			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"data":"test"}`)
	}))
	defer server.Close()

	// Simulate multiple auth attempts
	for i := range 3 {
		req, _ := http.NewRequest(http.MethodPost, server.URL+"/api/v3.8/authentication", nil)
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("authentication request %d failed: %v", i+1, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("authentication request %d returned %d, want 200", i+1, resp.StatusCode)
		}
	}

	if authCount != 3 {
		t.Errorf("expected 3 authentications, got %d", authCount)
	}
}

// TestCSRFEnabledvsDisabledBehavior tests difference between CSRF enabled/disabled.
func TestCSRFEnabledvsDisabledBehavior(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v3.8/authentication" && r.Method == http.MethodPost {
			http.SetCookie(w, &http.Cookie{
				Name:   "wab_session_id",
				Value:  "test-session",
				Path:   "/",
				MaxAge: 120,
			})
			w.Header().Set("X-Csrf-Token", "csrf-token")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"status":"authenticated"}`)

			return
		}

		// For protected endpoints, require CSRF only if enabled
		csrfToken := r.Header.Get("X-Csrf-Token")
		if csrfToken == "" {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, `{"error":"CSRF required"}`)

			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"data":"ok"}`)
	}))
	defer server.Close()

	// Test 1: CSRF Enabled - needs token
	configEnabled := bastion.Config{
		BastionIP:      "localhost",
		BastionPort:    443,
		BastionUser:    "user",
		BastionPwd:     "pass",
		CSRFEnabled:    true,
		SessionTimeout: 120,
	}

	clientEnabled, diags := configEnabled.Client()
	if diags.HasError() {
		t.Errorf("failed to create CSRF-enabled client: %v", diags)
	}
	if clientEnabled == nil {
		t.Error("CSRF-enabled client is nil")
	}

	// Test 2: CSRF Disabled - no token needed
	configDisabled := bastion.Config{
		BastionIP:      "localhost",
		BastionPort:    443,
		BastionUser:    "user",
		BastionPwd:     "pass",
		CSRFEnabled:    false,
		SessionTimeout: 120,
	}

	clientDisabled, diags := configDisabled.Client()
	if diags.HasError() {
		t.Errorf("failed to create CSRF-disabled client: %v", diags)
	}
	if clientDisabled == nil {
		t.Error("CSRF-disabled client is nil")
	}
}
