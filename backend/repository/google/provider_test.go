package google

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/yukihito-jokyu/topic2html/backend/apperr"
	domainauth "github.com/yukihito-jokyu/topic2html/backend/domain/auth"
	usecaseauth "github.com/yukihito-jokyu/topic2html/backend/usecase/auth"
)

func TestProviderAuthorizationURL(t *testing.T) {
	server := newProviderServer(t)
	defer server.Close()
	provider := newTestProvider(server)
	got, err := provider.AuthorizationURL(context.Background(), usecaseauth.AuthorizationRequest{
		State:         "state",
		Nonce:         "nonce",
		CodeChallenge: "challenge",
	})
	if err != nil {
		t.Fatalf("AuthorizationURL() error = %v", err)
	}
	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	query := parsed.Query()
	for key, want := range map[string]string{
		"response_type":         "code",
		"client_id":             "client",
		"redirect_uri":          "https://admin.example.test/auth/google/callback",
		"scope":                 "openid email",
		"state":                 "state",
		"nonce":                 "nonce",
		"code_challenge":        "challenge",
		"code_challenge_method": "S256",
	} {
		if query.Get(key) != want {
			t.Errorf("query %s = %q, want %q", key, query.Get(key), want)
		}
	}
	for _, testCase := range []struct {
		name    string
		request usecaseauth.AuthorizationRequest
	}{
		{
			name: "missing state",
		},
		{
			name: "missing nonce",
			request: usecaseauth.AuthorizationRequest{
				State: "state",
			},
		},
		{
			name: "missing code challenge",
			request: usecaseauth.AuthorizationRequest{
				State: "state",
				Nonce: "nonce",
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			if _, err := provider.AuthorizationURL(context.Background(), testCase.request); err == nil {
				t.Errorf("invalid request %+v succeeded", testCase.request)
			}
		})
	}
	for _, testCase := range []struct {
		name      string
		configure func(*Provider)
	}{
		{
			name: "missing client ID",
			configure: func(provider *Provider) {
				provider.config.ClientID = ""
			},
		},
		{
			name: "discovery failure",
			configure: func(provider *Provider) {
				provider.client = NewClient(roundTripperFunc(func(*http.Request) (*http.Response, error) { return nil, fmt.Errorf("network") }))
			},
		},
		{
			name: "invalid authorization endpoint",
			configure: func(provider *Provider) {
				provider.client = oidcTestClient{
					discovery: Discovery{
						AuthorizationEndpoint: "%",
					},
				}
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			localProvider := newTestProvider(server)
			testCase.configure(localProvider)
			if _, err := localProvider.AuthorizationURL(context.Background(), usecaseauth.AuthorizationRequest{
				State:         "s",
				Nonce:         "n",
				CodeChallenge: "c",
			}); err == nil {
				t.Fatal("invalid authorization URL succeeded")
			}
		})
	}
}

