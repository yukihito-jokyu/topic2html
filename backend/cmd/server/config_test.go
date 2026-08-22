package main

import (
	"strings"
	"testing"

	"github.com/yukihito-jokyu/topic2html/backend/repository/google"
)

func TestLoadConfig(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		lookup    LookupEnv
		configure func(map[string]string)
		wantURI   string
		wantError bool
	}{
		{
			name:    "valid",
			lookup:  lookup(validEnvironment()),
			wantURI: "https://admin.example.test/auth/google/callback",
		},
		{
			name: "loopback E2E discovery endpoint",
			configure: func(env map[string]string) {
				env["TOPIC2HTML_GOOGLE_DISCOVERY_ENDPOINT"] = "http://127.0.0.1:8181/.well-known/openid-configuration"
			},
			wantURI: "https://admin.example.test/auth/google/callback",
		},
		{
			name: "remote HTTP discovery endpoint",
			configure: func(env map[string]string) {
				env["TOPIC2HTML_GOOGLE_DISCOVERY_ENDPOINT"] = "http://provider.example.test/discovery"
			},
			wantError: true,
		},
		{
			name:      "nil lookup",
			wantError: true,
		},
		{
			name:      "missing secret",
			configure: func(env map[string]string) { delete(env, "TOPIC2HTML_GOOGLE_CLIENT_SECRET") },
			wantError: true,
		},
		{
			name:      "missing client ID",
			configure: func(env map[string]string) { delete(env, "TOPIC2HTML_GOOGLE_CLIENT_ID") },
			wantError: true,
		},
		{
			name:      "missing email",
			configure: func(env map[string]string) { delete(env, "TOPIC2HTML_ALLOWED_EMAIL") },
			wantError: true,
		},
		{
			name:      "missing database URL",
			configure: func(env map[string]string) { delete(env, "TOPIC2HTML_DATABASE_URL") },
			wantError: true,
		},
		{
			name:      "missing protection key",
			configure: func(env map[string]string) { delete(env, "TOPIC2HTML_PROTECTION_KEY") },
			wantError: true,
		},
		{
			name:      "missing broker endpoint",
			configure: func(env map[string]string) { delete(env, "TOPIC2HTML_CODEX_EXECUTION_BROKER_ENDPOINT") },
			wantError: true,
		},
		{
			name: "non unix broker endpoint",
			configure: func(env map[string]string) {
				env["TOPIC2HTML_CODEX_EXECUTION_BROKER_ENDPOINT"] = "tcp://127.0.0.1:8080"
			},
			wantError: true,
		},
		{
			name:      "origin has path",
			configure: func(env map[string]string) { env["TOPIC2HTML_TRUSTED_APP_ORIGIN"] = "https://admin.example.test/admin" },
			wantError: true,
		},
		{
			name:      "non loopback HTTP origin",
			configure: func(env map[string]string) { env["TOPIC2HTML_TRUSTED_APP_ORIGIN"] = "http://admin.example.test" },
			wantError: true,
		},
		{
			name:      "display name email",
			configure: func(env map[string]string) { env["TOPIC2HTML_ALLOWED_EMAIL"] = "Admin <admin@example.test>" },
			wantError: true,
		},
		{
			name: "non postgres database URL",
			configure: func(env map[string]string) {
				env["TOPIC2HTML_DATABASE_URL"] = "mysql://user:password@db.example.test/app"
			},
			wantError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lookupEnv := tt.lookup
			if tt.configure != nil {
				env := validEnvironment()
				tt.configure(env)
				lookupEnv = lookup(env)
			}
			cfg, err := loadConfig(lookupEnv)
			if (err != nil) != tt.wantError {
				t.Fatalf("loadConfig() error = %v, wantError %t", err, tt.wantError)
			}
			if !tt.wantError && cfg.OAuthCallbackURI != tt.wantURI {
				t.Errorf("OAuthCallbackURI = %q, want %q", cfg.OAuthCallbackURI, tt.wantURI)
			}
			if !tt.wantError && tt.configure == nil && cfg.GoogleDiscoveryEndpoint != google.DefaultDiscoveryEndpoint {
				t.Errorf("GoogleDiscoveryEndpoint = %q, want %q", cfg.GoogleDiscoveryEndpoint, google.DefaultDiscoveryEndpoint)
			}
		})
	}
}

