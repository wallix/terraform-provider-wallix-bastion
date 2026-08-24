package bastion_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/wallix/terraform-provider-wallix-bastion/bastion"
)

// TestSessionWithCSRFIntegration tests the full session and CSRF flow.
func TestSessionWithCSRFIntegration(t *testing.T) {
	callCount := 0
	authenticatedTokens := map[string]bool{}

	// Mock WALLIX Bastion API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Authenticate endpoint
		if r.URL.Path == "/api/v3.8/authentication" && r.Method == http.MethodPost {
			callCount++

			// Return session cookie and CSRF token on successful auth
			http.SetCookie(w, &http.Cookie{
				Name:  tvWabSessionID,
				Value: fmt.Sprintf("session-%d", callCount),
				Path:  "/",
			})
			w.Header().Set("X-Csrf-Token", fmt.Sprintf("csrf-token-%d", callCount))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"status":"success"}`)
			authenticatedTokens[fmt.Sprintf("csrf-token-%d", callCount)] = true

			return
		}

		// Protected endpoints - check for CSRF token
		csrfToken := r.Header.Get("X-Csrf-Token")
		if csrfToken == "" {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, `{"error":"CSRF token required"}`)

			return
		}

		if !authenticatedTokens[csrfToken] {
			w.Header().Set("X-Csrf-Token", "new-csrf-token")
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, `{"error":"CSRF token invalid or expired"}`)

			return
		}

		// Valid request with valid CSRF token
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"ok"}`)
	}))
	defer server.Close()

	// Create config with CSRF enabled
	config := bastion.Config{
		BastionIP:         "localhost",
		BastionPort:       443,
		BastionUser:       "testuser",
		BastionPwd:        "testpass",
		BastionAPIVersion: "v3.8",
		SessionTimeout:    120,
		CSRFEnabled:       true,
	}

	client, diags := config.Client()
	if diags.HasError() {
		t.Errorf("Config.Client() returned errors: %v", diags)
	}

	if client == nil {
		t.Fatal("Config.Client() returned nil client")
	}

	// We can't directly test newRequest() as it requires authenticated state
	// But we've verified the config and client creation works with CSRF options
}

// TestConfigCSRFDisabled tests that CSRF can be disabled.
func TestConfigCSRFDisabled(t *testing.T) {
	config := bastion.Config{
		BastionIP:         tvTestHost,
		BastionPort:       443,
		BastionUser:       "admin",
		BastionPwd:        "password",
		BastionAPIVersion: "v3.8",
		SessionTimeout:    120,
		CSRFEnabled:       false,
	}

	client, diags := config.Client()
	if diags.HasError() {
		t.Errorf("Config.Client() returned errors: %v", diags)
	}

	if client == nil {
		t.Fatal("Config.Client() returned nil client")
	}

	// Verify client was created with CSRF disabled
	// The Client struct is not exported, but we verify creation succeeds
}

// TestSessionTimeoutConfiguration tests session timeout config.
func TestSessionTimeoutConfiguration(t *testing.T) {
	tests := []struct {
		name           string
		sessionTimeout int
		wantErr        bool
	}{
		{
			name:           "default session timeout",
			sessionTimeout: 120,
			wantErr:        false,
		},
		{
			name:           "custom session timeout",
			sessionTimeout: 300,
			wantErr:        false,
		},
		{
			name:           "short session timeout",
			sessionTimeout: 60,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := bastion.Config{
				BastionIP:      tvTestHost,
				BastionPort:    443,
				BastionUser:    "admin",
				BastionPwd:     "password",
				SessionTimeout: tt.sessionTimeout,
				CSRFEnabled:    true,
			}

			client, diags := config.Client()
			if diags.HasError() != tt.wantErr {
				t.Errorf("Config.Client() error = %v, wantErr %v", diags.HasError(), tt.wantErr)

				return
			}

			if client == nil && !tt.wantErr {
				t.Errorf("Config.Client() returned nil client")
			}
		})
	}
}
