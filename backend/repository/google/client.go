package google

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// RequestTimeoutは外部HTTPの制限時間です。
const RequestTimeout = 10 * time.Second

// ClientはGoogleへのHTTP要求を送信します。
type Client struct {
	httpClient *http.Client
	newRequest func(context.Context, string, string, io.Reader) (*http.Request, error)
}

// NewClientはClientを作成します。
func NewClient(transport http.RoundTripper) *Client {
	if transport == nil {
		transport = http.DefaultTransport
	}

	return &Client{
		httpClient: &http.Client{Transport: transport, Timeout: RequestTimeout},
		newRequest: http.NewRequestWithContext,
	}
}

// Doは1回だけHTTP要求を送信します。
func (c *Client) Do(ctx context.Context, method, endpoint string, body io.Reader) (*http.Response, error) {
	req, err := c.newRequest(ctx, method, endpoint, body)
	if err != nil {
		return nil, err
	}

	return c.httpClient.Do(req)
}

// DiscoveryはOIDC endpointの集合です。
type Discovery struct {
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	JWKSURI               string `json:"jwks_uri"`
}

// TokenResponseはToken交換の安全な応答です。
type TokenResponse struct {
	IDToken string `json:"id_token"`
}

// JWKSはID Token検証用の鍵集合です。
type JWKS struct {
	Keys []json.RawMessage `json:"keys"`
}

// DiscoverはOIDC discoveryを一度だけ取得します。
func (c *Client) Discover(ctx context.Context, endpoint string) (Discovery, error) {
	var discovery Discovery
	if err := c.getJSON(ctx, endpoint, &discovery); err != nil {
		return Discovery{}, err
	}
	if err := requireEndpoint(discovery.AuthorizationEndpoint); err != nil {
		return Discovery{}, errors.New("google discovery is invalid")
	}
	if err := requireEndpoint(discovery.TokenEndpoint); err != nil {
		return Discovery{}, errors.New("google discovery is invalid")
	}
	if err := requireEndpoint(discovery.JWKSURI); err != nil {
		return Discovery{}, errors.New("google discovery is invalid")
	}

	return discovery, nil
}

// ExchangeAuthorizationCodeは認可codeを一度だけ交換します。
func (c *Client) ExchangeAuthorizationCode(ctx context.Context, endpoint, clientID, clientSecret, code, redirectURI, verifier string) (TokenResponse, error) {
	if err := requireEndpoint(endpoint); err != nil {
		return TokenResponse{}, errors.New("google token endpoint is invalid")
	}
	form := url.Values{
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"code":          {code},
		"code_verifier": {verifier},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {redirectURI},
	}
	req, err := c.newRequest(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return TokenResponse{}, errors.New("google token request is invalid")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	var token TokenResponse
	if err := c.doJSON(req, &token); err != nil {
		return TokenResponse{}, err
	}
	if strings.TrimSpace(token.IDToken) == "" {
		return TokenResponse{}, errors.New("google token response is invalid")
	}

	return token, nil
}

// FetchJWKSはID Token検証用の鍵集合を一度だけ取得します。
func (c *Client) FetchJWKS(ctx context.Context, endpoint string) (JWKS, error) {
	var jwks JWKS
	if err := c.getJSON(ctx, endpoint, &jwks); err != nil {
		return JWKS{}, err
	}
	if len(jwks.Keys) == 0 {
		return JWKS{}, errors.New("google JWKS response is invalid")
	}

	return jwks, nil
}

func (c *Client) getJSON(ctx context.Context, endpoint string, output any) error {
	if err := requireEndpoint(endpoint); err != nil {
		return errors.New("google endpoint is invalid")
	}
	req, err := c.newRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return errors.New("google request is invalid")
	}

	return c.doJSON(req, output)
}

func (c *Client) doJSON(req *http.Request, output any) error {
	// #nosec G704 -- validated endpoint
	response, err := c.httpClient.Do(req)
	if err != nil {
		return errors.New("google request failed")
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("google returned status %d", response.StatusCode)
	}
	decoder := json.NewDecoder(io.LimitReader(response.Body, 1<<20))
	if err := decoder.Decode(output); err != nil {
		return errors.New("google response is invalid")
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("google response is invalid")
	}

	return nil
}

func requireEndpoint(value string) error {
	endpoint, err := url.Parse(value)
	if err != nil || endpoint.Scheme == "" || endpoint.Host == "" {
		return errors.New("endpoint is invalid")
	}

	return nil
}
