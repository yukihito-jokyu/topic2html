package auth

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"testing"
	"time"

	"github.com/yukihito-jokyu/topic2html/backend/apperr"
	domainauth "github.com/yukihito-jokyu/topic2html/backend/domain/auth"
)

type oauthTestStore struct {
	replaceErr  error
	consume     domainauth.OAuthTransaction
	found       bool
	consumeErr  error
	createErr   error
	replaced    bool
	created     bool
	previous    domainauth.Hash
	createdData domainauth.AdminSession
}

func (s *oauthTestStore) ReplaceOAuthTransaction(_ context.Context, previous domainauth.Hash, _ domainauth.OAuthTransaction) error {
	s.previous = previous
	if s.replaceErr != nil {
		return s.replaceErr
	}
	s.replaced = true

	return nil
}

func (s *oauthTestStore) ConsumeOAuthTransaction(context.Context, domainauth.Hash, domainauth.Hash, domainauth.Time) (domainauth.OAuthTransaction, bool, error) {
	return s.consume, s.found, s.consumeErr
}

func (s *oauthTestStore) CreateAdminSession(_ context.Context, session domainauth.AdminSession) error {
	s.createdData = session
	if s.createErr != nil {
		return s.createErr
	}
	s.created = true

	return nil
}

func (*oauthTestStore) FindAdminSession(context.Context, domainauth.Hash) (domainauth.AdminSession, bool, error) {
	return domainauth.AdminSession{}, false, nil
}
func (*oauthTestStore) RevokeAdminSession(context.Context, domainauth.Hash, domainauth.Time) (bool, error) {
	return false, nil
}
func (*oauthTestStore) RunAuthorizedAdminStateChange(context.Context, domainauth.Hash, string, domainauth.Hash, domainauth.Time, AdminStateChangeOperation) (bool, error) {
	return false, nil
}
func (*oauthTestStore) DeleteExpiredProtectedRecords(context.Context, domainauth.Time) error {
	return nil
}

type oauthTestProvider struct {
	authorizationURL string
	authorizationErr error
	identity         domainauth.VerifiedIdentity
	exchangeErr      error
	gotCode          string
	gotVerifier      string
	gotNonceHash     domainauth.Hash
}

func (p *oauthTestProvider) AuthorizationURL(context.Context, AuthorizationRequest) (string, error) {
	return p.authorizationURL, p.authorizationErr
}

func (p *oauthTestProvider) ExchangeAndVerify(_ context.Context, code, verifier string, nonceHash domainauth.Hash) (domainauth.VerifiedIdentity, error) {
	p.gotCode, p.gotVerifier, p.gotNonceHash = code, verifier, nonceHash

	return p.identity, p.exchangeErr
}

type oauthTestSecurity struct {
	randomCalls int
	randomFail  int
	shortResult int
	sealErr     error
	openErr     error
	openValue   []byte
}

type oauthTestLogger struct{}

func (oauthTestLogger) Info(context.Context, string)         {}
func (oauthTestLogger) Error(context.Context, string, error) {}

func (s *oauthTestSecurity) RandomBytes(size int) ([]byte, error) {
	s.randomCalls++
	if s.randomFail == s.randomCalls {
		return nil, errors.New("random failure")
	}
	length := size
	if s.shortResult == s.randomCalls {
		length--
	}
	value := make([]byte, length)
	for index := range value {
		// #nosec G115 -- the deterministic test value is intentionally truncated to a byte.
		value[index] = byte(s.randomCalls)
	}

	return value, nil
}

func (*oauthTestSecurity) Hash(value []byte) domainauth.Hash {
	digest := sha256.Sum256(value)

	return append(domainauth.Hash(nil), digest[:]...)
}

func (s *oauthTestSecurity) Seal([]byte) (domainauth.Ciphertext, error) {
	if s.sealErr != nil {
		return nil, s.sealErr
	}

	return domainauth.Ciphertext("sealed"), nil
}

func (s *oauthTestSecurity) Open(domainauth.Ciphertext) ([]byte, error) {
	if s.openErr != nil {
		return nil, s.openErr
	}

	return s.openValue, nil
}

type oauthTestClock struct{ now time.Time }

func (c oauthTestClock) Now() domainauth.Time { return c.now }

