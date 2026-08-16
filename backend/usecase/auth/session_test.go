package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yukihito-jokyu/topic2html/backend/apperr"
	domainauth "github.com/yukihito-jokyu/topic2html/backend/domain/auth"
)

type sessionTestStore struct {
	session       domainauth.AdminSession
	found         bool
	findErr       error
	revoked       bool
	revokeErr     error
	runAuthorized bool
	runErr        error
}

func (*sessionTestStore) ReplaceOAuthTransaction(context.Context, domainauth.Hash, domainauth.OAuthTransaction) error {
	return nil
}
func (*sessionTestStore) ConsumeOAuthTransaction(context.Context, domainauth.Hash, domainauth.Hash, domainauth.Time) (domainauth.OAuthTransaction, bool, error) {
	return domainauth.OAuthTransaction{}, false, nil
}
func (*sessionTestStore) CreateAdminSession(context.Context, domainauth.AdminSession) error {
	return nil
}
func (s *sessionTestStore) FindAdminSession(context.Context, domainauth.Hash) (domainauth.AdminSession, bool, error) {
	return s.session, s.found, s.findErr
}
func (s *sessionTestStore) RevokeAdminSession(context.Context, domainauth.Hash, domainauth.Time) (bool, error) {
	return s.revoked, s.revokeErr
}
func (s *sessionTestStore) RunAuthorizedAdminStateChange(ctx context.Context, _ domainauth.Hash, _ string, _ domainauth.Hash, _ domainauth.Time, operation AdminStateChangeOperation) (bool, error) {
	if s.runErr != nil {
		return false, s.runErr
	}
	if !s.runAuthorized {
		return false, nil
	}
	if err := operation(ctx); err != nil {
		return false, err
	}

	return true, nil
}
func (*sessionTestStore) DeleteExpiredProtectedRecords(context.Context, domainauth.Time) error {
	return nil
}

func validSessionForTest(now time.Time, security *oauthTestSecurity) domainauth.AdminSession {
	return domainauth.AdminSession{
		ID:                  "00000000-0000-4000-8000-000000000001",
		ReferenceHash:       security.Hash([]byte("session")),
		AuthorizedEmail:     "admin@example.test",
		CSRFTokenHash:       security.Hash([]byte("csrf")),
		CSRFTokenCiphertext: domainauth.Ciphertext("sealed"),
		CreatedAt:           now,
		LastMutationAt:      now,
		AbsoluteExpiresAt:   now.Add(domainauth.SessionAbsoluteLifetime),
		IdleExpiresAt:       now.Add(domainauth.SessionIdleLifetime),
	}
}

func newSessionService(store *sessionTestStore, security *oauthTestSecurity) (*Service, time.Time) {
	provider := &oauthTestProvider{}
	service, now := newOAuthService(&oauthTestStore{}, provider, security)
	service.store = store

	return service, now
}

func TestSessionBootstrap(t *testing.T) {
	for _, tt := range []struct {
		name      string
		configure func(*sessionTestStore, *oauthTestSecurity, time.Time)
		wantAuth  bool
		wantError bool
	}{
		{
			name: "authenticated",
			configure: func(store *sessionTestStore, security *oauthTestSecurity, now time.Time) {
				store.session, store.found = validSessionForTest(now, security), true
				security.openValue = []byte("csrf")
			},
			wantAuth: true,
		},
		{
			name: "anonymous",
		},
		{
			name: "record failure",
			configure: func(store *sessionTestStore, _ *oauthTestSecurity, _ time.Time) {
				store.findErr = errors.New("database")
			},
			wantError: true,
		},
		{
			name: "ciphertext failure",
			configure: func(store *sessionTestStore, security *oauthTestSecurity, now time.Time) {
				store.session, store.found = validSessionForTest(now, security), true
				security.openErr = errors.New("ciphertext")
			},
			wantError: true,
		},
		{
			name: "hash mismatch",
			configure: func(store *sessionTestStore, security *oauthTestSecurity, now time.Time) {
				store.session, store.found = validSessionForTest(now, security), true
				security.openValue = []byte("other")
			},
			wantError: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			store := &sessionTestStore{}
			security := &oauthTestSecurity{}
			service, now := newSessionService(store, security)
			if tt.configure != nil {
				tt.configure(store, security, now)
			}
			output, err := service.Bootstrap(context.Background(), "session")
			if (err != nil) != tt.wantError || output.Authenticated != tt.wantAuth {
				t.Fatalf("output=%+v err=%v", output, err)
			}
			if tt.wantError && apperr.CodeOf(err) != apperr.CodeUnavailable {
				t.Fatalf("error code=%v", apperr.CodeOf(err))
			}
			if tt.wantAuth && output.CSRFToken != "csrf" {
				t.Fatalf("CSRF token=%q", output.CSRFToken)
			}
		})
	}
}

