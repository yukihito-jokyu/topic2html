package google

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestNewClient(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		transport http.RoundTripper
	}{
		{name: "specified transport", transport: &recordingTransport{}},
		{name: "default transport"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.transport)
			if got, want := client.httpClient.Timeout, RequestTimeout; got != want {
				t.Errorf("timeout = %s, want %s", got, want)
			}
			if tt.transport == nil && client.httpClient.Transport != http.DefaultTransport {
				t.Errorf("default transport was not selected")
			}
		})
	}
}

func TestClientDo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		transport *recordingTransport
		endpoint  string
		wantError bool
		wantCalls int
	}{
		{name: "success", transport: &recordingTransport{}, endpoint: "https://google.test/discovery", wantCalls: 1},
		{name: "transport error", transport: &recordingTransport{err: errors.New("unavailable")}, endpoint: "https://google.test/discovery", wantError: true, wantCalls: 1},
		{name: "invalid endpoint", transport: &recordingTransport{}, endpoint: "://", wantError: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := NewClient(tt.transport).Do(context.Background(), http.MethodGet, tt.endpoint, nil)
			if (err != nil) != tt.wantError {
				t.Fatalf("Do() error = %v, wantError %t", err, tt.wantError)
			}
			if response != nil {
				defer func() { _ = response.Body.Close() }()
			}
			if tt.transport.calls != tt.wantCalls {
				t.Errorf("transport calls = %d, want %d", tt.transport.calls, tt.wantCalls)
			}
		})
	}
}

func TestOIDCRequestConstructionFailure(t *testing.T) {
	tests := []struct {
		name      string
		operation func(*Client) error
	}{
		{
			name: "discovery request",
			operation: func(client *Client) error {
				_, err := client.Discover(context.Background(), "https://google.test/discovery")

				return err
			},
		},
		{
			name: "token request",
			operation: func(client *Client) error {
				_, err := client.ExchangeAuthorizationCode(context.Background(), "https://google.test/token", "client", "secret", "code", "https://app.test/callback", "verifier")

				return err
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(&recordingTransport{})
			client.newRequest = func(context.Context, string, string, io.Reader) (*http.Request, error) {
				return nil, errors.New("request construction failed")
			}
			if err := tt.operation(client); err == nil {
				t.Error("operation error = nil, want error")
			}
		})
	}
}

func TestOIDCRejectsInvalidEndpoints(t *testing.T) {
	tests := []struct {
		name      string
		operation func(*Client) error
	}{
		{
			name: "discovery",
			operation: func(client *Client) error {
				_, err := client.Discover(context.Background(), "://")

				return err
			},
		},
		{
			name: "token",
			operation: func(client *Client) error {
				_, err := client.ExchangeAuthorizationCode(context.Background(), "://", "client", "secret", "code", "https://app.test/callback", "verifier")

				return err
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.operation(NewClient(&recordingTransport{})); err == nil {
				t.Error("operation error = nil, want error")
			}
		})
	}
}

func TestOIDCTransportFailure(t *testing.T) {
	tests := []struct {
		name      string
		operation func(*Client) error
	}{
		{
			name: "discovery",
			operation: func(client *Client) error {
				_, err := client.Discover(context.Background(), "https://google.test/discovery")

				return err
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(&recordingTransport{err: errors.New("unavailable")})
			if err := tt.operation(client); err == nil {
				t.Error("operation error = nil, want error")
			}
		})
	}
}

type recordingTransport struct {
	calls int
	err   error
}

func (t *recordingTransport) RoundTrip(*http.Request) (*http.Response, error) {
	t.calls++
	if t.err != nil {
		return nil, t.err
	}

	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("{}")), Header: make(http.Header)}, nil
}