func newOAuthService(store *oauthTestStore, provider *oauthTestProvider, security *oauthTestSecurity) (*Service, time.Time) {
	now := time.Date(2026, 8, 15, 12, 0, 0, 0, time.UTC)
	service, err := NewService(Dependencies{
		Store:    store,
		Provider: provider,
		Security: security,
		Clock: oauthTestClock{
			now: now,
		},
		Logger: oauthTestLogger{},
	}, "https://admin.example.test", "admin@example.test")
	if err != nil {
		panic(err)
	}

	return service, now
}

func validOAuthRecord(now time.Time) domainauth.OAuthTransaction {
	return domainauth.OAuthTransaction{
		ID:                     "00000000-0000-4000-8000-000000000001",
		ReferenceHash:          domainauth.Hash{1},
		StateHash:              domainauth.Hash{2},
		NonceHash:              domainauth.Hash{3},
		PKCEVerifierCiphertext: domainauth.Ciphertext("sealed"),
		ReturnPath:             "/admin",
		CreatedAt:              now,
		ExpiresAt:              now.Add(domainauth.OAuthTransactionLifetime),
	}
}

func TestServiceStart(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name       string
		input      StartInput
		configure  func(*oauthTestStore, *oauthTestProvider, *oauthTestSecurity)
		wantError  bool
		wantStored bool
	}{
		{
			name: "success",
			input: StartInput{
				Origins: []string{"https://admin.example.test"},
			},
			wantStored: true,
		},
		{
			name: "explicit return path and replacement",
			input: StartInput{
				Origins:           []string{"https://admin.example.test"},
				ReturnPaths:       []string{"/admin"},
				PreviousReference: "old",
			},
			wantStored: true,
		},
		{
			name: "origin mismatch",
			input: StartInput{
				Origins: []string{"https://evil.example.test"},
			},
			wantError: true,
		},
		{
			name: "multiple origins",
			input: StartInput{
				Origins: []string{"https://admin.example.test", "https://admin.example.test"},
			},
			wantError: true,
		},
		{
			name: "invalid return path",
			input: StartInput{
				Origins:     []string{"https://admin.example.test"},
				ReturnPaths: []string{"/external"},
			},
			wantError: true,
		},
		{
			name: "duplicate return path",
			input: StartInput{
				Origins:     []string{"https://admin.example.test"},
				ReturnPaths: []string{"/admin", "/admin"},
			},
			wantError: true,
		},
		{
			name: "random failure",
			input: StartInput{
				Origins: []string{"https://admin.example.test"},
			},
			configure: func(_ *oauthTestStore, _ *oauthTestProvider, security *oauthTestSecurity) { security.randomFail = 3 },
			wantError: true,
		},
		{
			name: "state random failure",
			input: StartInput{
				Origins: []string{"https://admin.example.test"},
			},
			configure: func(_ *oauthTestStore, _ *oauthTestProvider, security *oauthTestSecurity) { security.randomFail = 2 },
			wantError: true,
		},
		{
			name: "verifier random failure",
			input: StartInput{
				Origins: []string{"https://admin.example.test"},
			},
			configure: func(_ *oauthTestStore, _ *oauthTestProvider, security *oauthTestSecurity) { security.randomFail = 4 },
			wantError: true,
		},
		{
			name: "transaction ID random failure",
			input: StartInput{
				Origins: []string{"https://admin.example.test"},
			},
			configure: func(_ *oauthTestStore, _ *oauthTestProvider, security *oauthTestSecurity) { security.randomFail = 5 },
			wantError: true,
		},
		{
			name: "seal failure",
			input: StartInput{
				Origins: []string{"https://admin.example.test"},
			},
			configure: func(_ *oauthTestStore, _ *oauthTestProvider, security *oauthTestSecurity) {
				security.sealErr = errors.New("seal")
			},
			wantError: true,
		},
		{
			name: "provider failure",
			input: StartInput{
				Origins: []string{"https://admin.example.test"},
			},
			configure: func(_ *oauthTestStore, provider *oauthTestProvider, _ *oauthTestSecurity) {
				provider.authorizationErr = errors.New("provider")
			},
			wantError: true,
		},
		{
			name: "empty provider URL",
			input: StartInput{
				Origins: []string{"https://admin.example.test"},
			},
			configure: func(_ *oauthTestStore, provider *oauthTestProvider, _ *oauthTestSecurity) {
				provider.authorizationURL = ""
			},
			wantError: true,
		},
		{
			name: "store failure",
			input: StartInput{
				Origins: []string{"https://admin.example.test"},
			},
			configure: func(store *oauthTestStore, _ *oauthTestProvider, _ *oauthTestSecurity) {
				store.replaceErr = errors.New("store")
			},
			wantError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &oauthTestStore{}
			provider := &oauthTestProvider{
				authorizationURL: "https://accounts.google.test/authorize",
			}
			security := &oauthTestSecurity{}
			service, now := newOAuthService(store, provider, security)
			if tt.configure != nil {
				tt.configure(store, provider, security)
			}
			output, err := service.Start(ctx, tt.input)
			if (err != nil) != tt.wantError {
				t.Fatalf("Start() error = %v, wantError %t", err, tt.wantError)
			}
			if store.replaced != tt.wantStored {
				t.Fatalf("stored = %t, want %t", store.replaced, tt.wantStored)
			}
			if tt.wantStored {
				if output.TransactionReference == "" || output.AuthorizationURL == "" {
					t.Fatal("successful start returned empty output")
				}
				if len(store.previous) == 0 || !bytes.Equal(store.previous, security.Hash([]byte("old"))) {
					if tt.input.PreviousReference != "" {
						t.Fatal("previous transaction was not hashed")
					}
				}
				if !store.replaced || now.IsZero() {
					t.Fatal("transaction was not stored")
				}
			}
		})
	}
	short := &oauthTestSecurity{
		shortResult: 1,
	}
	store := &oauthTestStore{}
	provider := &oauthTestProvider{
		authorizationURL: "https://accounts.google.test/authorize",
	}
	service, _ := newOAuthService(store, provider, short)
	if _, err := service.Start(ctx, StartInput{
		Origins: []string{"https://admin.example.test"},
	}); err == nil {
		t.Fatal("short random result succeeded")
	}
}