func TestProviderExchangeAndVerify(t *testing.T) {
	server := newProviderServer(t)
	defer server.Close()
	provider := newTestProvider(server)
	nonceHash := hashValue("nonce")
	identity, err := provider.ExchangeAndVerify(context.Background(), "code", "verifier", nonceHash)
	if err != nil {
		t.Fatalf("ExchangeAndVerify() error = %v", err)
	}
	if identity.Email != "admin@example.test" || !identity.EmailVerified {
		t.Fatalf("identity = %+v", identity)
	}
	for _, testCase := range []struct {
		name            string
		configure       func(*Provider)
		wantUnavailable bool
	}{
		{
			name: "discovery network failure",
			configure: func(provider *Provider) {
				provider.client = NewClient(roundTripperFunc(func(*http.Request) (*http.Response, error) {
					return nil, errors.New("network")
				}))
			},
			wantUnavailable: true,
		},
		{
			name: "empty discovery issuer",
			configure: func(provider *Provider) {
				provider.client = oidcTestClient{
					discovery: Discovery{},
				}
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			localProvider := newTestProvider(server)
			testCase.configure(localProvider)
			_, err := localProvider.ExchangeAndVerify(context.Background(), "code", "verifier", nonceHash)
			if err == nil || (testCase.wantUnavailable && apperr.CodeOf(err) != apperr.CodeUnavailable) {
				t.Fatalf("error=%v code=%q", err, apperr.CodeOf(err))
			}
		})
	}
	tests := []struct {
		name   string
		mutate func(*providerServer)
		check  func(*Provider)
	}{
		{
			name: "missing client secret",
			check: func(p *Provider) {
				p.config.ClientSecret = ""
			},
		},
		{
			name:  "missing code",
			check: func(p *Provider) {},
		},
		{
			name: "discovery failure",
			mutate: func(s *providerServer) {
				s.discoveryStatus = http.StatusBadGateway
			},
			check: func(p *Provider) {},
		},
		{
			name: "token failure",
			mutate: func(s *providerServer) {
				s.tokenStatus = http.StatusBadGateway
			},
			check: func(p *Provider) {},
		},
		{
			name: "JWKS failure",
			mutate: func(s *providerServer) {
				s.jwksStatus = http.StatusBadGateway
			},
			check: func(p *Provider) {},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			localServer := newProviderServer(t)
			defer localServer.Close()
			localProvider := newTestProvider(localServer)
			if tt.mutate != nil {
				tt.mutate(localServer)
			}
			if tt.check != nil {
				tt.check(localProvider)
			}
			code := "code"
			if tt.name == "missing code" {
				code = ""
			}
			if _, err := localProvider.ExchangeAndVerify(context.Background(), code, "verifier", nonceHash); err == nil {
				t.Fatal("invalid exchange succeeded")
			}
		})
	}
	for _, name := range []string{"bad format", "bad header", "bad payload", "bad claims", "bad nonce", "bad audience", "bad signature", "bad key"} {
		t.Run(name, func(t *testing.T) {
			localServer := newProviderServer(t)
			defer localServer.Close()
			localProvider := newTestProvider(localServer)
			claims := tokenClaimsFor(name)
			claims["iss"] = localServer.server.URL
			switch name {
			case "bad format":
				localServer.token = "bad-token"
			case "bad payload":
				headerBytes, _ := json.Marshal(tokenHeaderFor(name))
				header := base64.RawURLEncoding.EncodeToString(headerBytes)
				localServer.token = header + ".%%%.signature"
			case "bad key":
				header := tokenHeaderFor(name)
				header["kid"] = "missing"
				localServer.token = signedToken(localServer.key, claims, header)
			case "bad signature":
				localServer.token = signedToken(localServer.key, claims, tokenHeaderFor(name))
				localServer.token = tamperSignature(localServer.token)
			default:
				localServer.token = signedToken(localServer.key, claims, tokenHeaderFor(name))
			}
			if _, err := localProvider.ExchangeAndVerify(context.Background(), "code", "verifier", nonceHash); err == nil {
				t.Fatal("invalid token succeeded")
			}
		})
	}
}

func TestProviderHelpers(t *testing.T) {
	defaultProvider := NewProvider(nil, ProviderConfig{})
	if defaultProvider.config.DiscoveryEndpoint != DefaultDiscoveryEndpoint {
		t.Fatalf("default discovery endpoint = %q", defaultProvider.config.DiscoveryEndpoint)
	}
	if got, err := parseAudience(json.RawMessage(`"client"`)); err != nil || len(got) != 1 {
		t.Fatalf("single audience = %#v, %v", got, err)
	}
	if got, err := parseAudience(json.RawMessage(`["client","other"]`)); err != nil || len(got) != 2 {
		t.Fatalf("multiple audience = %#v, %v", got, err)
	}
	for _, testCase := range []struct {
		name string
		raw  string
	}{
		{
			name: "empty",
		},
		{
			name: "empty array",
			raw:  `[]`,
		},
		{
			name: "empty audience element",
			raw:  `["client",""]`,
		},
		{
			name: "object",
			raw:  `{}`,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			if _, err := parseAudience(json.RawMessage(testCase.raw)); err == nil {
				t.Errorf("invalid audience %q succeeded", testCase.raw)
			}
		})
	}
	if !contains([]string{"a", "b"}, "b") || contains([]string{"a"}, "b") {
		t.Fatal("contains result is incorrect")
	}
	if !validAudience([]string{"client"}, "client", "client") {
		t.Fatal("single audience with authorized party was rejected")
	}
	if validAudience([]string{"client", "other"}, "other", "client") {
		t.Fatal("multiple audience with another authorized party was accepted")
	}
	if _, err := findRSAKey(JWKS{
		Keys: []json.RawMessage{json.RawMessage(`{"kty":"EC","kid":"key"}`)},
	}, "key"); err == nil {
		t.Fatal("non-RSA key succeeded")
	}
	for _, testCase := range []struct {
		name string
		raw  string
	}{
		{
			name: "invalid modulus encoding",
			raw:  `{"kty":"RSA","kid":"key","n":"%%%","e":"AQAB"}`,
		},
		{
			name: "invalid exponent encoding",
			raw:  `{"kty":"RSA","kid":"key","n":"AQ","e":"%%%"}`,
		},
		{
			name: "invalid exponent value",
			raw:  `{"kty":"RSA","kid":"key","n":"AQ","e":"Ag"}`,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			if _, err := findRSAKey(JWKS{
				Keys: []json.RawMessage{json.RawMessage(testCase.raw)},
			}, "key"); err == nil {
				t.Errorf("invalid JWK %s succeeded", testCase.raw)
			}
		})
	}
	if got := digestSigningInput("a", "b"); len(got) != sha256.Size {
		t.Fatalf("digest size = %d", len(got))
	}
}