func TestSessionGuardsAndLogout(t *testing.T) {
	now := time.Date(2026, 8, 15, 12, 0, 0, 0, time.UTC)
	for _, tt := range []struct {
		name      string
		input     SessionInput
		configure func(*sessionTestStore, *oauthTestSecurity)
		method    func(*Service, context.Context, SessionInput) (GuardDecision, error)
		want      GuardDecision
		wantError bool
	}{
		{
			name: "expired read is unauthenticated",
			input: SessionInput{
				SessionReference: "session",
			},
			configure: func(store *sessionTestStore, security *oauthTestSecurity) {
				store.session, store.found = validSessionForTest(now, security), true
				expired := now.Add(-time.Second)
				store.session.IdleExpiresAt = expired
			},
			method: func(service *Service, ctx context.Context, input SessionInput) (GuardDecision, error) {
				return service.AuthorizeRead(ctx, input.SessionReference)
			},
			want: GuardUnauthenticated,
		},
		{
			name: "logout anonymous is idempotent",
			input: SessionInput{
				Origins: []string{"https://admin.example.test"},
			},
			method: (*Service).Logout,
			want:   GuardAllowed,
		},
		{
			name: "logout revokes authenticated session",
			input: SessionInput{
				SessionReference: "session",
				Origins:          []string{"https://admin.example.test"},
				CSRFToken:        "csrf",
			},
			configure: func(store *sessionTestStore, security *oauthTestSecurity) {
				store.session, store.found, store.revoked = validSessionForTest(now, security), true, true
			},
			method: (*Service).Logout,
			want:   GuardAllowed,
		},
		{
			name: "logout record failure",
			input: SessionInput{
				SessionReference: "session",
				Origins:          []string{"https://admin.example.test"},
			},
			configure: func(store *sessionTestStore, _ *oauthTestSecurity) {
				store.findErr = errors.New("database")
			},
			method:    (*Service).Logout,
			want:      GuardUnauthenticated,
			wantError: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			store := &sessionTestStore{}
			security := &oauthTestSecurity{}
			service, _ := newSessionService(store, security)
			if tt.configure != nil {
				tt.configure(store, security)
			}
			got, err := tt.method(service, context.Background(), tt.input)
			if got != tt.want || (err != nil) != tt.wantError {
				t.Fatalf("decision=%v err=%v", got, err)
			}
		})
	}
}

