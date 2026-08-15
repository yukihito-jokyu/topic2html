package google

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestOIDCBoundaryWithTestDouble(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		configure func(*googleTestDouble)
		operation func(*Client, string) error
		wantError bool
		wantCalls map[string]int
		wantToken bool
		wantJWKS  bool
	}{
		{
			name: "discovery token and JWKS",
			operation: func(client *Client, endpoint string) error {
				discovery, err := client.Discover(context.Background(), endpoint+"/discovery")
				if err != nil {
					return err
				}
				token, err := client.ExchangeAuthorizationCode(context.Background(), discovery.TokenEndpoint, "test-client", "test-secret", "test-code", "https://app.test/auth/google/callback", "test-verifier")
				if err != nil {
					return err
				}
				if token.IDToken != "test-id-token" {
					return errUnexpectedResponse
				}
				jwks, err := client.FetchJWKS(context.Background(), discovery.JWKSURI)
				if err != nil {
					return err
				}
				if len(jwks.Keys) != 1 {
					return errUnexpectedResponse
				}

				return nil
			},
			wantCalls: map[string]int{
				"/discovery": 1,
				"/token":     1,
				"/jwks":      1,
			},
			wantToken: true,
			wantJWKS:  true,
		},
		{
			name: "invalid authorization endpoint is rejected",
			configure: func(server *googleTestDouble) {
				server.discoveryBody = `{"authorization_endpoint":"","token_endpoint":"https://google.test/token","jwks_uri":"https://google.test/jwks"}`
			},
			operation: func(client *Client, endpoint string) error {
				_, err := client.Discover(context.Background(), endpoint+"/discovery")

				return err
			},
			wantError: true,
			wantCalls: map[string]int{
				"/discovery": 1,
			},
		},
		{
			name: "invalid token endpoint is rejected",
			configure: func(server *googleTestDouble) {
				server.discoveryBody = `{"authorization_endpoint":"https://google.test/authorize","token_endpoint":"","jwks_uri":"https://google.test/jwks"}`
			},
			operation: func(client *Client, endpoint string) error {
				_, err := client.Discover(context.Background(), endpoint+"/discovery")

				return err
			},
			wantError: true,
			wantCalls: map[string]int{
				"/discovery": 1,
			},
		},
		{
			name: "invalid JWKS endpoint is rejected",
			configure: func(server *googleTestDouble) {
				server.discoveryBody = `{"authorization_endpoint":"https://google.test/authorize","token_endpoint":"https://google.test/token","jwks_uri":""}`
			},
			operation: func(client *Client, endpoint string) error {
				_, err := client.Discover(context.Background(), endpoint+"/discovery")

				return err
			},
			wantError: true,
			wantCalls: map[string]int{
				"/discovery": 1,
			},
		},
		{
			name:      "discovery failure is not retried",
			configure: func(server *googleTestDouble) { server.status["/discovery"] = http.StatusServiceUnavailable },
			operation: func(client *Client, endpoint string) error {
				_, err := client.Discover(context.Background(), endpoint+"/discovery")

				return err
			},
			wantError: true,
			wantCalls: map[string]int{
				"/discovery": 1,
			},
		},
		{
			name:      "invalid token JSON is rejected",
			configure: func(server *googleTestDouble) { server.tokenBody = `{` },
			operation: func(client *Client, endpoint string) error {
				_, err := client.ExchangeAuthorizationCode(context.Background(), endpoint+"/token", "test-client", "test-secret", "test-code", "https://app.test/auth/google/callback", "test-verifier")

				return err
			},
			wantError: true,
			wantCalls: map[string]int{
				"/token": 1,
			},
			wantToken: true,
		},
		{
			name:      "trailing token JSON is rejected",
			configure: func(server *googleTestDouble) { server.tokenBody = `{"id_token":"test-id-token"}{}` },
			operation: func(client *Client, endpoint string) error {
				_, err := client.ExchangeAuthorizationCode(context.Background(), endpoint+"/token", "test-client", "test-secret", "test-code", "https://app.test/auth/google/callback", "test-verifier")

				return err
			},
			wantError: true,
			wantCalls: map[string]int{
				"/token": 1,
			},
			wantToken: true,
		},
		{
			name:      "invalid token response is rejected",
			configure: func(server *googleTestDouble) { server.tokenBody = `{}` },
			operation: func(client *Client, endpoint string) error {
				_, err := client.ExchangeAuthorizationCode(context.Background(), endpoint+"/token", "test-client", "test-secret", "test-code", "https://app.test/auth/google/callback", "test-verifier")

				return err
			},
			wantError: true,
			wantCalls: map[string]int{
				"/token": 1,
			},
			wantToken: true,
		},
		{
			name:      "JWKS failure is not retried",
			configure: func(server *googleTestDouble) { server.status["/jwks"] = http.StatusBadGateway },
			operation: func(client *Client, endpoint string) error {
				_, err := client.FetchJWKS(context.Background(), endpoint+"/jwks")

				return err
			},
			wantError: true,
			wantCalls: map[string]int{
				"/jwks": 1,
			},
			wantJWKS: true,
		},
		{
			name:      "invalid JWKS response is rejected",
			configure: func(server *googleTestDouble) { server.jwksBody = `{"keys":[]}` },
			operation: func(client *Client, endpoint string) error {
				_, err := client.FetchJWKS(context.Background(), endpoint+"/jwks")

				return err
			},
			wantError: true,
			wantCalls: map[string]int{
				"/jwks": 1,
			},
			wantJWKS: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newGoogleTestDouble()
			if tt.configure != nil {
				tt.configure(server)
			}
			httpServer := httptest.NewServer(server)
			t.Cleanup(httpServer.Close)
			err := tt.operation(NewClient(httpServer.Client().Transport), httpServer.URL)
			if (err != nil) != tt.wantError {
				t.Fatalf("operation error = %v, wantError %t", err, tt.wantError)
			}
			for path, wantCalls := range tt.wantCalls {
				if gotCalls := server.calls[path]; gotCalls != wantCalls {
					t.Errorf("%s calls = %d, want %d", path, gotCalls, wantCalls)
				}
			}
			if tt.wantToken && !server.validTokenRequest {
				t.Error("token request did not match the contract")
			}
			if tt.wantJWKS && server.calls["/jwks"] > 0 && server.lastMethod["/jwks"] != http.MethodGet {
				t.Errorf("JWKS method = %s, want GET", server.lastMethod["/jwks"])
			}
		})
	}
}