func TestOptionalLoopbackHTTPEndpoint(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		value     string
		present   bool
		want      string
		wantError bool
	}{
		{
			name: "default",
			want: "https://accounts.google.com/.well-known/openid-configuration",
		},
		{
			name:    "HTTPS endpoint",
			value:   "https://provider.example.test/discovery",
			present: true,
			want:    "https://provider.example.test/discovery",
		},
		{
			name:    "loopback HTTP endpoint",
			value:   "http://localhost:8181/discovery",
			present: true,
			want:    "http://localhost:8181/discovery",
		},
		{
			name:      "remote HTTP endpoint",
			value:     "http://provider.example.test/discovery",
			present:   true,
			wantError: true,
		},
		{
			name:      "query",
			value:     "https://provider.example.test/discovery?test=true",
			present:   true,
			wantError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := optionalLoopbackHTTPEndpoint(func(string) (string, bool) { return tt.value, tt.present }, "SETTING", "https://accounts.google.com/.well-known/openid-configuration")
			if (err != nil) != tt.wantError {
				t.Fatalf("optionalLoopbackHTTPEndpoint() error = %v, wantError %t", err, tt.wantError)
			}
			if !tt.wantError && got != tt.want {
				t.Errorf("optionalLoopbackHTTPEndpoint() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLoadConfigDoesNotExposeValues(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{
			name:  "database URL",
			key:   "TOPIC2HTML_DATABASE_URL",
			value: "only-for-test-not-a-production-secret",
		},
		{
			name:  "origin",
			key:   "TOPIC2HTML_TRUSTED_APP_ORIGIN",
			value: "http://admin.example.test/private",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := validEnvironment()
			env[tt.key] = tt.value
			_, err := loadConfig(lookup(env))
			if err == nil {
				t.Fatal("loadConfig() succeeded, want error")
			}
			if strings.Contains(err.Error(), tt.value) {
				t.Fatalf("error exposed a configuration value: %v", err)
			}
		})
	}
}

func TestRequired(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		value     string
		present   bool
		wantError bool
	}{
		{
			name:    "value",
			value:   "configured",
			present: true,
		},
		{
			name:      "missing",
			wantError: true,
		},
		{
			name:      "blank",
			value:     " \t",
			present:   true,
			wantError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := required(func(string) (string, bool) { return tt.value, tt.present }, "SETTING")
			if (err != nil) != tt.wantError {
				t.Fatalf("required() error = %v, wantError %t", err, tt.wantError)
			}
			if got != tt.value && !tt.wantError {
				t.Errorf("required() = %q, want %q", got, tt.value)
			}
		})
	}
}