func TestServiceCallback(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name      string
		input     CallbackInput
		configure func(*oauthTestStore, *oauthTestProvider, *oauthTestSecurity, time.Time)
		wantError bool
		wantCall  bool
	}{
		{
			name: "success",
			input: CallbackInput{
				TransactionReference: "tx",
				State:                "state",
				Code:                 "code",
			},
			configure: func(store *oauthTestStore, provider *oauthTestProvider, security *oauthTestSecurity, now time.Time) {
				store.consume, store.found = validOAuthRecord(now), true
				security.openValue = []byte("verifier")
				provider.identity = domainauth.VerifiedIdentity{
					Email:         "admin@example.test",
					EmailVerified: true,
				}
			},
			wantCall: true,
		},
		{
			name: "empty reference",
			input: CallbackInput{
				State: "state",
				Code:  "code",
			},
			wantError: true,
		},
		{
			name: "empty state",
			input: CallbackInput{
				TransactionReference: "tx",
				Code:                 "code",
			},
			wantError: true,
		},
		{
			name: "missing code and error",
			input: CallbackInput{
				TransactionReference: "tx",
				State:                "state",
			},
			wantError: true,
		},
		{
			name: "code and error together",
			input: CallbackInput{
				TransactionReference: "tx",
				State:                "state",
				Code:                 "code",
				ProviderError:        "access_denied",
			},
			wantError: true,
		},
		{
			name: "consume error",
			input: CallbackInput{
				TransactionReference: "tx",
				State:                "state",
				Code:                 "code",
			},
			configure: func(store *oauthTestStore, _ *oauthTestProvider, _ *oauthTestSecurity, _ time.Time) {
				store.consumeErr = errors.New("consume")
			},
			wantError: true,
		},
		{
			name: "transaction missing",
			input: CallbackInput{
				TransactionReference: "tx",
				State:                "state",
				Code:                 "code",
			},
			configure: func(store *oauthTestStore, _ *oauthTestProvider, _ *oauthTestSecurity, _ time.Time) {
				store.found = false
			},
			wantError: true,
		},
		{
			name: "provider cancellation",
			input: CallbackInput{
				TransactionReference: "tx",
				State:                "state",
				ProviderError:        "access_denied",
			},
			configure: func(store *oauthTestStore, _ *oauthTestProvider, _ *oauthTestSecurity, now time.Time) {
				store.consume, store.found = validOAuthRecord(now), true
			},
			wantError: true,
		},
		{
			name: "invalid transaction",
			input: CallbackInput{
				TransactionReference: "tx",
				State:                "state",
				Code:                 "code",
			},
			configure: func(store *oauthTestStore, _ *oauthTestProvider, _ *oauthTestSecurity, _ time.Time) {
				store.found = true
			},
			wantError: true,
		},
		{
			name: "open failure",
			input: CallbackInput{
				TransactionReference: "tx",
				State:                "state",
				Code:                 "code",
			},
			configure: func(store *oauthTestStore, _ *oauthTestProvider, security *oauthTestSecurity, now time.Time) {
				store.consume, store.found = validOAuthRecord(now), true
				security.openErr = errors.New("open")
			},
			wantError: true,
		},
		{
			name: "empty verifier",
			input: CallbackInput{
				TransactionReference: "tx",
				State:                "state",
				Code:                 "code",
			},
			configure: func(store *oauthTestStore, _ *oauthTestProvider, security *oauthTestSecurity, now time.Time) {
				store.consume, store.found = validOAuthRecord(now), true
				security.openValue = nil
			},
			wantError: true,
		},
		{
			name: "provider verification failure",
			input: CallbackInput{
				TransactionReference: "tx",
				State:                "state",
				Code:                 "code",
			},
			configure: func(store *oauthTestStore, provider *oauthTestProvider, security *oauthTestSecurity, now time.Time) {
				store.consume, store.found = validOAuthRecord(now), true
				security.openValue = []byte("verifier")
				provider.exchangeErr = errors.New("verify")
			},
			wantError: true,
			wantCall:  true,
		},
		{
			name: "unverified identity",
			input: CallbackInput{
				TransactionReference: "tx",
				State:                "state",
				Code:                 "code",
			},
			configure: func(store *oauthTestStore, provider *oauthTestProvider, security *oauthTestSecurity, now time.Time) {
				store.consume, store.found = validOAuthRecord(now), true
				security.openValue = []byte("verifier")
				provider.identity = domainauth.VerifiedIdentity{
					Email: "admin@example.test",
				}
			},
			wantError: true,
			wantCall:  true,
		},
		{
			name: "unauthorized identity",
			input: CallbackInput{
				TransactionReference: "tx",
				State:                "state",
				Code:                 "code",
			},
			configure: func(store *oauthTestStore, provider *oauthTestProvider, security *oauthTestSecurity, now time.Time) {
				store.consume, store.found = validOAuthRecord(now), true
				security.openValue = []byte("verifier")
				provider.identity = domainauth.VerifiedIdentity{
					Email:         "other@example.test",
					EmailVerified: true,
				}
			},
			wantError: true,
			wantCall:  true,
		},
		{
			name: "random failure",
			input: CallbackInput{
				TransactionReference: "tx",
				State:                "state",
				Code:                 "code",
			},
			configure: func(store *oauthTestStore, provider *oauthTestProvider, security *oauthTestSecurity, now time.Time) {
				store.consume, store.found = validOAuthRecord(now), true
				security.openValue = []byte("verifier")
				provider.identity = domainauth.VerifiedIdentity{
					Email:         "admin@example.test",
					EmailVerified: true,
				}
				security.randomFail = 2
			},
			wantError: true,
			wantCall:  true,
		},
		{
			name: "session ID random failure",
			input: CallbackInput{
				TransactionReference: "tx",
				State:                "state",
				Code:                 "code",
			},
			configure: func(store *oauthTestStore, provider *oauthTestProvider, security *oauthTestSecurity, now time.Time) {
				store.consume, store.found = validOAuthRecord(now), true
				security.openValue = []byte("verifier")
				provider.identity = domainauth.VerifiedIdentity{
					Email:         "admin@example.test",
					EmailVerified: true,
				}
				security.randomFail = 3
			},
			wantError: true,
			wantCall:  true,
		},
		{
			name: "session store failure",
			input: CallbackInput{
				TransactionReference: "tx",
				State:                "state",
				Code:                 "code",
			},
			configure: func(store *oauthTestStore, provider *oauthTestProvider, security *oauthTestSecurity, now time.Time) {
				store.consume, store.found = validOAuthRecord(now), true
				security.openValue = []byte("verifier")
				provider.identity = domainauth.VerifiedIdentity{
					Email:         "admin@example.test",
					EmailVerified: true,
				}
				store.createErr = errors.New("session")
			},
			wantError: true,
			wantCall:  true,
		},
		{
			name: "session csrf seal failure",
			input: CallbackInput{
				TransactionReference: "tx",
				State:                "state",
				Code:                 "code",
			},
			configure: func(store *oauthTestStore, provider *oauthTestProvider, security *oauthTestSecurity, now time.Time) {
				store.consume, store.found = validOAuthRecord(now), true
				security.openValue = []byte("verifier")
				provider.identity = domainauth.VerifiedIdentity{
					Email:         "admin@example.test",
					EmailVerified: true,
				}
				security.sealErr = errors.New("seal")
			},
			wantError: true,
			wantCall:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &oauthTestStore{}
			provider := &oauthTestProvider{}
			security := &oauthTestSecurity{}
			service, now := newOAuthService(store, provider, security)
			if tt.configure != nil {
				tt.configure(store, provider, security, now)
			}
			output, err := service.Callback(ctx, tt.input)
			if (err != nil) != tt.wantError {
				t.Fatalf("Callback() error = %v, wantError %t", err, tt.wantError)
			}
			if err == nil && store.created && output.SessionReference == "" {
				t.Fatal("session was stored without a reference")
			}
			if tt.wantCall && provider.gotCode != tt.input.Code {
				t.Errorf("provider code = %q, want %q", provider.gotCode, tt.input.Code)
			}
		})
	}
	short := &oauthTestSecurity{
		shortResult: 1,
	}
	store := &oauthTestStore{}
	provider := &oauthTestProvider{}
	service, now := newOAuthService(store, provider, short)
	store.consume, store.found = validOAuthRecord(now), true
	short.openValue = []byte("verifier")
	provider.identity = domainauth.VerifiedIdentity{
		Email:         "admin@example.test",
		EmailVerified: true,
	}
	if _, err := service.Callback(ctx, CallbackInput{
		TransactionReference: "tx",
		State:                "state",
		Code:                 "code",
	}); err == nil {
		t.Fatal("short random result succeeded")
	}
}