var errUnexpectedResponse = &unexpectedResponseError{}

type unexpectedResponseError struct{}

func (*unexpectedResponseError) Error() string { return "unexpected response" }

type googleTestDouble struct {
	calls             map[string]int
	lastMethod        map[string]string
	status            map[string]int
	tokenBody         string
	jwksBody          string
	discoveryBody     string
	validTokenRequest bool
}

func newGoogleTestDouble() *googleTestDouble {
	// #nosec G101 -- test fixture
	return &googleTestDouble{
		calls:      make(map[string]int),
		lastMethod: make(map[string]string),
		status:     make(map[string]int),
		tokenBody:  `{"id_token":"test-id-token"}`,
		jwksBody:   `{"keys":[{"kty":"RSA","kid":"test-key"}]}`,
	}
}

func (server *googleTestDouble) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	server.calls[request.URL.Path]++
	server.lastMethod[request.URL.Path] = request.Method
	if status := server.status[request.URL.Path]; status != 0 {
		writer.WriteHeader(status)

		return
	}
	writer.Header().Set("Content-Type", "application/json")
	switch request.URL.Path {
	case "/discovery":
		body := server.discoveryBody
		if body == "" {
			body = `{"authorization_endpoint":"` + serverURL(request) + `/authorize","token_endpoint":"` + serverURL(request) + `/token","jwks_uri":"` + serverURL(request) + `/jwks"}`
		}
		// #nosec G705 -- fixture body
		_, _ = io.WriteString(writer, body)
	case "/token":
		body, _ := io.ReadAll(request.Body)
		form, _ := url.ParseQuery(string(body))
		server.validTokenRequest = request.Method == http.MethodPost && request.Header.Get("Content-Type") == "application/x-www-form-urlencoded" && form.Get("grant_type") == "authorization_code" && form.Get("client_id") == "test-client" && form.Get("client_secret") == "test-secret" && form.Get("code") == "test-code" && form.Get("redirect_uri") == "https://app.test/auth/google/callback" && form.Get("code_verifier") == "test-verifier"
		_, _ = io.WriteString(writer, server.tokenBody)
	case "/jwks":
		_, _ = io.WriteString(writer, server.jwksBody)
	default:
		writer.WriteHeader(http.StatusNotFound)
	}
}

func serverURL(request *http.Request) string {
	return "http://" + request.Host
}

var _ http.Handler = (*googleTestDouble)(nil)