func TestVerifyIDTokenRejectsMalformedHeaderEncoding(t *testing.T) {
	provider := &Provider{
		config: ProviderConfig{
			ClientID: "client",
		},
		now: func() time.Time {
			return time.Now().Add(time.Hour)
		},
	}
	if _, err := provider.verifyIDToken("%%%.payload.signature", Discovery{
		Issuer: "issuer",
	}, JWKS{}, nil); err == nil {
		t.Fatal("malformed header succeeded")
	}
}

type providerServer struct {
	key             *rsa.PrivateKey
	server          *httptest.Server
	token           string
	discoveryStatus int
	tokenStatus     int
	jwksStatus      int
}

func (s *providerServer) Close() { s.server.Close() }

func newProviderServer(t *testing.T) *providerServer {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}
	fixture := &providerServer{
		key: key,
	}
	fixture.server = httptest.NewServer(fixture)
	claims := tokenClaimsFor("valid")
	claims["iss"] = fixture.server.URL
	fixture.token = signedToken(key, claims, tokenHeaderFor("valid"))

	return fixture
}

func (s *providerServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	status := map[string]int{
		"/discovery": s.discoveryStatus,
		"/token":     s.tokenStatus,
		"/jwks":      s.jwksStatus,
	}[request.URL.Path]
	if status != 0 {
		writer.WriteHeader(status)

		return
	}
	writer.Header().Set("Content-Type", "application/json")
	base := s.server.URL
	switch request.URL.Path {
	case "/discovery":
		_, _ = fmt.Fprintf(writer, `{"issuer":"%s","authorization_endpoint":"%s/authorize","token_endpoint":"%s/token","jwks_uri":"%s/jwks"}`, base, base, base, base)
	case "/token":
		_, _ = fmt.Fprintf(writer, `{"id_token":%q}`, s.token)
	case "/jwks":
		modulus := base64.RawURLEncoding.EncodeToString(s.key.N.Bytes())
		exponent := base64.RawURLEncoding.EncodeToString([]byte{1, 0, 1})
		_, _ = fmt.Fprintf(writer, `{"keys":[{"kty":"RSA","kid":"key","n":%q,"e":%q}]}`, modulus, exponent)
	default:
		writer.WriteHeader(http.StatusNotFound)
	}
}

func newTestProvider(server *providerServer) *Provider {
	return NewProvider(NewClient(server.server.Client().Transport), ProviderConfig{
		ClientID:          "client",
		ClientSecret:      "secret",
		RedirectURI:       "https://admin.example.test/auth/google/callback",
		DiscoveryEndpoint: server.server.URL + "/discovery",
	})
}

func hashValue(value string) domainauth.Hash {
	digest := sha256.Sum256([]byte(value))

	return digest[:]
}

func tokenClaimsFor(name string) map[string]any {
	claims := map[string]any{
		"iss":            "issuer-placeholder",
		"aud":            "client",
		"exp":            time.Now().Add(time.Hour).Unix(),
		"nonce":          "nonce",
		"email":          "admin@example.test",
		"email_verified": true,
	}
	switch name {
	case "bad claims":
		claims["email_verified"] = false
	case "bad nonce":
		claims["nonce"] = "wrong"
	case "bad audience":
		claims["aud"] = []string{"other"}
	}

	return claims
}

func tokenHeaderFor(name string) map[string]string {
	header := map[string]string{
		"alg": "RS256",
		"kid": "key",
	}
	if name == "bad header" {
		header["alg"] = "none"
	}

	return header
}

func signedToken(key *rsa.PrivateKey, claims map[string]any, header map[string]string) string {
	// The issuer is replaced by the test server URL after the server starts.
	headerBytes, _ := json.Marshal(header)
	claimBytes, _ := json.Marshal(claims)
	headerValue := base64.RawURLEncoding.EncodeToString(headerBytes)
	payload := base64.RawURLEncoding.EncodeToString(claimBytes)
	digest := sha256.Sum256([]byte(headerValue + "." + payload))
	signature, _ := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])

	return strings.Join([]string{headerValue, payload, base64.RawURLEncoding.EncodeToString(signature)}, ".")
}

func tamperSignature(token string) string {
	parts := strings.Split(token, ".")
	signature, _ := base64.RawURLEncoding.DecodeString(parts[2])
	signature[0] ^= 1
	parts[2] = base64.RawURLEncoding.EncodeToString(signature)

	return strings.Join(parts, ".")
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) { return f(request) }

type oidcTestClient struct {
	discovery Discovery
}

func (c oidcTestClient) Discover(context.Context, string) (Discovery, error) {
	return c.discovery, nil
}

func (oidcTestClient) ExchangeAuthorizationCode(context.Context, string, string, string, string, string, string) (TokenResponse, error) {
	return TokenResponse{}, errors.New("unused token exchange")
}

func (oidcTestClient) FetchJWKS(context.Context, string) (JWKS, error) {
	return JWKS{}, errors.New("unused jwks fetch")
}