func TestTrustedOrigin(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		value     string
		want      string
		wantError bool
	}{
		{
			name:  "HTTPS",
			value: "https://admin.example.test",
			want:  "https://admin.example.test",
		},
		{
			name:  "loopback hostname",
			value: "http://localhost:8080",
			want:  "http://localhost:8080",
		},
		{
			name:  "loopback IPv4",
			value: "http://127.0.0.1:8080",
			want:  "http://127.0.0.1:8080",
		},
		{
			name:      "invalid",
			value:     "://",
			wantError: true,
		},
		{
			name:      "relative",
			value:     "/admin",
			wantError: true,
		},
		{
			name:      "without host",
			value:     "https:",
			wantError: true,
		},
		{
			name:      "user info",
			value:     "https://user@admin.example.test",
			wantError: true,
		},
		{
			name:      "path",
			value:     "https://admin.example.test/admin",
			wantError: true,
		},
		{
			name:      "query",
			value:     "https://admin.example.test?next=admin",
			wantError: true,
		},
		{
			name:      "fragment",
			value:     "https://admin.example.test#admin",
			wantError: true,
		},
		{
			name:      "remote HTTP",
			value:     "http://admin.example.test",
			wantError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := trustedOrigin(tt.value)
			if (err != nil) != tt.wantError {
				t.Fatalf("trustedOrigin() error = %v, wantError %t", err, tt.wantError)
			}
			if !tt.wantError && got.String() != tt.want {
				t.Errorf("trustedOrigin() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsLoopbackHost(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		host string
		want bool
	}{
		{
			name: "localhost",
			host: "localhost",
			want: true,
		},
		{
			name: "mixed case localhost",
			host: "LOCALHOST",
			want: true,
		},
		{
			name: "IPv4",
			host: "127.0.0.1",
			want: true,
		},
		{
			name: "IPv6",
			host: "::1",
			want: true,
		},
		{
			name: "remote",
			host: "admin.example.test",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isLoopbackHost(tt.host); got != tt.want {
				t.Errorf("isLoopbackHost(%q) = %t, want %t", tt.host, got, tt.want)
			}
		})
	}
}

func TestExactEmail(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		value     string
		wantError bool
	}{
		{
			name:  "valid",
			value: "admin@example.test",
		},
		{
			name:      "invalid",
			value:     "admin",
			wantError: true,
		},
		{
			name:      "display name",
			value:     "Admin <admin@example.test>",
			wantError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := exactEmail(tt.value); (err != nil) != tt.wantError {
				t.Errorf("exactEmail(%q) error = %v, wantError %t", tt.value, err, tt.wantError)
			}
		})
	}
}

func TestPostgresURL(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		value     string
		wantError bool
	}{
		{
			name:  "postgres",
			value: "postgres://app:password@db.example.test/app", // #nosec G101 -- test-only DSN
		},
		{
			name:  "postgresql",
			value: "postgresql://app:password@db.example.test/app", // #nosec G101 -- test-only DSN
		},
		{
			name:      "invalid",
			value:     "://",
			wantError: true,
		},
		{
			name:      "wrong scheme",
			value:     "mysql://app:password@db.example.test/app", // #nosec G101 -- test-only DSN
			wantError: true,
		},
		{
			name:      "without host",
			value:     "postgres://app:password@/app",
			wantError: true,
		},
		{
			name:      "without user",
			value:     "postgres://db.example.test/app",
			wantError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := postgresURL(tt.value); (err != nil) != tt.wantError {
				t.Errorf("postgresURL(%q) error = %v, wantError %t", tt.value, err, tt.wantError)
			}
		})
	}
}

func TestPrivateUnixEndpoint(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		value     string
		wantError bool
	}{
		{
			name:  "absolute socket",
			value: "unix:///var/run/topic2html/broker.sock",
		},
		{
			name:      "network endpoint",
			value:     "tcp://127.0.0.1:8080",
			wantError: true,
		},
		{
			name:      "relative socket",
			value:     "unix://broker.sock",
			wantError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := privateUnixEndpoint(tt.value); (err != nil) != tt.wantError {
				t.Errorf("privateUnixEndpoint(%q) error = %v, wantError %t", tt.value, err, tt.wantError)
			}
		})
	}
}

func validEnvironment() map[string]string {
	// #nosec G101 -- test fixture
	return map[string]string{
		"TOPIC2HTML_TRUSTED_APP_ORIGIN":              "https://admin.example.test",
		"TOPIC2HTML_GOOGLE_CLIENT_ID":                "test-client-id",
		"TOPIC2HTML_GOOGLE_CLIENT_SECRET":            "test-client-secret",
		"TOPIC2HTML_ALLOWED_EMAIL":                   "admin@example.test",
		"TOPIC2HTML_DATABASE_URL":                    "postgres://app:password@db.example.test:5432/topic2html",
		"TOPIC2HTML_PROTECTION_KEY":                  "test-protection-key",
		"TOPIC2HTML_CODEX_EXECUTION_BROKER_ENDPOINT": "unix:///var/run/topic2html/broker.sock",
	}
}

func lookup(values map[string]string) LookupEnv {
	return func(key string) (string, bool) {
		value, ok := values[key]

		return value, ok
	}
}
