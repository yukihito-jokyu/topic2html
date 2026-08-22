package integration_test

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yukihito-jokyu/topic2html/backend/domain/auth"
	ginadapter "github.com/yukihito-jokyu/topic2html/backend/handler/gin"
	"github.com/yukihito-jokyu/topic2html/backend/observability"
	"github.com/yukihito-jokyu/topic2html/backend/repository/postgres"
	"github.com/yukihito-jokyu/topic2html/backend/repository/security"
	usecaseauth "github.com/yukihito-jokyu/topic2html/backend/usecase/auth"
)

var integrationDatabaseURL = flag.String("integration-database-url", "", "PostgreSQL URL for handler integration tests")

const adminSessionCookieName = "__Host-topic2html_admin_session"

// TestAdminSessionHTTPPostgresIntegrationはHTTP、usecase、実PostgreSQLを通じてsession操作を検証します。
func TestAdminSessionHTTPPostgresIntegration(t *testing.T) {
	if *integrationDatabaseURL == "" {
		t.Skip("integration-database-url is not configured")
	}
	ctx := context.Background()
	pool := newSessionIntegrationPool(t, ctx)
	if err := postgres.ApplyMigrations(ctx, pool); err != nil {
		t.Fatal(err)
	}
	protection, err := security.New("handler-integration-protection-key")
	if err != nil {
		t.Fatal(err)
	}
	store := postgres.NewStore(pool)
	service, err := usecaseauth.NewService(usecaseauth.Dependencies{
		Store:    store,
		Provider: integrationProvider{},
		Security: protection,
		Clock:    usecaseauth.SystemClock{},
		Logger:   observability.NewDiscardLogger(),
	}, "https://admin.example.test", "admin@example.test")
	if err != nil {
		t.Fatal(err)
	}
	reference := "session-reference"
	csrfToken := "csrf-token"
	ciphertext, err := protection.Seal([]byte(csrfToken))
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	if err := store.CreateAdminSession(ctx, auth.AdminSession{
		ID:                  "00000000-0000-0000-0000-000000000001",
		ReferenceHash:       protection.Hash([]byte(reference)),
		AuthorizedEmail:     "admin@example.test",
		CSRFTokenHash:       protection.Hash([]byte(csrfToken)),
		CSRFTokenCiphertext: ciphertext,
		CreatedAt:           now,
		LastMutationAt:      now,
		AbsoluteExpiresAt:   now.Add(auth.SessionAbsoluteLifetime),
		IdleExpiresAt:       now.Add(auth.SessionIdleLifetime),
	}); err != nil {
		t.Fatal(err)
	}
	router := ginadapter.NewRouter(service, service, observability.NewDiscardLogger())
	bootstrapCases := []struct {
		name          string
		withSession   bool
		authenticated bool
	}{
		{
			name:          "returns an anonymous session without a cookie",
			authenticated: false,
		},
		{
			name:          "returns the authenticated session with a valid cookie",
			withSession:   true,
			authenticated: true,
		},
	}
	for _, testCase := range bootstrapCases {
		t.Run(testCase.name, func(t *testing.T) {
			request := httptest.NewRequestWithContext(ctx, http.MethodGet, "/admin/auth/session", nil)
			if testCase.withSession {
				request.AddCookie(&http.Cookie{
					Name:     adminSessionCookieName,
					Value:    reference,
					Secure:   true,
					HttpOnly: true,
					SameSite: http.SameSiteLaxMode,
				})
			}
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != http.StatusOK {
				t.Fatalf("status=%d", response.Code)
			}
			var body struct {
				Authenticated bool `json:"authenticated"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body.Authenticated != testCase.authenticated {
				t.Fatalf("authenticated=%t want=%t", body.Authenticated, testCase.authenticated)
			}
		})
	}

	bootstrap := httptest.NewRequestWithContext(ctx, http.MethodGet, "/admin/auth/session", nil)
	bootstrap.AddCookie(&http.Cookie{
		Name:     adminSessionCookieName,
		Value:    reference,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	bootstrapResponse := httptest.NewRecorder()
	router.ServeHTTP(bootstrapResponse, bootstrap)
	if bootstrapResponse.Code != http.StatusOK || bootstrapResponse.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("bootstrap status/cache = %d/%q", bootstrapResponse.Code, bootstrapResponse.Header().Get("Cache-Control"))
	}
	var bootstrapBody struct {
		Authenticated bool   `json:"authenticated"`
		CSRFToken     string `json:"csrf_token"`
	}
	if err := json.Unmarshal(bootstrapResponse.Body.Bytes(), &bootstrapBody); err != nil {
		t.Fatal(err)
	}
	if !bootstrapBody.Authenticated || bootstrapBody.CSRFToken != csrfToken {
		t.Fatalf("bootstrap body = %#v", bootstrapBody)
	}

	logout := httptest.NewRequestWithContext(ctx, http.MethodPost, "/admin/auth/logout", nil)
	logout.Header.Set("Origin", "https://admin.example.test")
	logout.Header.Set("X-CSRF-Token", csrfToken)
	logout.AddCookie(&http.Cookie{
		Name:     adminSessionCookieName,
		Value:    reference,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	logoutResponse := httptest.NewRecorder()
	router.ServeHTTP(logoutResponse, logout)
	if logoutResponse.Code != http.StatusOK || logoutResponse.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("logout status/cache = %d/%q", logoutResponse.Code, logoutResponse.Header().Get("Cache-Control"))
	}
	assertDeletedSessionCookie(t, logoutResponse)
	session, found, err := store.FindAdminSession(ctx, protection.Hash([]byte(reference)))
	if err != nil || !found || session.RevokedAt == nil {
		t.Fatalf("session revocation: found=%t revoked=%v err=%v", found, session.RevokedAt, err)
	}

	unauthenticated := httptest.NewRequestWithContext(ctx, http.MethodGet, "/admin/auth/session", nil)
	unauthenticated.AddCookie(&http.Cookie{
		Name:     adminSessionCookieName,
		Value:    reference,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	unauthenticatedResponse := httptest.NewRecorder()
	router.ServeHTTP(unauthenticatedResponse, unauthenticated)
	if unauthenticatedResponse.Code != http.StatusOK || unauthenticatedResponse.Body.String() != "{\"authenticated\":false}" {
		t.Fatalf("unauthenticated bootstrap = %d/%q", unauthenticatedResponse.Code, unauthenticatedResponse.Body.String())
	}
	assertDeletedSessionCookie(t, unauthenticatedResponse)
}

func newSessionIntegrationPool(t *testing.T, ctx context.Context) *pgxpool.Pool {
	t.Helper()
	bootstrap, err := pgxpool.New(ctx, *integrationDatabaseURL)
	if err != nil {
		t.Fatal(err)
	}
	schema := fmt.Sprintf("topic2html_handler_integration_%d", time.Now().UnixNano())
	if _, err := bootstrap.Exec(ctx, `CREATE SCHEMA `+schema); err != nil {
		bootstrap.Close()
		t.Fatal(err)
	}
	bootstrap.Close()
	config, err := pgxpool.ParseConfig(*integrationDatabaseURL)
	if err != nil {
		t.Fatal(err)
	}
	config.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		pool.Close()
		cleanup, err := pgxpool.New(ctx, *integrationDatabaseURL)
		if err == nil {
			_, _ = cleanup.Exec(ctx, `DROP SCHEMA `+schema+` CASCADE`)
			cleanup.Close()
		}
	})

	return pool
}

func assertDeletedSessionCookie(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()
	for _, cookie := range response.Result().Cookies() {
		if cookie.Name == adminSessionCookieName {
			if cookie.MaxAge >= 0 || !cookie.Secure || !cookie.HttpOnly || cookie.Path != "/" || cookie.SameSite != http.SameSiteLaxMode {
				t.Fatalf("deleted session cookie = %#v", cookie)
			}

			return
		}
	}
	t.Fatal("deleted session cookie is missing")
}

type integrationProvider struct{}

func (integrationProvider) AuthorizationURL(context.Context, usecaseauth.AuthorizationRequest) (string, error) {
	return "https://accounts.example.test/authorize", nil
}

func (integrationProvider) ExchangeAndVerify(context.Context, string, string, auth.Hash) (auth.VerifiedIdentity, error) {
	return auth.VerifiedIdentity{}, nil
}
