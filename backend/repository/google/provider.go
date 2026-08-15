package google

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"net/url"
	"strings"
	"time"

	"github.com/yukihito-jokyu/topic2html/backend/apperr"
	domainauth "github.com/yukihito-jokyu/topic2html/backend/domain/auth"
	usecaseauth "github.com/yukihito-jokyu/topic2html/backend/usecase/auth"
)

// DefaultDiscoveryEndpointはGoogleのOIDC discovery endpointです。
const DefaultDiscoveryEndpoint = "https://accounts.google.com/.well-known/openid-configuration"

// ProviderConfigはGoogle adapterのServer限定設定です。
type ProviderConfig struct {
	ClientID          string
	ClientSecret      string
	RedirectURI       string
	DiscoveryEndpoint string
}

// ProviderはGoogle OAuth/OIDCの外部境界です。
type Provider struct {
	client oidcClient
	config ProviderConfig
	now    func() time.Time
}

type oidcClient interface {
	Discover(context.Context, string) (Discovery, error)
	ExchangeAuthorizationCode(context.Context, string, string, string, string, string, string) (TokenResponse, error)
	FetchJWKS(context.Context, string) (JWKS, error)
}

// NewProviderはGoogle providerを作成します。
func NewProvider(client oidcClient, config ProviderConfig) *Provider {
	if config.DiscoveryEndpoint == "" {
		config.DiscoveryEndpoint = DefaultDiscoveryEndpoint
	}

	return &Provider{
		client: client,
		config: config,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

// AuthorizationURLは認可要求URLを組み立てます。
func (p *Provider) AuthorizationURL(ctx context.Context, request usecaseauth.AuthorizationRequest) (string, error) {
	if p == nil || p.client == nil || p.config.ClientID == "" || p.config.RedirectURI == "" || request.State == "" || request.Nonce == "" || request.CodeChallenge == "" {
		return "", errors.New("google authorization request is invalid")
	}
	discovery, err := p.client.Discover(ctx, p.config.DiscoveryEndpoint)
	if err != nil {
		return "", apperr.New(apperr.CodeUnavailable)
	}
	endpoint, err := url.Parse(discovery.AuthorizationEndpoint)
	if err != nil || endpoint.Scheme == "" || endpoint.Host == "" {
		return "", errors.New("google authorization endpoint is invalid")
	}
	query := endpoint.Query()
	query.Set("response_type", "code")
	query.Set("client_id", p.config.ClientID)
	query.Set("redirect_uri", p.config.RedirectURI)
	query.Set("scope", "openid email")
	query.Set("state", request.State)
	query.Set("nonce", request.Nonce)
	query.Set("code_challenge", request.CodeChallenge)
	query.Set("code_challenge_method", "S256")
	endpoint.RawQuery = query.Encode()

	return endpoint.String(), nil
}

// ExchangeAndVerifyはcode交換とOIDC ID token検証をServer側で行います。
func (p *Provider) ExchangeAndVerify(ctx context.Context, code, verifier string, expectedNonceHash domainauth.Hash) (domainauth.VerifiedIdentity, error) {
	if p == nil || p.client == nil || p.config.ClientID == "" || p.config.ClientSecret == "" || p.config.RedirectURI == "" || code == "" || verifier == "" {
		return domainauth.VerifiedIdentity{}, errors.New("google exchange request is invalid")
	}
	discovery, err := p.client.Discover(ctx, p.config.DiscoveryEndpoint)
	if err != nil {
		return domainauth.VerifiedIdentity{}, apperr.New(apperr.CodeUnavailable)
	}
	if discovery.Issuer == "" {
		return domainauth.VerifiedIdentity{}, errors.New("google discovery is invalid")
	}
	token, err := p.client.ExchangeAuthorizationCode(ctx, discovery.TokenEndpoint, p.config.ClientID, p.config.ClientSecret, code, p.config.RedirectURI, verifier)
	if err != nil {
		return domainauth.VerifiedIdentity{}, apperr.New(apperr.CodeUnavailable)
	}
	jwks, err := p.client.FetchJWKS(ctx, discovery.JWKSURI)
	if err != nil {
		return domainauth.VerifiedIdentity{}, apperr.New(apperr.CodeUnavailable)
	}

	return p.verifyIDToken(token.IDToken, discovery, jwks, expectedNonceHash)
}

type idTokenHeader struct {
	Algorithm string `json:"alg"`
	KeyID     string `json:"kid"`
}

type idTokenClaims struct {
	Issuer        string          `json:"iss"`
	Audience      json.RawMessage `json:"aud"`
	AuthorizedFor string          `json:"azp"`
	ExpiresAt     int64           `json:"exp"`
	Nonce         string          `json:"nonce"`
	Email         string          `json:"email"`
	EmailVerified bool            `json:"email_verified"`
}

type rsaJWK struct {
	KeyType  string `json:"kty"`
	KeyID    string `json:"kid"`
	Modulus  string `json:"n"`
	Exponent string `json:"e"`
}

func (p *Provider) verifyIDToken(raw string, discovery Discovery, jwks JWKS, expectedNonceHash domainauth.Hash) (domainauth.VerifiedIdentity, error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return domainauth.VerifiedIdentity{}, errors.New("id token format is invalid")
	}
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return domainauth.VerifiedIdentity{}, errors.New("id token header is invalid")
	}
	var header idTokenHeader
	if json.Unmarshal(headerBytes, &header) != nil || header.Algorithm != "RS256" || header.KeyID == "" {
		return domainauth.VerifiedIdentity{}, errors.New("id token header is invalid")
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return domainauth.VerifiedIdentity{}, errors.New("id token claims are invalid")
	}
	var claims idTokenClaims
	if json.Unmarshal(payloadBytes, &claims) != nil || !validClaims(claims, discovery.Issuer, p.now()) {
		return domainauth.VerifiedIdentity{}, errors.New("id token claims are invalid")
	}
	nonceHash := sha256.Sum256([]byte(claims.Nonce))
	if subtle.ConstantTimeCompare(nonceHash[:], expectedNonceHash) != 1 {
		return domainauth.VerifiedIdentity{}, errors.New("id token nonce is invalid")
	}
	audience, err := parseAudience(claims.Audience)
	if err != nil || !validAudience(audience, claims.AuthorizedFor, p.config.ClientID) {
		return domainauth.VerifiedIdentity{}, errors.New("id token audience is invalid")
	}
	key, err := findRSAKey(jwks, header.KeyID)
	if err != nil {
		return domainauth.VerifiedIdentity{}, err
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || rsa.VerifyPKCS1v15(key, crypto.SHA256, digestSigningInput(parts[0], parts[1]), signature) != nil {
		return domainauth.VerifiedIdentity{}, errors.New("id token signature is invalid")
	}

	return domainauth.VerifiedIdentity{
		Email:         claims.Email,
		EmailVerified: true,
	}, nil
}

func validClaims(claims idTokenClaims, issuer string, now time.Time) bool {
	return claims.Issuer == issuer &&
		claims.ExpiresAt > now.Unix() &&
		claims.Nonce != "" &&
		claims.Email != "" &&
		claims.EmailVerified
}

func validAudience(audience []string, authorizedFor, clientID string) bool {
	if !contains(audience, clientID) {
		return false
	}
	if len(audience) > 1 {
		return authorizedFor == clientID
	}

	return authorizedFor == "" || authorizedFor == clientID
}

func parseAudience(raw json.RawMessage) ([]string, error) {
	var one string
	if json.Unmarshal(raw, &one) == nil && one != "" {
		return []string{one}, nil
	}
	var many []string
	if json.Unmarshal(raw, &many) != nil || len(many) == 0 {
		return nil, errors.New("audience is invalid")
	}
	for _, value := range many {
		if value == "" {
			return nil, errors.New("audience is invalid")
		}
	}

	return many, nil
}

func contains(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}

	return false
}

func findRSAKey(jwks JWKS, keyID string) (*rsa.PublicKey, error) {
	for _, raw := range jwks.Keys {
		var jwk rsaJWK
		if json.Unmarshal(raw, &jwk) != nil || jwk.KeyType != "RSA" || jwk.KeyID != keyID {
			continue
		}
		modulus, err := base64.RawURLEncoding.DecodeString(jwk.Modulus)
		if err != nil || len(modulus) == 0 {
			continue
		}
		exponentBytes, err := base64.RawURLEncoding.DecodeString(jwk.Exponent)
		if err != nil || len(exponentBytes) == 0 || len(exponentBytes) > 4 {
			continue
		}
		exponent := 0
		for _, value := range exponentBytes {
			exponent = exponent<<8 | int(value)
		}
		if exponent < 3 || exponent%2 == 0 {
			continue
		}

		return &rsa.PublicKey{
			N: new(big.Int).SetBytes(modulus),
			E: exponent,
		}, nil
	}

	return nil, errors.New("id token signing key is unavailable")
}

func digestSigningInput(header, payload string) []byte {
	digest := sha256.Sum256([]byte(header + "." + payload))

	return digest[:]
}