func TestNewServiceRejectsInvalidDependencies(t *testing.T) {
	valid := Dependencies{
		Store:    &oauthTestStore{},
		Provider: &oauthTestProvider{},
		Security: &oauthTestSecurity{},
		Clock:    oauthTestClock{},
		Logger:   oauthTestLogger{},
	}
	for _, test := range []struct {
		name         string
		dependencies Dependencies
		origin       string
		email        string
	}{
		{
			name: "missing store",
			dependencies: Dependencies{
				Provider: valid.Provider,
				Security: valid.Security,
				Clock:    valid.Clock,
				Logger:   valid.Logger,
			},
			origin: "https://admin.example.test",
			email:  "admin@example.test",
		},
		{
			name: "missing provider",
			dependencies: Dependencies{
				Store:    valid.Store,
				Security: valid.Security,
				Clock:    valid.Clock,
				Logger:   valid.Logger,
			},
			origin: "https://admin.example.test",
			email:  "admin@example.test",
		},
		{
			name: "missing security",
			dependencies: Dependencies{
				Store:    valid.Store,
				Provider: valid.Provider,
				Clock:    valid.Clock,
				Logger:   valid.Logger,
			},
			origin: "https://admin.example.test",
			email:  "admin@example.test",
		},
		{
			name: "missing clock",
			dependencies: Dependencies{
				Store:    valid.Store,
				Provider: valid.Provider,
				Security: valid.Security,
				Logger:   valid.Logger,
			},
			origin: "https://admin.example.test",
			email:  "admin@example.test",
		},
		{
			name: "missing logger",
			dependencies: Dependencies{
				Store:    valid.Store,
				Provider: valid.Provider,
				Security: valid.Security,
				Clock:    valid.Clock,
			},
			origin: "https://admin.example.test",
			email:  "admin@example.test",
		},
		{
			name:         "missing origin",
			dependencies: valid,
			email:        "admin@example.test",
		},
		{
			name:         "missing email",
			dependencies: valid,
			origin:       "https://admin.example.test",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewService(test.dependencies, test.origin, test.email)
			if apperr.CodeOf(err) != apperr.CodeInvalidConfiguration {
				t.Fatalf("NewService() code = %q, want %q", apperr.CodeOf(err), apperr.CodeInvalidConfiguration)
			}
		})
	}
}

func TestSystemClock(t *testing.T) {
	if (SystemClock{}).Now().IsZero() {
		t.Fatal("system clock returned zero")
	}
}