func TestRunAuthorizedAdminStateChange(t *testing.T) {
	for _, tt := range []struct {
		name      string
		configure func(*sessionTestStore)
		operation AdminStateChangeOperation
		want      GuardDecision
		wantError bool
	}{
		{
			name: "success",
			configure: func(store *sessionTestStore) {
				store.runAuthorized = true
			},
			operation: func(context.Context) error { return nil },
			want:      GuardAllowed,
		},
		{
			name:      "authorization changed",
			operation: func(context.Context) error { return nil },
			want:      GuardUnauthenticated,
		},
		{
			name:      "origin rejected",
			operation: func(context.Context) error { return nil },
			want:      GuardForbidden,
		},
		{
			name:      "csrf changed",
			operation: func(context.Context) error { return nil },
			want:      GuardForbidden,
		},
		{
			name: "store failure",
			configure: func(store *sessionTestStore) {
				store.runErr = errors.New("database")
			},
			operation: func(context.Context) error { return nil },
			want:      GuardUnauthenticated,
			wantError: true,
		},
		{
			name: "session lookup failure",
			configure: func(store *sessionTestStore) {
				store.findErr = errors.New("database")
			},
			operation: func(context.Context) error { return nil },
			want:      GuardUnauthenticated,
			wantError: true,
		},
		{
			name: "business operation failure",
			configure: func(store *sessionTestStore) {
				store.runAuthorized = true
			},
			operation: func(context.Context) error { return errors.New("business") },
			want:      GuardUnauthenticated,
			wantError: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			store := &sessionTestStore{}
			security := &oauthTestSecurity{}
			service, now := newSessionService(store, security)
			store.session, store.found = validSessionForTest(now, security), true
			if tt.configure != nil {
				tt.configure(store)
			}
			csrfToken := "csrf"
			origin := "https://admin.example.test"
			if tt.name == "csrf changed" {
				csrfToken = "wrong"
			}
			if tt.name == "origin rejected" {
				origin = "https://evil.example.test"
			}
			decision, err := service.RunAuthorizedAdminStateChange(context.Background(), SessionInput{
				SessionReference: "session",
				Origins:          []string{origin},
				CSRFToken:        csrfToken,
			}, tt.operation)
			if decision != tt.want || (err != nil) != tt.wantError {
				t.Fatalf("decision=%v err=%v", decision, err)
			}
		})
	}
}

func TestAuthorizeReadUnavailable(t *testing.T) {
	store := &sessionTestStore{
		findErr: errors.New("database"),
	}
	service, _ := newSessionService(store, &oauthTestSecurity{})
	decision, err := service.AuthorizeRead(context.Background(), "session")
	if decision != GuardUnauthenticated || apperr.CodeOf(err) != apperr.CodeUnavailable {
		t.Fatalf("decision=%v err=%v", decision, err)
	}
}

func TestLogoutRejectedAndRevokeFailure(t *testing.T) {
	for _, tt := range []struct {
		name      string
		input     SessionInput
		configure func(*sessionTestStore, *oauthTestSecurity, time.Time)
		want      GuardDecision
		wantError bool
	}{
		{
			name: "invalid origin",
			input: SessionInput{
				Origins: []string{"https://evil.example.test"},
			},
			want: GuardForbidden,
		},
		{
			name: "invalid csrf",
			input: SessionInput{
				SessionReference: "session",
				Origins:          []string{"https://admin.example.test"},
				CSRFToken:        "wrong",
			},
			configure: func(store *sessionTestStore, security *oauthTestSecurity, now time.Time) {
				store.session, store.found = validSessionForTest(now, security), true
			},
			want: GuardForbidden,
		},
		{
			name: "revoke failure",
			input: SessionInput{
				SessionReference: "session",
				Origins:          []string{"https://admin.example.test"},
				CSRFToken:        "csrf",
			},
			configure: func(store *sessionTestStore, security *oauthTestSecurity, now time.Time) {
				store.session, store.found, store.revokeErr = validSessionForTest(now, security), true, errors.New("database")
			},
			want:      GuardUnauthenticated,
			wantError: true,
		},
		{
			name: "already revoked concurrently",
			input: SessionInput{
				SessionReference: "session",
				Origins:          []string{"https://admin.example.test"},
				CSRFToken:        "csrf",
			},
			configure: func(store *sessionTestStore, security *oauthTestSecurity, now time.Time) {
				store.session, store.found = validSessionForTest(now, security), true
			},
			want: GuardAllowed,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			store := &sessionTestStore{}
			security := &oauthTestSecurity{}
			service, now := newSessionService(store, security)
			if tt.configure != nil {
				tt.configure(store, security, now)
			}
			decision, err := service.Logout(context.Background(), tt.input)
			if decision != tt.want || (err != nil) != tt.wantError {
				t.Fatalf("decision=%v err=%v", decision, err)
			}
		})
	}
}
