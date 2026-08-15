package main

import (
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"strings"
)

const callbackPath = "/auth/google/callback"

// LookupEnvは環境変数の取得関数です。
type LookupEnv func(string) (string, bool)

// ConfigはServer限定の設定です。
type Config struct {
	TrustedAppOrigin string
	OAuthCallbackURI string
	GoogleClientID   string
	GoogleSecret     string
	AllowedEmail     string
	DatabaseURL      string
	ProtectionKey    string
}

// loadConfigは設定を検証します。
func loadConfig(lookup LookupEnv) (Config, error) {
	if lookup == nil {
		return Config{}, errors.New("configuration lookup is required")
	}

	origin, err := required(lookup, "TOPIC2HTML_TRUSTED_APP_ORIGIN")
	if err != nil {
		return Config{}, err
	}
	parsedOrigin, err := trustedOrigin(origin)
	if err != nil {
		return Config{}, err
	}
	clientID, err := required(lookup, "TOPIC2HTML_GOOGLE_CLIENT_ID")
	if err != nil {
		return Config{}, err
	}
	clientSecret, err := required(lookup, "TOPIC2HTML_GOOGLE_CLIENT_SECRET")
	if err != nil {
		return Config{}, err
	}
	allowedEmail, err := required(lookup, "TOPIC2HTML_ALLOWED_EMAIL")
	if err != nil {
		return Config{}, err
	}
	if err := exactEmail(allowedEmail); err != nil {
		return Config{}, err
	}
	databaseURL, err := required(lookup, "TOPIC2HTML_DATABASE_URL")
	if err != nil {
		return Config{}, err
	}
	if err := postgresURL(databaseURL); err != nil {
		return Config{}, err
	}
	protectionKey, err := required(lookup, "TOPIC2HTML_PROTECTION_KEY")
	if err != nil {
		return Config{}, err
	}

	callback := *parsedOrigin
	callback.Path = callbackPath
	callback.RawPath = ""
	return Config{TrustedAppOrigin: parsedOrigin.String(), OAuthCallbackURI: callback.String(), GoogleClientID: clientID, GoogleSecret: clientSecret, AllowedEmail: allowedEmail, DatabaseURL: databaseURL, ProtectionKey: protectionKey}, nil
}

func required(lookup LookupEnv, name string) (string, error) {
	value, ok := lookup(name)
	if !ok || strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("%s is required", name)
	}
	return value, nil
}

func trustedOrigin(value string) (*url.URL, error) {
	u, err := url.Parse(value)
	if err != nil || !u.IsAbs() || u.Host == "" || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		return nil, errors.New("trusted app origin must be an absolute origin")
	}
	if u.Scheme == "https" || (u.Scheme == "http" && isLoopbackHost(u.Hostname())) {
		return u, nil
	}
	return nil, errors.New("trusted app origin must use HTTPS outside loopback development")
}

func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func exactEmail(value string) error {
	address, err := mail.ParseAddress(value)
	if err != nil || address.Address != value {
		return errors.New("allowed email must be one exact email address")
	}
	return nil
}

func postgresURL(value string) error {
	u, err := url.Parse(value)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") || u.Host == "" || u.User == nil {
		return errors.New("database URL must be a PostgreSQL connection URL")
	}
	return nil
}
